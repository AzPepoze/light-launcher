package appsvc

import (
	"path/filepath"
	"slices"
	"testing"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

type scanEventRecorder struct {
	events []types.HomeScanEvent
}

func (r *scanEventRecorder) Emit(channel string, data any) {
	if channel == homeScanChannel {
		r.events = append(r.events, data.(types.HomeScanEvent))
	}
}

// progressEvents drops the per-folder group events, leaving only progress.
func progressEvents(events []types.HomeScanEvent) []types.HomeScanEvent {
	progress := make([]types.HomeScanEvent, 0, len(events))
	for _, event := range events {
		if event.Group == nil {
			progress = append(progress, event)
		}
	}
	return progress
}

func TestGetAutoScannedGamesReportsFolderProgress(t *testing.T) {
	svc := newTestService(t)
	folders := make([]string, 0, 2)
	for range 2 {
		folder := config.CleanPath(t.TempDir())
		folders = append(folders, folder)
		mustWrite(t, filepath.Join(folder, "game.exe"))
		if err := svc.addScanFolder(folder); err != nil {
			t.Fatalf("addScanFolder: %v", err)
		}
	}
	recorder := &scanEventRecorder{}
	svc.events = recorder

	if _, err := svc.getAutoScannedGames(true); err != nil {
		t.Fatalf("getAutoScannedGames: %v", err)
	}
	want := []types.HomeScanEvent{
		{State: types.HomeScanRunning, Done: 0, Total: 2, Found: 0, Current: filepath.Base(folders[0])},
		{State: types.HomeScanRunning, Done: 0, Total: 2, Found: 1, Current: filepath.Base(folders[0])},
		{State: types.HomeScanRunning, Done: 1, Total: 2, Found: 1, Current: filepath.Base(folders[1])},
		{State: types.HomeScanRunning, Done: 1, Total: 2, Found: 2, Current: filepath.Base(folders[1])},
		{State: types.HomeScanDone, Done: 2, Total: 2, Found: 2},
	}
	if got := progressEvents(recorder.events); !slices.Equal(got, want) {
		t.Fatalf("scan events = %+v, want %+v", got, want)
	}
}

func TestGetAutoScannedGamesSendsGroupPerFolderBeforeDone(t *testing.T) {
	svc := newTestService(t)
	folders := make([]string, 0, 2)
	for range 2 {
		folder := config.CleanPath(t.TempDir())
		folders = append(folders, folder)
		mustWrite(t, filepath.Join(folder, "game.exe"))
		if err := svc.addScanFolder(folder); err != nil {
			t.Fatalf("addScanFolder: %v", err)
		}
	}
	recorder := &scanEventRecorder{}
	svc.events = recorder

	if _, err := svc.getAutoScannedGames(true); err != nil {
		t.Fatalf("getAutoScannedGames: %v", err)
	}
	events := recorder.events
	last := len(events) - 1
	if events[last].State != types.HomeScanDone {
		t.Fatalf("last scan event = %+v, want done", events[last])
	}
	var groupEvents []types.HomeScanEvent
	for _, event := range events[:last] {
		if event.Group != nil {
			groupEvents = append(groupEvents, event)
		}
	}
	if len(groupEvents) != len(folders) {
		t.Fatalf("group events = %d, want %d", len(groupEvents), len(folders))
	}
	for i, event := range groupEvents {
		if event.State != types.HomeScanRunning || event.Done != i+1 {
			t.Fatalf("group event %d = %+v, want running with done %d", i, event, i+1)
		}
		if event.Group.FolderPath != folders[i] {
			t.Fatalf("group event %d folder = %q, want %q", i, event.Group.FolderPath, folders[i])
		}
	}
}

func TestGetAutoScannedGamesReportsFoundWithinOneFolder(t *testing.T) {
	svc := newTestService(t)
	folder := config.CleanPath(t.TempDir())
	for _, name := range []string{"first", "second"} {
		mustWrite(t, filepath.Join(folder, name, name+".exe"))
	}
	if err := svc.addScanFolder(folder); err != nil {
		t.Fatalf("addScanFolder: %v", err)
	}
	recorder := &scanEventRecorder{}
	svc.events = recorder

	if _, err := svc.getAutoScannedGames(true); err != nil {
		t.Fatalf("getAutoScannedGames: %v", err)
	}
	want := []types.HomeScanEvent{
		{State: types.HomeScanRunning, Done: 0, Total: 1, Found: 0, Current: filepath.Base(folder)},
		{State: types.HomeScanRunning, Done: 0, Total: 1, Found: 1, Current: filepath.Base(folder)},
		{State: types.HomeScanRunning, Done: 0, Total: 1, Found: 2, Current: filepath.Base(folder)},
		{State: types.HomeScanDone, Done: 1, Total: 1, Found: 2},
	}
	if got := progressEvents(recorder.events); !slices.Equal(got, want) {
		t.Fatalf("scan events = %+v, want %+v", got, want)
	}
}

func TestGetAutoScannedGamesWithoutFoldersReportsNoEvents(t *testing.T) {
	svc := newTestService(t)
	recorder := &scanEventRecorder{}
	svc.events = recorder

	if _, err := svc.getAutoScannedGames(true); err != nil {
		t.Fatalf("getAutoScannedGames: %v", err)
	}
	if len(recorder.events) != 0 {
		t.Fatalf("scan events = %+v, want none", recorder.events)
	}
}

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
