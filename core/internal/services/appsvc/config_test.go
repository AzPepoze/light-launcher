package appsvc

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"light-launcher/core/internal/types"
)

// isolateConfigHome points every config and home lookup at a temp folder.
func isolateConfigHome(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("APPDATA", dir)
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}

func TestGenerateIDIsSixteenHexChars(t *testing.T) {
	first := generateID()
	second := generateID()
	hexID := regexp.MustCompile(`^[0-9a-f]{16}$`)
	if !hexID.MatchString(first) || !hexID.MatchString(second) {
		t.Fatalf("ids must be 16 hex chars, got %q and %q", first, second)
	}
	if first == second {
		t.Fatal("two generated ids must differ")
	}
}

func TestLoadPrefixConfigReturnsDefaultsForUnknownPrefix(t *testing.T) {
	isolateConfigHome(t)
	cfg := loadPrefixConfig("NonExistentPrefix_123")
	if cfg.Name != "NonExistentPrefix_123" {
		t.Fatalf("Name = %q", cfg.Name)
	}
	if cfg.Extras.Gamescope.Width != "1280" || cfg.Extras.Gamescope.Height != "720" {
		t.Errorf("gamescope size = %sx%s, want 1280x720", cfg.Extras.Gamescope.Width, cfg.Extras.Gamescope.Height)
	}
	if cfg.Extras.Memory.Value != "4G" {
		t.Errorf("memory = %q, want 4G", cfg.Extras.Memory.Value)
	}
	if cfg.Extras.Lsfg.Multiplier != "2" {
		t.Errorf("lsfg multiplier = %q, want 2", cfg.Extras.Lsfg.Multiplier)
	}
}

func TestSaveAndLoadJSONRoundTrip(t *testing.T) {
	target := filepath.Join(t.TempDir(), "nested", "value.json")
	type payload struct {
		Hello string `json:"hello"`
		Test  int    `json:"test"`
	}
	if err := saveJSON(target, payload{Hello: "world", Test: 123}); err != nil {
		t.Fatalf("saveJSON: %v", err)
	}
	loaded, err := loadJSON[payload](target)
	if err != nil {
		t.Fatalf("loadJSON: %v", err)
	}
	if loaded.Hello != "world" || loaded.Test != 123 {
		t.Fatalf("loaded = %+v", loaded)
	}
}

func TestSaveJSONKeepsAmpersandsUnescaped(t *testing.T) {
	data, err := marshalJSON(map[string]string{"path": "a&b<c>"})
	if err != nil {
		t.Fatalf("marshalJSON: %v", err)
	}
	if !strings.Contains(string(data), "a&b<c>") {
		t.Fatalf("marshalJSON escaped HTML characters: %s", data)
	}
}

func TestLoadJSONRejectsEmptyFiles(t *testing.T) {
	target := filepath.Join(t.TempDir(), "empty.json")
	if err := os.WriteFile(target, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := loadJSON[types.AppSettings](target)
	if err == nil || !strings.Contains(err.Error(), "is empty") {
		t.Fatalf("loadJSON of empty file error = %v, want an 'is empty' error", err)
	}
}
