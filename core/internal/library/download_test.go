package library

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"light-launcher/core/internal/contract"
)

// startCall is one StartDownload call, with its params decoded from the JSON the plugin receives.
type startCall struct {
	plugin string
	gameID int64
	params any
}

// prepareSources records StartDownload calls. Sources and Call come from the shared fake.
type prepareSources struct {
	fakeSources
	startErr error
	starts   []startCall
}

func (p *prepareSources) StartDownload(_ context.Context, plugin string, params any, gameID int64) (int64, error) {
	raw, err := json.Marshal(params)
	if err != nil {
		return 0, err
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return 0, err
	}
	p.starts = append(p.starts, startCall{plugin: plugin, gameID: gameID, params: decoded})
	if p.startErr != nil {
		return 0, p.startErr
	}
	return 77, nil
}

// recordingDownloads records the requests queued without a plugin.
type recordingDownloads struct {
	requests []contract.DownloadRequest
}

func (r *recordingDownloads) Enqueue(_ context.Context, req contract.DownloadRequest) (int64, error) {
	r.requests = append(r.requests, req)
	return 55, nil
}

// f95Info is the source list entry of a plugin that can resolve files.
func f95Info() []contract.SourceInfo {
	return []contract.SourceInfo{{Plugin: "f95plugin", SourceID: "f95", Label: "F95", Capabilities: []string{CapabilityResolveFiles}}}
}

// savePluginGame stores a game of the f95 source with two saved files. It returns the game id and the file ids.
func savePluginGame(t *testing.T, lib *Library, sourceID, title string) (int64, []int64) {
	t.Helper()
	ctx := context.Background()
	gameID, err := lib.Upsert(ctx, contract.GameDTO{
		Source: "f95", SourceID: sourceID, URL: "https://f95.example/threads/" + sourceID, Title: title, Version: "1.0",
	})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	files := []contract.FileCandidate{
		{
			ID: "pixel-1", Label: "Windows build", URL: "https://files.example/win.zip", Kind: "direct",
			Platform: contract.PlatformWindows, Headers: map[string]string{"X-Token": "t"},
		},
		{ID: "mega-1", Label: "Mirror", URL: "https://mega.example/x", Kind: "mirror", Host: "mega", RequiresAuth: true},
	}
	if err := lib.ReplaceFiles(ctx, gameID, files); err != nil {
		t.Fatalf("save files: %v", err)
	}
	stored, err := lib.ListFiles(ctx, gameID)
	if err != nil {
		t.Fatalf("list files: %v", err)
	}
	ids := make([]int64, 0, len(stored))
	for _, file := range stored {
		ids = append(ids, file.ID)
	}
	return gameID, ids
}

func TestDownloadFilePreparesThroughPlugin(t *testing.T) {
	sources := &prepareSources{fakeSources: fakeSources{infos: f95Info()}}
	downloads := &recordingDownloads{}
	lib := openTestLibrary(t, Options{Sources: sources, Downloads: downloads})
	gameID, fileIDs := savePluginGame(t, lib, "9", "Plugin Game")

	id, err := lib.DownloadFile(context.Background(), gameID, fileIDs[0])
	if err != nil {
		t.Fatalf("DownloadFile: %v", err)
	}
	if id != 77 {
		t.Fatalf("download id = %d, want 77", id)
	}
	if len(downloads.requests) != 0 {
		t.Fatalf("a plugin game was queued directly: %+v", downloads.requests)
	}
	if len(sources.starts) != 1 {
		t.Fatalf("StartDownload calls = %d, want 1", len(sources.starts))
	}
	call := sources.starts[0]
	if call.plugin != "f95plugin" || call.gameID != gameID {
		t.Fatalf("StartDownload(plugin %q, game %d), want f95plugin, %d", call.plugin, call.gameID, gameID)
	}
	want := `{
		"source": "f95",
		"file": {"id": "pixel-1", "label": "Windows build", "url": "https://files.example/win.zip", "kind": "direct",
			"platform": "windows", "headers": {"X-Token": "t"}},
		"game": {"id": ` + strconv.FormatInt(gameID, 10) + `, "source": "f95", "sourceId": "9", "url": "https://f95.example/threads/9",
			"title": "Plugin Game", "version": "1.0", "updatedAt": null, "extra": null}
	}`
	assertJSON(t, call.params, want)
}

