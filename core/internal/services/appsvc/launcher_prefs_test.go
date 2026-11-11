package appsvc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestLauncherPreferencesSetAndClearPersistToDisk(t *testing.T) {
	path := filepath.Join(t.TempDir(), "launcher-preferences.json")
	folder := filepath.Join(t.TempDir(), "Game")
	launcher := filepath.Join(folder, "Other.exe")

	if err := newLauncherPreferences(path).Set(folder, launcher); err != nil {
		t.Fatalf("Set: %v", err)
	}

	reopened := newLauncherPreferences(path)
	if got := reopened.Preferred()[preferenceKey(folder)]; got != launcher {
		t.Fatalf("Preferred after reopen = %q, want %q", got, launcher)
	}
	assertFileVersion(t, path, launcherPreferencesVersion)

	if err := reopened.Clear(folder); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if _, ok := newLauncherPreferences(path).Preferred()[preferenceKey(folder)]; ok {
		t.Fatal("Clear did not remove the saved launcher")
	}
}

func TestLauncherPreferencesUnreadableFileCountsAsEmpty(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"empty", ""},
		{"corrupt", "{"},
		{"wrong version", `{"version": 99, "folders": {"a": "b"}}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := writeRawPreferences(t, test.content)
			if got := newLauncherPreferences(path).Preferred(); len(got) != 0 {
				t.Fatalf("Preferred = %v, want empty", got)
			}
		})
	}
}

func TestLauncherPreferencesBacksUpUnreadableFileBeforeSave(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"corrupt", "{ not json"},
		{"wrong version", `{"version": 99, "folders": {"a": "b"}}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := writeRawPreferences(t, test.content)
			folder := filepath.Join(t.TempDir(), "Game")

			if err := newLauncherPreferences(path).Set(folder, filepath.Join(folder, "Game.exe")); err != nil {
				t.Fatalf("Set: %v", err)
			}

			backup, err := os.ReadFile(path + ".bak")
			if err != nil {
				t.Fatalf("backup missing: %v", err)
			}
			if string(backup) != test.content {
				t.Fatalf("backup = %q, want the original %q", backup, test.content)
			}
			assertFileVersion(t, path, launcherPreferencesVersion)
			if got := newLauncherPreferences(path).Preferred(); len(got) != 1 {
				t.Fatalf("saved folders = %v, want only the new one", got)
			}
		})
	}
}

func TestLauncherPreferencesConcurrentSetKeepsEveryEntry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "launcher-preferences.json")
	prefs := newLauncherPreferences(path)
	const writers = 40

	var wg sync.WaitGroup
	for i := range writers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			folder := filepath.Join("games", fmt.Sprintf("Game%02d", i))
			if err := prefs.Set(folder, filepath.Join(folder, "Game.exe")); err != nil {
				t.Errorf("Set %d: %v", i, err)
			}
		}()
	}
	wg.Wait()

	if got := len(newLauncherPreferences(path).Preferred()); got != writers {
		t.Fatalf("saved %d folders, want %d", got, writers)
	}
}

func writeRawPreferences(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "launcher-preferences.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertFileVersion(t *testing.T, path string, want int) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var header struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		t.Fatalf("%s is not JSON: %v", path, err)
	}
	if header.Version != want {
		t.Fatalf("version = %d, want %d", header.Version, want)
	}
}
