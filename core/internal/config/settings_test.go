package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"light-launcher/core/internal/types"
)

func TestLoadMissingSettingsReturnsDefaults(t *testing.T) {
	settings := NewSettingsStore(filepath.Join(t.TempDir(), "settings.json")).Load()
	if !settings.TransparentMode {
		t.Error("TransparentMode must default to true")
	}
	if settings.ScanFolders == nil || settings.ScanFolderConfigs == nil || settings.Blacklist == nil {
		t.Error("default lists must be empty, not nil")
	}
}

func TestSettingsRoundTripAndVersion(t *testing.T) {
	target := filepath.Join(t.TempDir(), "settings.json")
	store := NewSettingsStore(target)
	enabled := false
	want := types.AppSettings{TrackPlaytime: &enabled, ScanFolders: []string{}, ScanFolderConfigs: []types.ScanFolderConfig{}, Blacklist: []string{}}
	if err := store.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var raw map[string]any
	if err := readJSON(t, target, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["version"] != float64(DataVersion) {
		t.Errorf("version = %v, want %d", raw["version"], DataVersion)
	}

	loaded := store.Load()
	if loaded.TrackPlaytime == nil || *loaded.TrackPlaytime {
		t.Errorf("TrackPlaytime = %v, want false", loaded.TrackPlaytime)
	}
}

func TestSettingsUnknownVersionResetsToDefaults(t *testing.T) {
	target := filepath.Join(t.TempDir(), "settings.json")
	writeRaw(t, target, `{"version": 7, "TransparentMode": false}`)

	settings := NewSettingsStore(target).Load()
	if !settings.TransparentMode {
		t.Error("an unknown version must reset to defaults")
	}
}

func TestSettingsCorruptFileResetsToDefaults(t *testing.T) {
	target := filepath.Join(t.TempDir(), "settings.json")
	writeRaw(t, target, "{")

	if !NewSettingsStore(target).Load().TransparentMode {
		t.Error("a corrupt file must reset to defaults")
	}
}

func TestSettingsUpdateRewritesFile(t *testing.T) {
	target := filepath.Join(t.TempDir(), "settings.json")
	store := NewSettingsStore(target)
	err := store.Update(func(value *types.AppSettings) error {
		value.Blacklist = append(value.Blacklist, "/games/bad.exe")
		return nil
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got := store.Load().Blacklist; len(got) != 1 || got[0] != "/games/bad.exe" {
		t.Fatalf("Blacklist = %v", got)
	}
}

func TestSettingsNormalizationSavesBackfills(t *testing.T) {
	target := filepath.Join(t.TempDir(), "settings.json")
	writeRaw(t, target, `{"version": 1, "ScanFolders": ["/games"], "ScanFolderConfigs": null, "Blacklist": null}`)

	settings := NewSettingsStore(target).Load()
	if len(settings.ScanFolderConfigs) != 1 || settings.ScanFolderConfigs[0].Depth != 2 {
		t.Fatalf("ScanFolderConfigs = %+v", settings.ScanFolderConfigs)
	}
	if len(settings.ScanFolderConfigs[0].ExcludeNames) != len(defaultExcludeNames) {
		t.Errorf("ExcludeNames = %v, want the defaults", settings.ScanFolderConfigs[0].ExcludeNames)
	}

	var raw struct {
		Version int `json:"version"`
	}
	if err := readJSON(t, target, &raw); err != nil {
		t.Fatal(err)
	}
	if raw.Version != DataVersion {
		t.Errorf("normalized file version = %d, want %d", raw.Version, DataVersion)
	}
}

func TestSettingsDefaultsTrackingOn(t *testing.T) {
	settings := NewSettingsStore(filepath.Join(t.TempDir(), "settings.json")).Load()
	if settings.TrackPlaytime != nil {
		t.Errorf("TrackPlaytime = %v, want nil (on)", settings.TrackPlaytime)
	}
}

func readJSON(t *testing.T, path string, target any) error {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func writeRaw(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
