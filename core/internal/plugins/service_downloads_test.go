package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"light-launcher/core/internal/contract"
)

// fakeDownloads records the last request and returns a fixed error when one is set.
type fakeDownloads struct {
	last contract.DownloadRequest
	err  error
}

func (f *fakeDownloads) Enqueue(_ context.Context, req contract.DownloadRequest) (int64, error) {
	f.last = req
	return 42, f.err
}

// fakeLibrary returns ids 7, 8, ... for each upserted game.
type fakeLibrary struct{ next int64 }

func (f *fakeLibrary) Upsert(_ context.Context, _ contract.GameDTO) (int64, error) {
	f.next++
	return 6 + f.next, nil
}

func TestLibraryUpsertReturnsIdsForDownloads(t *testing.T) {
	env := testEnv(t, Permissions{}, nil)
	env.deps.library = &fakeLibrary{}
	result, err := callService(t, env, "library.upsert", map[string]any{
		"games": []contract.GameDTO{
			{Source: "demo", SourceID: "1", URL: "https://x.example/1", Title: "A"},
			{Source: "demo", SourceID: "2", URL: "https://x.example/2", Title: "B"},
		},
	})
	if err != nil {
		t.Fatalf("library.upsert: %v", err)
	}
	got := wireJSON(t, result)
	if want := `{"ids":[7,8],"upserted":2}`; got != want {
		t.Fatalf("result = %s, want %s", got, want)
	}
}

// downloadEnv returns a plugin env that may fetch from files.example and, when games is set, use $GAMES.
func downloadEnv(t *testing.T, fs []string, games string) (*pluginEnv, *fakeDownloads) {
	t.Helper()
	folders := map[string]string{}
	if games != "" {
		folders["$GAMES"] = games
	}
	env := testEnv(t, Permissions{Network: []string{"files.example"}, FS: fs}, folders)
	fake := &fakeDownloads{}
	env.deps.downloads = fake
	return env, fake
}

func enqueueParams(extra map[string]any) map[string]any {
	params := map[string]any{"url": "https://files.example/game.zip", "filename": "game.zip"}
	for key, value := range extra {
		params[key] = value
	}
	return params
}

func TestDownloadsOldParamsStillWork(t *testing.T) {
	env, fake := downloadEnv(t, nil, "")
	result, err := callService(t, env, "downloads.enqueue", enqueueParams(nil))
	if err != nil {
		t.Fatalf("plain enqueue: %v", err)
	}
	if result.(map[string]int64)["id"] != 42 {
		t.Fatalf("result = %v", result)
	}
	if fake.last.Root != contract.RootDownloads || fake.last.Subfolder != "" || fake.last.Install || fake.last.DeleteArchive != nil {
		t.Fatalf("plain enqueue added folder options: %+v", fake.last)
	}
}

func TestDownloadsForwardFolderOptions(t *testing.T) {
	games := t.TempDir()
	env, fake := downloadEnv(t, []string{"$GAMES"}, games)
	keep := false
	_, err := callService(t, env, "downloads.enqueue", enqueueParams(map[string]any{
		"gameId":           5,
		"root":             "games",
		"subfolder":        "Team/Demo",
		"install":          true,
		"installSubfolder": "Pack",
		"deleteArchive":    keep,
	}))
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	want := contract.DownloadRequest{
		GameID: 5, URL: "https://files.example/game.zip", Filename: "game.zip",
		Root: "games", Subfolder: "Team/Demo", Install: true, InstallSubfolder: "Pack", DeleteArchive: &keep,
	}
	got := fake.last
	if got.GameID != want.GameID || got.Root != want.Root || got.Subfolder != want.Subfolder ||
		got.Install != want.Install || got.InstallSubfolder != want.InstallSubfolder ||
		got.DeleteArchive == nil || *got.DeleteArchive != keep {
		t.Fatalf("forwarded = %+v, want %+v", got, want)
	}
}

func TestDownloadsToGamesNeedsGamesGrant(t *testing.T) {
	env, _ := downloadEnv(t, nil, t.TempDir())
	_, err := callService(t, env, "downloads.enqueue", enqueueParams(map[string]any{"root": "games"}))
	expectCode(t, err, CodePermissionDenied)

	_, err = callService(t, env, "downloads.enqueue", enqueueParams(map[string]any{"install": true}))
	expectCode(t, err, CodePermissionDenied)
}

func TestDownloadsInstallNeedsGamesGrantEvenFromDownloadsRoot(t *testing.T) {
	env, fake := downloadEnv(t, []string{"$DOWNLOADS"}, "")
	_, err := callService(t, env, "downloads.enqueue", enqueueParams(map[string]any{"install": true}))
	expectCode(t, err, CodePermissionDenied)
	if fake.last.URL != "" {
		t.Fatal("a denied request must not reach the downloader")
	}
}

func TestDownloadsRejectUnsafeFolderOptions(t *testing.T) {
	env, fake := downloadEnv(t, []string{"$GAMES"}, t.TempDir())
	bad := map[string]map[string]any{
		"traversal":          {"subfolder": "../x"},
		"absolute":           {"subfolder": "/tmp/x"},
		"backslash":          {"subfolder": `a\b`},
		"drive letter":       {"subfolder": "C:/x"},
		"unknown root":       {"root": "music"},
		"install traversal":  {"install": true, "installSubfolder": ".."},
		"negative game id":   {"gameId": -1},
		"empty segment":      {"subfolder": "a//b"},
		"install drive path": {"install": true, "installSubfolder": "D:/games"},
	}
	for name, extra := range bad {
		fake.last = contract.DownloadRequest{}
		_, err := callService(t, env, "downloads.enqueue", enqueueParams(extra))
		if err == nil {
			t.Errorf("%s: enqueue succeeded", name)
			continue
		}
		expectCode(t, err, CodeInvalidParams)
		if fake.last.URL != "" {
			t.Errorf("%s: a rejected request reached the downloader", name)
		}
	}
}

func TestDownloadsMapHostFolderErrors(t *testing.T) {
	env, fake := downloadEnv(t, []string{"$GAMES"}, t.TempDir())
	fake.err = fmt.Errorf("%w: the games folder is not set", contract.ErrBadFolder)
	_, err := callService(t, env, "downloads.enqueue", enqueueParams(map[string]any{"root": "games"}))
	expectCode(t, err, CodeInvalidParams)

	fake.err = fmt.Errorf("disk full")
	_, err = callService(t, env, "downloads.enqueue", enqueueParams(nil))
	expectCode(t, err, CodeInternal)
}

// TestDownloadParamNamesMatchProtocol decodes the params exactly as the SDK schema names them.
func TestDownloadParamNamesMatchProtocol(t *testing.T) {
	games := t.TempDir()
	env, fake := downloadEnv(t, []string{"$GAMES"}, games)
	raw := `{"url":"https://files.example/a.zip","filename":"a.zip","headers":{"X-Token":"t"},` +
		`"gameId":3,"root":"games","subfolder":"Mods","install":true,` +
		`"installSubfolder":"Pack","deleteArchive":true}`
	if _, err := env.call(context.Background(), "downloads.enqueue", json.RawMessage(raw)); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if fake.last.Headers["X-Token"] != "t" || fake.last.GameID != 3 || fake.last.Root != "games" ||
		fake.last.Subfolder != "Mods" || fake.last.InstallSubfolder != "Pack" ||
		fake.last.DeleteArchive == nil || !*fake.last.DeleteArchive {
		t.Fatalf("decoded = %+v", fake.last)
	}
}
