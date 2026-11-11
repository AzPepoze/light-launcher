package activity

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

type stubSettings struct {
	value types.AppSettings
}

func (s stubSettings) Load() types.AppSettings {
	return s.value
}

func newStoreAt(t *testing.T, path string, track *bool) *Store {
	t.Helper()
	return New(path, stubSettings{value: types.AppSettings{TrackPlaytime: track}})
}

func newStore(t *testing.T, track *bool) *Store {
	t.Helper()
	return newStoreAt(t, filepath.Join(t.TempDir(), "activity.json"), track)
}

func TestRecordLaunchStoresActivityAndActiveMarker(t *testing.T) {
	store := newStore(t, nil)
	gamePath := filepath.Join(t.TempDir(), "Game.exe")
	if err := store.RecordLaunch(types.LaunchOptions{ID: "abc", Name: "Game", GamePath: gamePath}); err != nil {
		t.Fatalf("RecordLaunch: %v", err)
	}

	activities := store.List()
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

func TestRecordLaunchKeepsPlaytimeAndProfile(t *testing.T) {
	store := newStore(t, nil)
	gamePath := "/games/Game.exe"
	seed(t, store, gamePath, `{"gamePath": "/games/Game.exe", "gameName": "Game", "lastPlayedAt": 10, "totalPlaytimeSeconds": 60}`)

	if err := store.RecordLaunch(types.LaunchOptions{ID: "abc", Name: "Game", GamePath: gamePath}); err != nil {
		t.Fatalf("RecordLaunch: %v", err)
	}

	entry := store.List()[0]
	if !isActive(&entry) || entry.TotalPlaytimeSeconds != 60 || entry.ProfileID != "abc" {
		t.Errorf("entry = %+v, want an active session, kept playtime 60 and profile abc", entry)
	}
	if time.Now().UnixMilli()-entry.LastPlayedAt > 5000 {
		t.Errorf("lastPlayedAt = %d, want about now", entry.LastPlayedAt)
	}
}

func TestSyncClearsStaleMarkers(t *testing.T) {
	store := newStore(t, nil)
	gamePath := filepath.Join(t.TempDir(), "Game.exe")
	if err := store.RecordLaunch(types.LaunchOptions{Name: "Game", GamePath: gamePath}); err != nil {
		t.Fatalf("RecordLaunch: %v", err)
	}

	activities, err := store.Sync(nil)
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if len(activities) != 1 || isActive(&activities[0]) {
		t.Fatalf("stale marker was not cleared: %+v", activities)
	}
}

func TestSyncMarksRunningSessions(t *testing.T) {
	store := newStore(t, nil)
	session := types.RunningSession{GamePath: "/new/Game.exe", GameName: "New", StartedAt: 500}
	if _, err := store.Sync([]types.RunningSession{session}); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	entries := store.List()
	if len(entries) != 1 || !isActive(&entries[0]) || *entries[0].ActiveSince != 500 {
		t.Fatalf("entries = %+v, want an active session since 500", entries)
	}
}

func TestSyncDisabledTrackingClearsMarkers(t *testing.T) {
	store := newStore(t, nil)
	gamePath := filepath.Join(t.TempDir(), "Game.exe")
	if err := store.RecordLaunch(types.LaunchOptions{Name: "Game", GamePath: gamePath}); err != nil {
		t.Fatalf("RecordLaunch: %v", err)
	}
	disabled := false
	store.settings = stubSettings{value: types.AppSettings{TrackPlaytime: &disabled}}

	activities, err := store.Sync([]types.RunningSession{{GamePath: gamePath, GameName: "Game"}})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if len(activities) != 1 || isActive(&activities[0]) {
		t.Fatalf("markers must be cleared when tracking is off: %+v", activities)
	}
}

func TestRecordLaunchSkippedWhenTrackingIsOff(t *testing.T) {
	disabled := false
	path := filepath.Join(t.TempDir(), "activity.json")
	store := newStoreAt(t, path, &disabled)

	if err := store.RecordLaunch(types.LaunchOptions{Name: "Game", GamePath: "/games/Game.exe"}); err != nil {
		t.Fatalf("RecordLaunch: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("activity.json was written while playtime tracking is off")
	}
}

func TestFinalizeAddsElapsedPlaytime(t *testing.T) {
	store := newStore(t, nil)
	gamePath := "/games/Game.exe"
	start := time.UnixMilli(1_000)
	seed(t, store, gamePath, `{"gamePath": "/games/Game.exe", "gameName": "Game", "lastPlayedAt": 1000, "activeSince": 1000}`)

	if err := store.Finalize(gamePath, "Game", start, start.Add(90*time.Second)); err != nil {
		t.Fatalf("Finalize: %v", err)
	}

	entry := store.List()[0]
	if entry.TotalPlaytimeSeconds != 90 || isActive(&entry) {
		t.Errorf("entry = %+v, want 90 seconds and no active marker", entry)
	}
}

func TestFinalizeClampsEndBeforeStart(t *testing.T) {
	store := newStore(t, nil)
	gamePath := "/games/Game.exe"
	start := time.UnixMilli(5_000)
	seed(t, store, gamePath, `{"gamePath": "/games/Game.exe", "gameName": "Game", "lastPlayedAt": 5000, "activeSince": 5000}`)

	if err := store.Finalize(gamePath, "Game", start, start.Add(-time.Minute)); err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	if entry := store.List()[0]; entry.TotalPlaytimeSeconds != 0 {
		t.Errorf("TotalPlaytimeSeconds = %d, want 0", entry.TotalPlaytimeSeconds)
	}
}

func TestPathVariantsShareOneEntry(t *testing.T) {
	store := newStore(t, nil)
	for _, path := range []string{`C:\Games\Game.exe`, `c:/games/game.exe`, `C:\Games\..\Games\Game.exe`} {
		if err := store.RecordLaunch(types.LaunchOptions{Name: "Game", GamePath: path}); err != nil {
			t.Fatalf("RecordLaunch(%q): %v", path, err)
		}
	}
	if got := len(store.List()); got != 1 {
		t.Fatalf("List length = %d, want one shared entry", got)
	}
}

func TestCorruptAndOldStoresReset(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"corrupt", "{"},
		{"missing version", `{"games": {"/game": {"gameName": "Old"}}}`},
		{"unknown version", `{"version": 99, "games": {"/game": {"gameName": "Old"}}}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "activity.json")
			if err := os.WriteFile(path, []byte(test.content), 0o644); err != nil {
				t.Fatal(err)
			}
			if got := New(path, stubSettings{}).List(); len(got) != 0 {
				t.Fatalf("List = %+v, want an empty store", got)
			}
		})
	}
}

func TestConcurrentRecordLaunchKeepsEveryEntry(t *testing.T) {
	store := newStore(t, nil)
	var wait sync.WaitGroup
	for i := 0; i < 20; i++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			path := fmt.Sprintf("/games/Game%d.exe", index)
			if err := store.RecordLaunch(types.LaunchOptions{Name: "Game", GamePath: path}); err != nil {
				t.Errorf("RecordLaunch: %v", err)
			}
		}(i)
	}
	wait.Wait()

	if got := len(store.List()); got != 20 {
		t.Fatalf("List length = %d, want 20", got)
	}
}

func seed(t *testing.T, store *Store, gamePath, entry string) {
	t.Helper()
	var activity types.GameActivity
	if err := json.Unmarshal([]byte(entry), &activity); err != nil {
		t.Fatalf("seed entry: %v", err)
	}
	if err := store.file.Save(activityFile{Games: map[string]*types.GameActivity{
		config.PathKey(gamePath): &activity,
	}}); err != nil {
		t.Fatalf("seed: %v", err)
	}
}
