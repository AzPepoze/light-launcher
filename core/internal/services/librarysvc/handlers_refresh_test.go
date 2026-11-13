package librarysvc

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/library"
)

// refreshSources serves source.fetchGame and source.resolveFiles, and records the file the
// download path hands to source.prepareDownload.
type refreshSources struct {
	fakeSources
	files    []contract.FileCandidate
	prepared contract.FileCandidate
}

func (r *refreshSources) Sources() []contract.SourceInfo {
	return []contract.SourceInfo{{Plugin: "f95plugin", SourceID: "f95", Label: "F95",
		Capabilities: []string{library.CapabilityResolveFiles}}}
}

func (r *refreshSources) Call(_ context.Context, _, method string, _ any, out any) error {
	switch method {
	case "source.fetchGame":
		return decodeReply(map[string]any{"game": contract.GameDTO{
			Source: "f95", SourceID: "9", URL: "https://f95/9", Title: "Refreshed", Version: "2.0",
		}}, out)
	case "source.resolveFiles":
		return decodeReply(map[string]any{"files": r.files}, out)
	}
	return errors.New("unexpected method " + method)
}

func (r *refreshSources) StartDownload(_ context.Context, _ string, params any, _ int64) (int64, error) {
	raw, err := json.Marshal(params)
	if err != nil {
		return 0, err
	}
	var decoded struct {
		File contract.FileCandidate `json:"file"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return 0, err
	}
	r.prepared = decoded.File
	return 42, nil
}

// noFetchSources has a plugin that does not answer source.fetchGame.
type noFetchSources struct {
	refreshSources
}

func (noFetchSources) Call(context.Context, string, string, any, any) error {
	return errors.New("method not found: source.fetchGame")
}

func decodeReply(value any, out any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, out)
}

func newRefreshRouter(t *testing.T, sources contract.Sources) (*api.Router, *library.Library, *recordingEvents) {
	t.Helper()
	lib, err := library.Open(context.Background(), library.Options{
		DBPath:   filepath.Join(t.TempDir(), "library.db"),
		Sources:  sources,
		MediaDir: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { lib.Close() })
	router := api.NewRouter()
	events := &recordingEvents{}
	Register(router, events, lib, sources)
	return router, lib, events
}

func createPluginGame(t *testing.T, lib *library.Library) int64 {
	t.Helper()
	id, err := lib.Upsert(context.Background(), contract.GameDTO{
		Source: "f95", SourceID: "9", URL: "https://f95/9", Title: "Stored",
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestLibraryRefreshFilesReturnsUpdatedDetailAndEmitsEvent(t *testing.T) {
	sources := &refreshSources{files: []contract.FileCandidate{{
		ID: "plugin-file-1", Label: "Windows", URL: "https://f95/9.zip", Kind: "direct",
	}}}
	router, lib, events := newRefreshRouter(t, sources)
	ctx := context.Background()
	id := createPluginGame(t, lib)
	if err := lib.ReplaceFiles(ctx, id, []contract.FileCandidate{{
		Label: "Old", URL: "https://f95/old.zip", Kind: "direct",
	}}); err != nil {
		t.Fatal(err)
	}
	favorite := true
	playtime := int64(60)
	if _, err := lib.Patch(ctx, id, library.Patch{Favorite: &favorite, PlaytimeS: &playtime}); err != nil {
		t.Fatal(err)
	}

	detail := call(t, router, "LibraryRefreshFiles", idJSON(id, "")).(library.GameDetail)
	if detail.Title != "Refreshed" || detail.Version == nil || *detail.Version != "2.0" {
		t.Errorf("game = %q %v, want the refreshed metadata", detail.Title, detail.Version)
	}
	if len(detail.Files) != 1 || detail.Files[0].Label != "Windows" {
		t.Fatalf("files = %+v, want the refreshed file", detail.Files)
	}
	if detail.Favorite != 1 || detail.PlaytimeS != 60 {
		t.Errorf("favorite = %d, playtime = %d, want 1 and 60", detail.Favorite, detail.PlaytimeS)
	}
	if len(events.events) == 0 || events.events[len(events.events)-1] != "library.game" {
		t.Errorf("events = %v, want library.game", events.events)
	}

	// The refreshed plugin file id is the one prepareDownload receives.
	queued := call(t, router, "LibraryDownloadFile",
		`{"gameId":`+strconv.FormatInt(id, 10)+`,"fileId":`+strconv.FormatInt(detail.Files[0].ID, 10)+`}`).(downloadFileResult)
	if queued.ID != 42 {
		t.Errorf("download id = %d, want 42", queued.ID)
	}
	if sources.prepared.ID != "plugin-file-1" {
		t.Errorf("prepareDownload file id = %q, want plugin-file-1", sources.prepared.ID)
	}
}

func TestLibraryRefreshFilesWithoutFetchGameReturnsError(t *testing.T) {
	router, lib, events := newRefreshRouter(t, &noFetchSources{})
	id := createPluginGame(t, lib)

	_, err := router.Dispatch(context.Background(), "LibraryRefreshFiles", json.RawMessage(idJSON(id, "")))
	if !errors.Is(err, library.ErrFetchGame) {
		t.Fatalf("err = %v, want ErrFetchGame", err)
	}
	if !strings.Contains(err.Error(), "fetch game") {
		t.Errorf("err = %v, want it to say the game fetch failed", err)
	}
	if len(events.events) != 0 {
		t.Errorf("events = %v, want none after a failed refresh", events.events)
	}
}
