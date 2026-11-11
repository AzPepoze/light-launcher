package librarysvc

import (
	"context"
	"encoding/json"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"testing"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/library"
)

// fakeSources answers the search method with one game, and nothing else.
type fakeSources struct{}

func (fakeSources) Sources() []contract.SourceInfo {
	return []contract.SourceInfo{{Plugin: "f95plugin", SourceID: "f95", Label: "F95", Capabilities: []string{library.CapabilityBrowse}}}
}

func (fakeSources) Call(_ context.Context, _, method string, _ any, out any) error {
	if method != "source.search" {
		return nil
	}
	raw, _ := json.Marshal(map[string][]contract.GameDTO{"games": {{Source: "f95", SourceID: "9", URL: "https://f95/9", Title: "Found Game"}}})
	return json.Unmarshal(raw, out)
}

// recordingEvents keeps the events the service emits.
type recordingEvents struct {
	mu     sync.Mutex
	events []string
}

func (r *recordingEvents) Emit(channel string, _ any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, channel)
}

func setup(t *testing.T) (*api.Router, *recordingEvents) {
	t.Helper()
	lib, err := library.Open(context.Background(), library.Options{
		DBPath:   filepath.Join(t.TempDir(), "library.db"),
		Sources:  fakeSources{},
		MediaDir: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { lib.Close() })
	router := api.NewRouter()
	events := &recordingEvents{}
	Register(router, events, lib, fakeSources{})
	return router, events
}

func call(t *testing.T, router *api.Router, method string, payload string) any {
	t.Helper()
	result, err := router.Dispatch(context.Background(), method, json.RawMessage(payload))
	if err != nil {
		t.Fatalf("%s: %v", method, err)
	}
	return result
}

func TestRegisterAddsEveryLibraryMethod(t *testing.T) {
	router, _ := setup(t)
	want := []string{
		"LibraryList", "LibraryGet", "LibraryUpdate", "LibraryRemove", "LibrarySearch", "LibraryPreview",
		"LibraryAddFromUrl", "LibraryLinkLocal", "LibraryCheckUpdates", "LibrarySetFavorite",
		"LibraryImportLegacy", "LibraryScanFolder", "LibraryLaunch", "LibraryLinkFolder", "LibraryUpdatesSummary",
		"IncomingList", "IncomingLink", "IncomingIgnore", "IncomingCreateLocal",
	}
	got := router.Methods()
	sort.Strings(got)
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("methods = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("method %d = %s, want %s", i, got[i], want[i])
		}
	}
}

func TestLibraryUpdateChangesNotesAndEmitsEvent(t *testing.T) {
	router, events := setup(t)
	id := createManual(t, router)

	game := call(t, router, "LibraryUpdate", idJSON(id, `"notes":"loved it"`)).(library.Game)
	if game.Notes == nil || *game.Notes != "loved it" {
		t.Errorf("notes = %v", game.Notes)
	}
	if len(events.events) == 0 || events.events[len(events.events)-1] != "library.game" {
		t.Errorf("events = %v, want library.game", events.events)
	}
}

func TestLibraryUpdateRejectsEmptyTitle(t *testing.T) {
	router, _ := setup(t)
	id := createManual(t, router)
	if _, err := router.Dispatch(context.Background(), "LibraryUpdate", json.RawMessage(idJSON(id, `"title":"  "`))); err == nil {
		t.Error("an empty title must be rejected")
	}
}

func TestLibrarySearchAggregatesBrowsableSources(t *testing.T) {
	router, _ := setup(t)
	results := call(t, router, "LibrarySearch", `{"query":"found"}`).([]library.SearchResult)
	if len(results) != 1 || len(results[0].Games) != 1 || results[0].Games[0].Title != "Found Game" {
		t.Errorf("results = %+v", results)
	}
}

func TestLibraryListAndRemoveKeepHistory(t *testing.T) {
	router, _ := setup(t)
	id := createManual(t, router)

	page := call(t, router, "LibraryList", `{"search":"manual"}`).(library.Listing)
	if page.Total != 1 {
		t.Errorf("list total = %d, want 1", page.Total)
	}
	call(t, router, "LibraryRemove", idJSON(id, ""))
	page = call(t, router, "LibraryList", `{}`).(library.Listing)
	if page.Total != 0 {
		t.Errorf("list total after remove = %d, want 0", page.Total)
	}
}

func TestIncomingListWithoutWatchFolderIsNotConfigured(t *testing.T) {
	router, _ := setup(t)
	report := call(t, router, "IncomingList", ``).(library.IncomingReport)
	if report.Configured {
		t.Error("no watch folder means not configured")
	}
}

func TestBadPayloadIsAnError(t *testing.T) {
	router, _ := setup(t)
	if _, err := router.Dispatch(context.Background(), "LibraryGet", json.RawMessage(`{"id":"nope"}`)); err == nil {
		t.Error("a string id must be rejected")
	}
}

// idJSON builds a payload with an id and optional extra members, such as `"notes":"x"`.
func idJSON(id int64, extra string) string {
	if extra == "" {
		return `{"id":` + strconv.FormatInt(id, 10) + `}`
	}
	return `{"id":` + strconv.FormatInt(id, 10) + `,` + extra + `}`
}

// createManual saves a manual game through the UI method and returns its id.
func createManual(t *testing.T, router *api.Router) int64 {
	t.Helper()
	game := call(t, router, "IncomingCreateLocal", `{"name":"Manual_Game_v1.zip"}`).(library.Game)
	return game.ID
}
