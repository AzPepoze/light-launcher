package main

import (
	"testing"
	"time"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/plugins"
)

func TestSyncCapable(t *testing.T) {
	cases := []struct {
		name   string
		source contract.SourceInfo
		want   bool
	}{
		{"games", contract.SourceInfo{Capabilities: []string{"browse", "sync:games"}}, true},
		{"index", contract.SourceInfo{Capabilities: []string{"sync:index"}}, true},
		{"none", contract.SourceInfo{Capabilities: []string{"browse", "updateCheck"}}, false},
	}
	for _, tc := range cases {
		if got := syncCapable(tc.source); got != tc.want {
			t.Errorf("%s: syncCapable = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestSyncIntervalFromSettings(t *testing.T) {
	withSchema := func(value any) plugins.SettingsView {
		return plugins.SettingsView{
			Schema: map[string]any{"properties": map[string]any{"syncIntervalHours": map[string]any{"type": "integer"}}},
			Values: map[string]any{"syncIntervalHours": value},
		}
	}
	cases := []struct {
		name string
		view plugins.SettingsView
		want time.Duration
	}{
		{"saved hours", withSchema(float64(6)), 6 * time.Hour},
		{"saved zero", withSchema(float64(0)), syncIntervalDefault},
		{"wrong type", withSchema("6"), syncIntervalDefault},
		{
			name: "schema has no setting",
			view: plugins.SettingsView{
				Schema: map[string]any{"properties": map[string]any{"other": map[string]any{}}},
				Values: map[string]any{"syncIntervalHours": float64(6)},
			},
			want: syncIntervalDefault,
		},
		{"no schema", plugins.SettingsView{}, syncIntervalDefault},
	}
	for _, tc := range cases {
		if got := syncIntervalFromSettings(tc.view); got != tc.want {
			t.Errorf("%s: interval = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestPluginRunnable(t *testing.T) {
	cases := []struct {
		name string
		info plugins.PluginInfo
		want bool
	}{
		{"stopped", plugins.PluginInfo{ID: "f95", Enabled: true, Status: "stopped"}, true},
		{"running", plugins.PluginInfo{ID: "f95", Enabled: true, Status: "running"}, true},
		{"disabled", plugins.PluginInfo{ID: "f95", Enabled: false, Status: "disabled"}, false},
		{"broken", plugins.PluginInfo{ID: "f95", Enabled: true, Status: "broken"}, false},
		{"missing", plugins.PluginInfo{}, false},
	}
	for _, tc := range cases {
		if got := pluginRunnable(tc.info); got != tc.want {
			t.Errorf("%s: pluginRunnable = %v, want %v", tc.name, got, tc.want)
		}
	}
}
