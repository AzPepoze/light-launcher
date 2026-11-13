package library

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/types"
)

// refreshInfo is the source list entry of a plugin that can fetch a game and resolve its files.
func refreshInfo(capabilities ...string) []contract.SourceInfo {
	if capabilities == nil {
		capabilities = []string{CapabilityResolveFiles}
	}
	return []contract.SourceInfo{{Plugin: "f95plugin", SourceID: "f95", Label: "F95", Capabilities: capabilities}}
}

// refreshedDTO is the game a fake source reports on a refresh.
func refreshedDTO() contract.GameDTO {
	rating := 4.5
	return contract.GameDTO{
		Source:          "f95",
		SourceID:        "42",
		URL:             "https://f95.example/threads/42",
		Title:           "Fresh Title",
		Version:         "2.0",
		Developer:       "New Studio",
		Engine:          "Ren'Py",
		Tags:            []string{"source tag"},
		DescriptionHTML: "<p>fresh description</p>",
		Changelog:       "fresh changelog",
		Rating:          &rating,
		Status:          "ongoing",
		UpdatedAt:       "2026-11-01T10:00:00Z",
		Extra:           map[string]any{"threadId": "42"},
	}
}

func boolPtr(value bool) *bool {
	return &value
}

func int64Ptr(value int64) *int64 {
	return &value
}

// upsertF95Game stores a plugin game of the f95 source and returns the saved row.
func upsertF95Game(t *testing.T, lib *Library, sourceID, title string) Game {
	t.Helper()
	ctx := context.Background()
	id, err := lib.Upsert(ctx, contract.GameDTO{
		Source: "f95", SourceID: sourceID, URL: "https://f95.example/threads/" + sourceID, Title: title,
	})
	if err != nil {
		t.Fatalf("upsert %s: %v", title, err)
	}
	game, err := lib.Get(ctx, id)
	if err != nil {
		t.Fatalf("get %d: %v", id, err)
	}
	return game
}

