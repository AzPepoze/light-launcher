package runtimesvc

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

// cmdlines builds a readCmdline function from a pid to NUL-separated arguments map.
func cmdlines(values map[int]string) func(int) (string, error) {
	return func(pid int) (string, error) {
		value, ok := values[pid]
		if !ok {
			return "", errors.New("no such process")
		}
		return value, nil
	}
}

func instanceCmdline(gamePath string, gameName string) string {
	args := []string{"light-launcher-instance", "--game", gamePath}
	if gameName != "" {
		args = append(args, "--game-name", gameName)
	}
	return joinNul(args)
}

func joinNul(args []string) string {
	result := ""
	for _, arg := range args {
		result += arg + "\x00"
	}
	return result
}

func TestParseGameArgumentsUsesLastValues(t *testing.T) {
	cmdline := joinNul([]string{"instance", "--game", "/a/One.exe", "--game", "/b/Two.exe", "--game-name", "Two"})

	gamePath, gameName := parseGameArguments(cmdline)
	if gamePath != "/b/Two.exe" || gameName != "Two" {
		t.Errorf("parsed = %q %q, want /b/Two.exe Two", gamePath, gameName)
	}

	if gamePath, _ := parseGameArguments(joinNul([]string{"instance", "--gamescope"})); gamePath != "" {
		t.Errorf("no --game gave %q, want empty", gamePath)
	}
}

func TestParsePids(t *testing.T) {
	got := parsePids("123\n\n abc\n0\n456\n")
	if want := []int{123, 456}; !reflect.DeepEqual(got, want) {
		t.Errorf("parsePids = %v, want %v", got, want)
	}
}

func TestRefreshKeepsStartTimeWhileRunning(t *testing.T) {
	isolateHome(t)
	tracker := &sessionTracker{firstSeen: map[int]int64{}}
	read := cmdlines(map[int]string{42: instanceCmdline("/games/Game.exe", "Game")})

	first, err := tracker.refresh([]int{42}, read, 1000)
	if err != nil {
		t.Fatalf("first refresh: %v", err)
	}
	second, err := tracker.refresh([]int{42}, read, 5000)
	if err != nil {
		t.Fatalf("second refresh: %v", err)
	}

	want := types.RunningSession{Pid: 42, GamePath: "/games/Game.exe", GameName: "Game", StartedAt: 1000}
	if len(first) != 1 || first[0] != want {
		t.Errorf("first scan = %+v, want %+v", first, want)
	}
	if len(second) != 1 || second[0].StartedAt != 1000 {
		t.Errorf("second scan start = %+v, want the first start time 1000", second)
	}
}

func TestRefreshForgetsExitedProcesses(t *testing.T) {
	isolateHome(t)
	tracker := &sessionTracker{firstSeen: map[int]int64{}}
	read := cmdlines(map[int]string{42: instanceCmdline("/games/Game.exe", "")})

	if _, err := tracker.refresh([]int{42}, read, 1000); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if _, err := tracker.refresh(nil, read, 2000); err != nil {
		t.Fatalf("refresh without pids: %v", err)
	}
	restarted, err := tracker.refresh([]int{42}, read, 9000)
	if err != nil {
		t.Fatalf("refresh after restart: %v", err)
	}
	if len(restarted) != 1 || restarted[0].StartedAt != 9000 {
		t.Errorf("restarted session = %+v, want a new start time 9000", restarted)
	}
}

func TestRefreshSkipsUnreadableAndNonGameProcesses(t *testing.T) {
	isolateHome(t)
	tracker := &sessionTracker{firstSeen: map[int]int64{}}
	read := cmdlines(map[int]string{
		7: joinNul([]string{"instance", "--gamescope"}),
	})

	sessions, err := tracker.refresh([]int{7, 8}, read, 1000)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("sessions = %+v, want none", sessions)
	}
}

func TestRefreshDefaultsGameNameFromFileName(t *testing.T) {
	isolateHome(t)
	tracker := &sessionTracker{firstSeen: map[int]int64{}}
	read := cmdlines(map[int]string{5: instanceCmdline("/games/My.Game.exe", "")})

	sessions, err := tracker.refresh([]int{5}, read, 1000)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if len(sessions) != 1 || sessions[0].GameName != "My.Game" {
		t.Errorf("sessions = %+v, want game name My.Game", sessions)
	}
}

func TestSyncActivitySessionsMovesActiveMarker(t *testing.T) {
	isolateHome(t)
	writeFile(t, activityStorePath(), `{"games": {"/old/game.exe": {"gamePath": "/old/Game.exe", "gameName": "Old", "lastPlayedAt": 100, "totalPlaytimeSeconds": 60, "activeSince": 100}}}`)

	session := types.RunningSession{Pid: 1, GamePath: "/new/Game.exe", GameName: "New", StartedAt: 500}
	if err := syncActivitySessions([]types.RunningSession{session}); err != nil {
		t.Fatalf("syncActivitySessions: %v", err)
	}

	store := loadActivityStore()
	if oldEntry := store.Games["/old/game.exe"]; isActive(oldEntry) {
		t.Errorf("old game still active: %+v", oldEntry)
	}
	newEntry, ok := store.Games["/new/game.exe"]
	if !ok || !isActive(newEntry) || *newEntry.ActiveSince != 500 || newEntry.GameName != "New" {
		t.Errorf("new game entry = %+v, want active since 500", newEntry)
	}
}

func TestRecordLaunchStartsSessionAndKeepsPlaytime(t *testing.T) {
	isolateHome(t)
	writeFile(t, activityStorePath(), `{"games": {"/games/game.exe": {"gamePath": "/games/Game.exe", "gameName": "Game", "lastPlayedAt": 10, "totalPlaytimeSeconds": 60}}}`)

	options := types.LaunchOptions{ID: "abc", Name: "Game", GamePath: "/games/Game.exe"}
	if err := recordLaunch(options); err != nil {
		t.Fatalf("recordLaunch: %v", err)
	}

	entry := loadActivityStore().Games["/games/game.exe"]
	if !isActive(entry) || entry.TotalPlaytimeSeconds != 60 || entry.ProfileID != "abc" {
		t.Errorf("entry = %+v, want an active session, kept playtime 60 and profile abc", entry)
	}
	if time.Now().UnixMilli()-entry.LastPlayedAt > 5000 {
		t.Errorf("lastPlayedAt = %d, want about now", entry.LastPlayedAt)
	}
}

func TestRecordLaunchSkipsWhenPlaytimeTrackingIsOff(t *testing.T) {
	isolateHome(t)
	writeFile(t, filepath.Join(config.GetBaseDirectory(), "settings.json"), `{"TrackPlaytime": false}`)

	if err := recordLaunch(types.LaunchOptions{Name: "Game", GamePath: "/games/Game.exe"}); err != nil {
		t.Fatalf("recordLaunch: %v", err)
	}
	if pathExists(activityStorePath()) {
		t.Error("activity.json was written while playtime tracking is off")
	}
}

func TestKillSessionIgnoresInvalidAndMissingPids(t *testing.T) {
	if err := killSession(0); err != nil {
		t.Errorf("killSession(0) = %v, want nil", err)
	}
	if err := killSession(-3); err != nil {
		t.Errorf("killSession(-3) = %v, want nil", err)
	}
	if err := killSession(2147483646); err != nil {
		t.Errorf("killSession of a missing pid = %v, want nil", err)
	}
}