// assertJSON checks that got, a decoded JSON value, has the same content as the JSON text want.
func assertJSON(t *testing.T, got any, want string) {
	t.Helper()
	var expected any
	if err := json.Unmarshal([]byte(want), &expected); err != nil {
		t.Fatalf("bad expected JSON: %v", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("params = %v, want %s", got, want)
	}
}

func TestDownloadFileQueuesLinkWithoutPlugin(t *testing.T) {
	cases := []struct {
		name  string
		infos []contract.SourceInfo
	}{
		{name: "source without resolveFiles", infos: []contract.SourceInfo{{Plugin: "f95plugin", SourceID: "f95", Capabilities: []string{CapabilityBrowse}}}},
		{name: "no plugin for the source", infos: nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sources := &prepareSources{fakeSources: fakeSources{infos: c.infos}}
			downloads := &recordingDownloads{}
			lib := openTestLibrary(t, Options{Sources: sources, Downloads: downloads})
			gameID, fileIDs := savePluginGame(t, lib, "8", "Linked Game")

			id, err := lib.DownloadFile(context.Background(), gameID, fileIDs[0])
			if err != nil {
				t.Fatalf("DownloadFile: %v", err)
			}
			if id != 55 {
				t.Fatalf("download id = %d, want 55", id)
			}
			if len(sources.starts) != 0 {
				t.Fatalf("StartDownload was called for a game without a prepare step")
			}
			want := contract.DownloadRequest{
				GameID: gameID, URL: "https://files.example/win.zip", Headers: map[string]string{"X-Token": "t"}, Install: true,
			}
			if !reflect.DeepEqual(downloads.requests, []contract.DownloadRequest{want}) {
				t.Fatalf("queued = %+v, want %+v", downloads.requests, want)
			}
		})
	}
}

func TestDownloadFileRejectsBadRequests(t *testing.T) {
	// fileID picks the file id to request. own are the files of the game; other is a file of a second game.
	cases := []struct {
		name     string
		fileID   func(own []int64, other int64) int64
		startErr error
		want     error
		message  string
	}{
		{
			name:   "unknown file id",
			fileID: func([]int64, int64) int64 { return 999 },
			want:   ErrNotFound,
		},
		{
			name:   "file of another game",
			fileID: func(_ []int64, other int64) int64 { return other },
			want:   ErrNotFound,
		},
		{
			name:     "plugin refuses",
			fileID:   func(own []int64, _ int64) int64 { return own[0] },
			startErr: errors.New("prepare download with f95plugin: the window stopped without a result"),
			message:  "the window stopped without a result",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sources := &prepareSources{fakeSources: fakeSources{infos: f95Info()}, startErr: c.startErr}
			downloads := &recordingDownloads{}
			lib := openTestLibrary(t, Options{Sources: sources, Downloads: downloads})
			gameID, fileIDs := savePluginGame(t, lib, "9", "Plugin Game")
			_, otherFiles := savePluginGame(t, lib, "10", "Other Game")

			_, err := lib.DownloadFile(context.Background(), gameID, c.fileID(fileIDs, otherFiles[0]))
			if err == nil {
				t.Fatal("DownloadFile succeeded, want an error")
			}
			if c.want != nil && !errors.Is(err, c.want) {
				t.Fatalf("err = %v, want %v", err, c.want)
			}
			if c.message != "" && !strings.Contains(err.Error(), c.message) {
				t.Fatalf("err = %q, want it to contain %q", err, c.message)
			}
			if len(downloads.requests) != 0 {
				t.Fatalf("a rejected request was queued: %+v", downloads.requests)
			}
		})
	}
}

func TestDownloadFileUnknownGame(t *testing.T) {
	lib := openTestLibrary(t, Options{Sources: &prepareSources{fakeSources: fakeSources{infos: f95Info()}}, Downloads: &recordingDownloads{}})
	if _, err := lib.DownloadFile(context.Background(), 999, 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestDownloadFileWithoutDownloadQueue(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	gameID, fileIDs := savePluginGame(t, lib, "7", "Linked Game")
	if _, err := lib.DownloadFile(context.Background(), gameID, fileIDs[0]); !errors.Is(err, ErrNoDownloads) {
		t.Fatalf("err = %v, want ErrNoDownloads", err)
	}
}
