package library

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// gameInFolder saves a game whose install folder is dir, creating the folder with one file.
func gameInFolder(t *testing.T, lib *Library, sourceID, dir string) Game {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "game.bin"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	game := upsertGame(t, lib, sourceID, "Installed")
	if _, err := lib.Patch(context.Background(), game.ID, Patch{InstallDir: &dir}); err != nil {
		t.Fatal(err)
	}
	return game
}

func TestRemoveGameDeletesFolderInsideGamesDir(t *testing.T) {
	games := t.TempDir()
	lib := openTestLibrary(t, Options{GamesDir: func() string { return games }})
	dir := filepath.Join(games, "Harbor")
	game := gameInFolder(t, lib, "rm-1", dir)

	deleted, err := lib.RemoveGame(context.Background(), game.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if !deleted {
		t.Error("the folder inside the games folder must be deleted")
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("install folder still exists: %v", err)
	}
	history, err := lib.ListDeleted(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || !history[0].DeletedFiles {
		t.Errorf("history = %+v, want deletedFiles true", history)
	}
}

func TestRemoveGameKeepsFolderOutsideGamesDir(t *testing.T) {
	games := t.TempDir()
	lib := openTestLibrary(t, Options{GamesDir: func() string { return games }})
	dir := filepath.Join(t.TempDir(), "Elsewhere")
	game := gameInFolder(t, lib, "rm-2", dir)

	deleted, err := lib.RemoveGame(context.Background(), game.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if deleted {
		t.Error("a folder outside the games folder must not be deleted")
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("folder must still exist: %v", err)
	}
	history, _ := lib.ListDeleted(context.Background())
	if len(history) != 1 || history[0].DeletedFiles {
		t.Errorf("history = %+v, want deletedFiles false", history)
	}
}

func TestRemoveGameWithoutDeleteFilesKeepsFolder(t *testing.T) {
	games := t.TempDir()
	lib := openTestLibrary(t, Options{GamesDir: func() string { return games }})
	dir := filepath.Join(games, "Kept")
	game := gameInFolder(t, lib, "rm-3", dir)

	if deleted, err := lib.RemoveGame(context.Background(), game.ID, false); err != nil || deleted {
		t.Fatalf("deleted = %v, err = %v; want kept without an error", deleted, err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("folder must still exist: %v", err)
	}
}

func TestRemoveGameNeverDeletesTheGamesDirItself(t *testing.T) {
	games := t.TempDir()
	lib := openTestLibrary(t, Options{GamesDir: func() string { return games }})
	game := gameInFolder(t, lib, "rm-4", games)

	if deleted, _ := lib.RemoveGame(context.Background(), game.ID, true); deleted {
		t.Error("the games folder itself must never be deleted")
	}
	if _, err := os.Stat(games); err != nil {
		t.Errorf("games folder must still exist: %v", err)
	}
}
