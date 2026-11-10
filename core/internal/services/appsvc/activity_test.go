package appsvc

import (
	"path/filepath"
	"testing"

	"light-launcher/core/internal/types"
)

func TestRecordLaunchStoresActivityAndActiveMarker(t *testing.T) {
	isolateConfigHome(t)
	gamePath := filepath.Join(t.TempDir(), "Game.exe")
	if err := recordLaunch(types.LaunchOptions{ID: "abc", Name: "Game", GamePath: gamePath}); err != nil {
		t.Fatalf("recordLaunch: %v", err)
	}

	activities := getActivities()
	if len(activities) != 1 {
		t.Fatalf("activities = %+v, want one entry", activities)
	}
	if activities[0].GameName != "Game" || activities[0].ProfileID != "abc" {
		t.Fatalf("activity = %+v", activities[0])
	}
	if !isActive(&activities[0]) {
		t.Fatal("a launched game must carry an active marker")
	}
}

func TestSyncActivitiesClearsStaleMarkers(t *testing.T) {
	isolateConfigHome(t)
	gamePath := filepath.Join(t.TempDir(), "Game.exe")
	if err := recordLaunch(types.LaunchOptions{Name: "Game", GamePath: gamePath}); err != nil {
		t.Fatalf("recordLaunch: %v", err)
	}

	activities, err := syncActivities(nil)
	if err != nil {
		t.Fatalf("syncActivities: %v", err)
	}
	if len(activities) != 1 || isActive(&activities[0]) {
		t.Fatalf("stale marker was not cleared: %+v", activities)
	}
}

func TestSyncActivitiesDisabledTrackingClearsMarkers(t *testing.T) {
	isolateConfigHome(t)
	gamePath := filepath.Join(t.TempDir(), "Game.exe")
	if err := recordLaunch(types.LaunchOptions{Name: "Game", GamePath: gamePath}); err != nil {
		t.Fatalf("recordLaunch: %v", err)
	}
	disabled := false
	if err := saveAppSettings(types.AppSettings{TrackPlaytime: &disabled, ScanFolders: []string{}, ScanFolderConfigs: []types.ScanFolderConfig{}, Blacklist: []string{}}); err != nil {
		t.Fatalf("saveAppSettings: %v", err)
	}

	activities, err := syncActivities([]types.RunningSession{{GamePath: gamePath, GameName: "Game"}})
	if err != nil {
		t.Fatalf("syncActivities: %v", err)
	}
	if len(activities) != 1 || isActive(&activities[0]) {
		t.Fatalf("markers must be cleared when tracking is off: %+v", activities)
	}
}
