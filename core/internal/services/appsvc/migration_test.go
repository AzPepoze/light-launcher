package appsvc

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"light-launcher/core/internal/config"
)

func TestLegacyDirectoryIsHomeLightLauncher(t *testing.T) {
	isolateConfigHome(t)
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory")
	}
	if got := legacyDirectory(); got != filepath.Join(home, "LightLauncher") {
		t.Fatalf("legacyDirectory = %q", got)
	}
}

func TestMigrateIfNeededWithoutLegacyFolderReturnsFalse(t *testing.T) {
	isolateConfigHome(t)
	migrated, err := migrateIfNeeded()
	if err != nil || migrated {
		t.Fatalf("migrateIfNeeded = %v, %v; want false and no error", migrated, err)
	}
}

func TestMigrateIfNeededMovesSettingsAndRepairsPrefixPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("legacy prefix paths use forward slashes")
	}
	isolateConfigHome(t)
	legacy := legacyDirectory()
	legacyConfig := filepath.Join(legacy, "config", "executables", "abc", "config.json")
	mustWriteFile(t, filepath.Join(legacy, "settings.json"), `{"TransparentMode":true}`)
	legacyPrefix := filepath.Join(legacy, "prefixes", "Pfx")
	mustWriteFile(t, legacyConfig, `{"ID":"abc","Name":"Game","PrefixPath":"`+legacyPrefix+`","Extras":{"Gamescope":{"Width":"1280"}}}`)

	migrated, err := migrateIfNeeded()
	if err != nil || !migrated {
		t.Fatalf("migrateIfNeeded = %v, %v; want true and no error", migrated, err)
	}
	if _, err := os.Stat(settingsFilePath()); err != nil {
		t.Fatalf("settings were not moved: %v", err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatalf("legacy folder should be removed, stat error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(config.GetConfigDirectory(), "abc", "config.json"))
	if err != nil {
		t.Fatalf("migrated config missing: %v", err)
	}
	var cfg struct {
		PrefixPath string
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("migrated config is not JSON: %v", err)
	}
	want := filepath.Join(config.GetPrefixBaseDirectory(), "Pfx")
	if cfg.PrefixPath != want || strings.Contains(string(data), "LightLauncher") {
		t.Fatalf("PrefixPath = %q, want %q", cfg.PrefixPath, want)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
