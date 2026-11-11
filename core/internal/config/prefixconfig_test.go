package config

import (
	"path/filepath"
	"testing"

	"light-launcher/core/internal/types"
)

type stubSettings struct {
	value types.AppSettings
}

func (s stubSettings) Load() types.AppSettings {
	return s.value
}

func TestPrefixStoreDefaultsWhenMissing(t *testing.T) {
	store := NewPrefixStore(stubSettings{})
	cfg := store.Load("Missing")
	if cfg.Name != "Missing" {
		t.Errorf("Name = %q", cfg.Name)
	}
	if cfg.Extras.Gamescope.Width != DefaultWidth || cfg.Extras.Memory.Value != DefaultMemoryValue {
		t.Errorf("extras = %+v", cfg.Extras)
	}
}

func TestPrefixStoreRoundTrip(t *testing.T) {
	custom := t.TempDir()
	store := NewPrefixStore(stubSettings{value: types.AppSettings{CustomPrefixDir: custom}})
	cfg := DefaultPrefixOptions("Game")
	cfg.ProtonPath = "GE-Proton9-25"
	if err := store.Save("Game", cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded := store.Load("Game")
	if loaded.ProtonPath != "GE-Proton9-25" {
		t.Errorf("ProtonPath = %q", loaded.ProtonPath)
	}
	if got, want := store.ConfigPath("Game"), filepath.Join(custom, "Game", "light-launcher.json"); got != want {
		t.Errorf("ConfigPath = %q, want %q", got, want)
	}
}

func TestPrefixStoreUnknownVersionUsesDefaults(t *testing.T) {
	custom := t.TempDir()
	target := filepath.Join(custom, "Game", "light-launcher.json")
	writeRaw(t, target, `{"version": 9, "Name": "Old", "ProtonPath": "old"}`)

	cfg := NewPrefixStore(stubSettings{value: types.AppSettings{CustomPrefixDir: custom}}).Load("Game")
	if cfg.Name != "Game" || cfg.ProtonPath != "" {
		t.Errorf("cfg = %+v, want defaults", cfg)
	}
}
