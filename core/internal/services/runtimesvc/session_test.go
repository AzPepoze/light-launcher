package runtimesvc

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"light-launcher/core/internal/activity"
	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

// seedActivity stores one entry under the key the store derives from gamePath, so it works on every OS.
func seedActivity(t *testing.T, path, gamePath, entry string) {
	t.Helper()
	key, err := json.Marshal(config.PathKey(gamePath))
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, path, fmt.Sprintf(`{"version": %d, "games": {`, activity.Version)+string(key)+`: `+entry+`}}`)
}

func newTestTracker(t *testing.T) *sessionTracker {
	t.Helper()
	return newSessionTracker(newTestService(t).activity)
}

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
	skipUnlessLinux(t)
	isolateHome(t)
	tracker := newTestTracker(t)
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
	tracker := newTestTracker(t)
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
	tracker := newTestTracker(t)
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
	tracker := newTestTracker(t)
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
	svc := newTestService(t)
	seedActivity(t, config.ActivityFilePath(), "/old/Game.exe", `{"gamePath": "/old/Game.exe", "gameName": "Old", "lastPlayedAt": 100, "totalPlaytimeSeconds": 60, "activeSince": 100}`)

	session := types.RunningSession{Pid: 1, GamePath: "/new/Game.exe", GameName: "New", StartedAt: 500}
	if _, err := svc.activity.Sync([]types.RunningSession{session}); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	entries := svc.activity.List()
	for _, entry := range entries {
		if config.PathKey(entry.GamePath) == config.PathKey("/old/Game.exe") && entry.ActiveSince != nil {
			t.Errorf("old game still active: %+v", entry)
		}
	}
	newEntry := findActivity(entries, "/new/Game.exe")
	if newEntry == nil || newEntry.ActiveSince == nil || *newEntry.ActiveSince != 500 || newEntry.GameName != "New" {
		t.Errorf("new game entry = %+v, want active since 500", newEntry)
	}
}

func TestRecordLaunchStartsSessionAndKeepsPlaytime(t *testing.T) {
	isolateHome(t)
	svc := newTestService(t)
	seedActivity(t, config.ActivityFilePath(), "/games/Game.exe", `{"gamePath": "/games/Game.exe", "gameName": "Game", "lastPlayedAt": 10, "totalPlaytimeSeconds": 60}`)

	options := types.LaunchOptions{ID: "abc", Name: "Game", GamePath: "/games/Game.exe"}
	if err := svc.activity.RecordLaunch(options); err != nil {
		t.Fatalf("RecordLaunch: %v", err)
	}

	entry := findActivity(svc.activity.List(), "/games/Game.exe")
	if entry == nil || entry.ActiveSince == nil || entry.TotalPlaytimeSeconds != 60 || entry.ProfileID != "abc" {
		t.Errorf("entry = %+v, want an active session, kept playtime 60 and profile abc", entry)
	}
	if time.Now().UnixMilli()-entry.LastPlayedAt > 5000 {
		t.Errorf("lastPlayedAt = %d, want about now", entry.LastPlayedAt)
	}
}

func TestRecordLaunchSkipsWhenPlaytimeTrackingIsOff(t *testing.T) {
	isolateHome(t)
	svc := newTestService(t)
	disabled := false
	if err := svc.settings.Save(types.AppSettings{
		TrackPlaytime:     &disabled,
		ScanFolders:       []string{},
		ScanFolderConfigs: []types.ScanFolderConfig{},
		Blacklist:         []string{},
	}); err != nil {
		t.Fatalf("Save settings: %v", err)
	}

	if err := svc.activity.RecordLaunch(types.LaunchOptions{Name: "Game", GamePath: "/games/Game.exe"}); err != nil {
		t.Fatalf("RecordLaunch: %v", err)
	}
	if config.PathExists(config.ActivityFilePath()) {
		t.Error("activity.json was written while playtime tracking is off")
	}
}

func findActivity(entries []types.GameActivity, gamePath string) *types.GameActivity {
	key := config.PathKey(gamePath)
	for i := range entries {
		if config.PathKey(entries[i].GamePath) == key {
			return &entries[i]
		}
	}
	return nil
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
