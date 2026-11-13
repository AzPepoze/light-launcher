package plugins

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// initializeSettingsMarker matches the marker file the fixture plugin checks in its data folder.
const initializeSettingsMarker = "initialize-probes-settings"

// TestHostServicesAnswerDuringInitialize covers the startup path F95 uses: its initialize handler
// reads settings before it answers. The host must serve that request while the handshake runs,
// or the plugin waits for its own timeout and starts with defaults.
func TestHostServicesAnswerDuringInitialize(t *testing.T) {
	sup, env, _ := newFixtureSupervisor(t, Permissions{})
	settings, err := json.Marshal(map[string]string{"region": "eu"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(env.dataDir, settingsFileName), settings, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(env.dataDir, initializeSettingsMarker), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 8*time.Second)
	defer cancel()
	if _, err := sup.connect(ctx); err != nil {
		t.Fatalf("handshake: %v", err)
	}
	raw, err := callFixture(sup, "source.initializeProbe", nil)
	if err != nil {
		t.Fatalf("source.initializeProbe: %v", err)
	}
	var probe struct {
		Received bool           `json:"received"`
		Settings map[string]any `json:"settings"`
		Error    string         `json:"error"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		t.Fatalf("probe %s: %v", raw, err)
	}
	if !probe.Received {
		t.Fatalf("settings.get during initialize was not answered (error %q)", probe.Error)
	}
	if probe.Settings["region"] != "eu" {
		t.Fatalf("settings during initialize = %v, want region eu", probe.Settings)
	}
}
