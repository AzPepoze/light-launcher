package appsvc

import (
	"path/filepath"
	"testing"
)

func TestAddScanFolderSavesDefaultConfig(t *testing.T) {
	isolateConfigHome(t)
	folder := cleanPath(t.TempDir())
	if err := addScanFolder(folder); err != nil {
		t.Fatalf("addScanFolder: %v", err)
	}
	settings := loadAppSettings()
	if !containsCleanPath(settings.ScanFolders, folder) {
		t.Fatalf("ScanFolders = %v, want %s", settings.ScanFolders, folder)
	}
	cfg := getScanFolderConfig(folder)
	if cfg.Depth != 2 || len(cfg.ExcludeNames) != len(defaultExcludeNames) {
		t.Fatalf("default scan config = %+v", cfg)
	}
}

func TestRenameScanFolderRemapsBlacklist(t *testing.T) {
	isolateConfigHome(t)
	oldFolder := cleanPath(t.TempDir())
	newFolder := cleanPath(t.TempDir())
	if err := addScanFolder(oldFolder); err != nil {
		t.Fatalf("addScanFolder: %v", err)
	}
	blacklisted := filepath.Join(oldFolder, "sub", "game.exe")
	if err := blacklistGame(blacklisted); err != nil {
		t.Fatalf("blacklistGame: %v", err)
	}

	renamed, err := renameScanFolder(oldFolder, newFolder, 3, []string{"Steam.exe"})
	if err != nil {
		t.Fatalf("renameScanFolder: %v", err)
	}
	if renamed != newFolder {
		t.Fatalf("renamed = %q, want %q", renamed, newFolder)
	}
	settings := loadAppSettings()
	want := filepath.Join(newFolder, "sub", "game.exe")
	if len(settings.Blacklist) != 1 || settings.Blacklist[0] != want {
		t.Fatalf("Blacklist = %v, want [%s]", settings.Blacklist, want)
	}
	if got := getScanFolderConfig(newFolder); got.Depth != 3 || len(got.ExcludeNames) != 1 {
		t.Fatalf("renamed folder config = %+v", got)
	}
}

func TestGetAutoScannedGamesCreatesConfigForNewExecutable(t *testing.T) {
	isolateConfigHome(t)
	folder := cleanPath(t.TempDir())
	mustWrite(t, filepath.Join(folder, "game.exe"))
	mustWrite(t, filepath.Join(folder, "Steam.exe"))
	if err := addScanFolder(folder); err != nil {
		t.Fatalf("addScanFolder: %v", err)
	}

	groups, err := getAutoScannedGames(true)
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
	if len(listGameConfigs()) != 1 {
		t.Fatal("expected a saved config for the scanned executable")
	}
}
