package appsvc

import (
	"path/filepath"
	"testing"

	"light-launcher/core/internal/config"
)

func TestAddScanFolderSavesDefaultConfig(t *testing.T) {
	svc := newTestService(t)
	folder := config.CleanPath(t.TempDir())
	if err := svc.addScanFolder(folder); err != nil {
		t.Fatalf("addScanFolder: %v", err)
	}
	settings := svc.settings.Load()
	if !containsCleanPath(settings.ScanFolders, folder) {
		t.Fatalf("ScanFolders = %v, want %s", settings.ScanFolders, folder)
	}
	cfg := svc.getScanFolderConfig(folder)
	if cfg.Depth != 2 || len(cfg.ExcludeNames) != len(config.DefaultScanExcludeNames()) {
		t.Fatalf("default scan config = %+v", cfg)
	}
}

func TestRenameScanFolderRemapsBlacklist(t *testing.T) {
	svc := newTestService(t)
	oldFolder := config.CleanPath(t.TempDir())
	newFolder := config.CleanPath(t.TempDir())
	if err := svc.addScanFolder(oldFolder); err != nil {
		t.Fatalf("addScanFolder: %v", err)
	}
	blacklisted := filepath.Join(oldFolder, "sub", "game.exe")
	if err := svc.blacklistGame(blacklisted); err != nil {
		t.Fatalf("blacklistGame: %v", err)
	}

	renamed, err := svc.renameScanFolder(oldFolder, newFolder, 3, []string{"Steam.exe"})
	if err != nil {
		t.Fatalf("renameScanFolder: %v", err)
	}
	if renamed != newFolder {
		t.Fatalf("renamed = %q, want %q", renamed, newFolder)
	}
	settings := svc.settings.Load()
	want := filepath.Join(newFolder, "sub", "game.exe")
	if len(settings.Blacklist) != 1 || settings.Blacklist[0] != want {
		t.Fatalf("Blacklist = %v, want [%s]", settings.Blacklist, want)
	}
	if got := svc.getScanFolderConfig(newFolder); got.Depth != 3 || len(got.ExcludeNames) != 1 {
		t.Fatalf("renamed folder config = %+v", got)
	}
}

func TestGetAutoScannedGamesCreatesConfigForNewExecutable(t *testing.T) {
	svc := newTestService(t)
	folder := config.CleanPath(t.TempDir())
	mustWrite(t, filepath.Join(folder, "game.exe"))
	mustWrite(t, filepath.Join(folder, "Steam.exe"))
	if err := svc.addScanFolder(folder); err != nil {
		t.Fatalf("addScanFolder: %v", err)
	}

	groups, err := svc.getAutoScannedGames(true)
	if err != nil {
		t.Fatalf("getAutoScannedGames: %v", err)
	}
	if len(groups) != 1 || len(groups[0].Games) != 1 {
		t.Fatalf("groups = %+v, want one game (Steam.exe is excluded)", groups)
	}
	game := groups[0].Games[0]
	if game.Name != "game" || !game.IsAutoScanned {
		t.Fatalf("game = %+v", game)
	}
	if len(svc.games.List()) != 1 {
		t.Fatal("expected a saved config for the scanned executable")
	}
}
