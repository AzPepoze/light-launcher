package library

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLinkFolderSetsInstallFolder(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	game := upsertGame(t, lib, "link-1", "Linked")
	folder := t.TempDir()
	if err := os.WriteFile(filepath.Join(folder, "data.bin"), []byte("12345"), 0o644); err != nil {
		t.Fatal(err)
	}

	linked, err := lib.LinkFolder(ctx, game.ID, folder)
	if err != nil {
		t.Fatal(err)
	}
	if linked.InstallDir == nil || *linked.InstallDir != folder {
		t.Errorf("installDir = %v, want %s", linked.InstallDir, folder)
	}
	if linked.InstallSize == nil || *linked.InstallSize != 5 {
		t.Errorf("installSize = %v, want 5", linked.InstallSize)
	}
	if linked.Status != "installed" {
		t.Errorf("status = %q, want installed", linked.Status)
	}
}

func TestLinkFolderRefusesOtherGamesFolderAndFiles(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	owner := upsertGame(t, lib, "link-owner", "Owner")
	other := upsertGame(t, lib, "link-other", "Other")
	folder := t.TempDir()
	if _, err := lib.LinkFolder(ctx, owner.ID, folder); err != nil {
		t.Fatal(err)
	}
	if _, err := lib.LinkFolder(ctx, other.ID, folder); err == nil {
		t.Error("a folder owned by another game must be refused")
	}

	file := filepath.Join(folder, "data.bin")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := lib.LinkFolder(ctx, other.ID, file); err == nil {
		t.Error("a file is not a folder and must be refused")
	}
}
