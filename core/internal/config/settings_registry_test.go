package config

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/types"
)

func TestDefaultSettingsHaveNoRegistryURL(t *testing.T) {
	if url := DefaultSettings().PluginRegistryURL; url != "" {
		t.Fatalf("PluginRegistryURL = %q, want empty", url)
	}
}

func TestSettingsRegistryURLRoundTrip(t *testing.T) {
	target := filepath.Join(t.TempDir(), "settings.json")
	store := NewSettingsStore(target)
	want := "https://plugins.example/index.json"
	if err := store.Save(types.AppSettings{PluginRegistryURL: want}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded := store.Load()
	if loaded.PluginRegistryURL != want {
		t.Fatalf("PluginRegistryURL = %q, want %q", loaded.PluginRegistryURL, want)
	}
	var raw map[string]any
	if err := readJSON(t, target, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["pluginRegistryURL"] != want {
		t.Errorf("settings JSON pluginRegistryURL = %v, want %q", raw["pluginRegistryURL"], want)
	}
}

func TestSaveAppSettingsRejectsBadRegistryURLs(t *testing.T) {
	for _, raw := range []string{
		"http://plugins.example/index.json",
		"ftp://plugins.example/index.json",
		"plugins.example/index.json",
		"https://",
	} {
		err := validateAppSettings(types.AppSettings{PluginRegistryURL: raw})
		if err == nil {
			t.Fatalf("%q was accepted", raw)
		}
		var fieldsErr *api.FieldsError
		if !errors.As(err, &fieldsErr) || len(fieldsErr.Fields) != 1 || fieldsErr.Fields[0].Field != "pluginRegistryURL" {
			t.Fatalf("error for %q = %v, want a pluginRegistryURL field error", raw, err)
		}
		if !strings.Contains(fieldsErr.Fields[0].Message, "https") {
			t.Errorf("message for %q = %q, want an https hint", raw, fieldsErr.Fields[0].Message)
		}
	}
	for _, raw := range []string{"", "https://plugins.example/index.json"} {
		if err := validateAppSettings(types.AppSettings{PluginRegistryURL: raw}); err != nil {
			t.Fatalf("%q was rejected: %v", raw, err)
		}
	}
}

func TestNormalizeTrimsRegistryURL(t *testing.T) {
	settings := &types.AppSettings{PluginRegistryURL: "  https://plugins.example/index.json  "}
	if !normalizePluginRegistryURL(settings) {
		t.Fatal("surrounding spaces did not report a change")
	}
	if settings.PluginRegistryURL != "https://plugins.example/index.json" {
		t.Fatalf("PluginRegistryURL = %q", settings.PluginRegistryURL)
	}
	if normalizePluginRegistryURL(settings) {
		t.Error("an already trimmed URL reported a change")
	}
}