func TestRefreshGameFilesReplacesPluginFileIDsAndKeepsUserFields(t *testing.T) {
	ctx := context.Background()
	var (
		fetchParams   fetchGameParams
		resolveParams gameParams
	)
	sources := &fakeSources{
		infos: refreshInfo(),
		replies: map[string]func(params any) (any, error){
			methodFetchGame: func(params any) (any, error) {
				fetchParams = params.(fetchGameParams)
				return fetchGameReply{Game: refreshedDTO()}, nil
			},
			methodResolveFiles: func(params any) (any, error) {
				resolveParams = params.(gameParams)
				return filesReply{Files: []contract.FileCandidate{{
					ID: "fresh-win", Label: "Windows", URL: "https://files.example/fresh.zip",
					Kind: "direct", Platform: contract.PlatformWindows,
				}}}, nil
			},
		},
	}
	lib := openTestLibrary(t, Options{Sources: sources})

	gameID, err := lib.Upsert(ctx, contract.GameDTO{
		Source: "f95", SourceID: "42", URL: "https://f95.example/threads/42",
		Title: "Old Title", Version: "1.0",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := lib.ReplaceFiles(ctx, gameID, []contract.FileCandidate{{
		Label: "Old file", URL: "https://files.example/old.zip", Kind: "direct",
	}}); err != nil {
		t.Fatal(err)
	}
	before, err := lib.ListFiles(ctx, gameID)
	if err != nil {
		t.Fatal(err)
	}
	if len(before) != 1 {
		t.Fatalf("stored files = %d, want the old one", len(before))
	}

	installDir := filepath.Join(t.TempDir(), "game")
	exePath := filepath.Join(installDir, "game.exe")
	patch := Patch{
		InstallDir:       &installDir,
		InstallSize:      int64Ptr(1234),
		ExePath:          &exePath,
		ExeArgs:          strPtr("--windowed"),
		PlaytimeS:        int64Ptr(3600),
		Favorite:         boolPtr(true),
		Notes:            strPtr("keep this note"),
		Status:           strPtr("installed"),
		Tags:             &[]string{"user tag"},
		InstalledVersion: strPtr("1.0"),
		LastPlayedAt:     strPtr("2026-01-02T03:04:05.000Z"),
	}
	if _, err := lib.Patch(ctx, gameID, patch); err != nil {
		t.Fatal(err)
	}
	local := types.LaunchOptions{ID: "saved-local", Name: "Local settings"}
	raw, err := json.Marshal(local)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lib.db.ExecContext(ctx, `update games set local_config = ? where id = ?`, string(raw), gameID); err != nil {
		t.Fatal(err)
	}

	detail, err := lib.RefreshGameFiles(ctx, gameID)
	if err != nil {
		t.Fatalf("RefreshGameFiles: %v", err)
	}

	// The plugin was asked about the stored game, and the file query carried the refreshed row.
	if fetchParams.Source != "f95" || fetchParams.Ref.ID != "42" || fetchParams.Ref.URL != "https://f95.example/threads/42" {
		t.Errorf("fetch params = %+v, want the stored source, id and url", fetchParams)
	}
	if resolveParams.Source != "f95" || resolveParams.Game.ID != gameID ||
		resolveParams.Game.Title != "Fresh Title" || resolveParams.Game.Version == nil || *resolveParams.Game.Version != "2.0" {
		t.Errorf("resolve params = %+v, want the refreshed game", resolveParams)
	}

	// Metadata comes from the reply.
	checkString(t, "title", &detail.Title, "Fresh Title")
	checkString(t, "version", detail.Version, "2.0")
	checkString(t, "developer", detail.Developer, "New Studio")
	checkString(t, "engine", detail.Engine, "Ren'Py")
	checkString(t, "description", detail.Description, "<p>fresh description</p>")
	checkString(t, "changelog", detail.Changelog, "fresh changelog")
	checkString(t, "sourceUpdatedAt", detail.SourceUpdatedAt, "2026-11-01T10:00:00Z")
	checkString(t, "status", &detail.Status, "installed")
	if detail.Rating == nil || *detail.Rating != 4.5 {
		t.Errorf("rating = %v, want 4.5", detail.Rating)
	}

	// The file list is replaced, with the plugin file id prepareDownload needs.
	if len(detail.Files) != 1 || detail.Files[0].Label != "Windows" {
		t.Fatalf("files = %+v, want the refreshed Windows file", detail.Files)
	}
	if detail.Files[0].ID == before[0].ID {
		t.Errorf("file row id = %d, want a new row", detail.Files[0].ID)
	}
	candidate, err := lib.candidate(ctx, gameID, detail.Files[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if candidate.ID != "fresh-win" || candidate.URL != "https://files.example/fresh.zip" {
		t.Errorf("stored candidate = %+v, want the plugin id and url", candidate)
	}
	stored, err := lib.ListFiles(ctx, gameID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != 1 {
		t.Errorf("stored files = %d, want only the refreshed one", len(stored))
	}

	// The user's fields survive the refresh.
	checkString(t, "install dir", detail.InstallDir, installDir)
	checkString(t, "exe path", detail.ExePath, exePath)
	checkString(t, "exe args", detail.ExeArgs, "--windowed")
	checkString(t, "notes", detail.Notes, "keep this note")
	checkString(t, "installed version", detail.InstalledVersion, "1.0")
	checkString(t, "last played", detail.LastPlayedAt, "2026-01-02T03:04:05.000Z")
	if detail.InstallSize == nil || *detail.InstallSize != 1234 {
		t.Errorf("install size = %v, want 1234", detail.InstallSize)
	}
	if detail.PlaytimeS != 3600 {
		t.Errorf("playtime = %d, want 3600", detail.PlaytimeS)
	}
	if detail.Favorite != 1 {
		t.Errorf("favorite = %d, want 1", detail.Favorite)
	}
	if !slices.Equal(detail.Tags, []string{"user tag"}) {
		t.Errorf("tags = %v, want the user tags", detail.Tags)
	}
	if detail.LocalConfig == nil || detail.LocalConfig.ID != "saved-local" {
		t.Errorf("local config = %+v, want the saved one", detail.LocalConfig)
	}
}

func TestRefreshGameFilesWithoutFetchGameLeavesTheRowAlone(t *testing.T) {
	ctx := context.Background()
	resolved := false
	sources := &fakeSources{
		infos: refreshInfo(),
		replies: map[string]func(params any) (any, error){
			methodResolveFiles: func(params any) (any, error) {
				resolved = true
				return filesReply{Files: []contract.FileCandidate{{ID: "x", Label: "X", URL: "https://x", Kind: "direct"}}}, nil
			},
		},
	}
	lib := openTestLibrary(t, Options{Sources: sources})
	game := upsertF95Game(t, lib, "42", "Stored Title")
	if err := lib.ReplaceFiles(ctx, game.ID, []contract.FileCandidate{{
		Label: "Stored file", URL: "https://files.example/old.zip", Kind: "direct",
	}}); err != nil {
		t.Fatal(err)
	}

	_, err := lib.RefreshGameFiles(ctx, game.ID)
	if !errors.Is(err, ErrFetchGame) {
		t.Fatalf("err = %v, want ErrFetchGame", err)
	}
	if !strings.Contains(err.Error(), "fetch game") {
		t.Errorf("err = %v, want it to say the game fetch failed", err)
	}
	if resolved {
		t.Error("the files were resolved after the game fetch failed")
	}
	stored, err := lib.Get(ctx, game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Title != game.Title {
		t.Errorf("title = %q, want the stored %q", stored.Title, game.Title)
	}
	files, err := lib.ListFiles(ctx, game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Label != "Stored file" {
		t.Errorf("files = %+v, want the stored one", files)
	}
}

func TestRefreshGameFilesWithoutResolveFilesKeepsStoredFiles(t *testing.T) {
	ctx := context.Background()
	sources := &fakeSources{
		infos: []contract.SourceInfo{{Plugin: "f95plugin", SourceID: "f95", Label: "F95"}},
		replies: map[string]func(params any) (any, error){
			methodFetchGame: func(any) (any, error) {
				return fetchGameReply{Game: refreshedDTO()}, nil
			},
		},
	}
	lib := openTestLibrary(t, Options{Sources: sources})
	game := upsertF95Game(t, lib, "42", "Stored Title")
	if err := lib.ReplaceFiles(ctx, game.ID, []contract.FileCandidate{{
		Label: "Stored file", URL: "https://files.example/old.zip", Kind: "direct",
	}}); err != nil {
		t.Fatal(err)
	}

	detail, err := lib.RefreshGameFiles(ctx, game.ID)
	if err != nil {
		t.Fatalf("RefreshGameFiles: %v", err)
	}
	if detail.Title != "Fresh Title" {
		t.Errorf("title = %q, want the refreshed one", detail.Title)
	}
	if len(detail.Files) != 1 || detail.Files[0].Label != "Stored file" {
		t.Errorf("files = %+v, want the stored one untouched", detail.Files)
	}
}

func TestRefreshGameFilesRejectsGameWithoutPluginSource(t *testing.T) {
	lib := openTestLibrary(t, Options{Sources: &fakeSources{infos: refreshInfo()}})
	game := upsertGame(t, lib, "42", "Manual Game")

	_, err := lib.RefreshGameFiles(context.Background(), game.ID)
	if !errors.Is(err, ErrUnknownSource) {
		t.Fatalf("err = %v, want ErrUnknownSource", err)
	}
}

func TestRefreshGameFilesRejectsUnknownGame(t *testing.T) {
	lib := openTestLibrary(t, Options{Sources: &fakeSources{infos: refreshInfo()}})
	if _, err := lib.RefreshGameFiles(context.Background(), 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestStrictRefreshGameFilesSendsSDKShapes(t *testing.T) {
	ctx := context.Background()
	lib := openTestLibrary(t, Options{Sources: strictF95Sources()})
	gameID, err := lib.Upsert(ctx, contract.GameDTO{
		Source: "f95", SourceID: "42", URL: "https://f95/42", Title: "Old Strict", Version: "0.5",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := lib.ReplaceFiles(ctx, gameID, []contract.FileCandidate{{
		Label: "Old", URL: "https://f95/old.zip", Kind: "direct",
	}}); err != nil {
		t.Fatal(err)
	}

	detail, err := lib.RefreshGameFiles(ctx, gameID)
	if err != nil {
		t.Fatalf("RefreshGameFiles: %v", err)
	}
	if detail.Title != "Strict Game" {
		t.Errorf("title = %q, want the strict reply", detail.Title)
	}
	if len(detail.Files) != 1 || detail.Files[0].Label != "Windows" {
		t.Fatalf("files = %+v, want the strict resolved file", detail.Files)
	}
	candidate, err := lib.candidate(ctx, gameID, detail.Files[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if candidate.ID != "dl" {
		t.Errorf("plugin file id = %q, want dl", candidate.ID)
	}
}
