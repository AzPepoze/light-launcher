package main

import (
	"context"
	"errors"
	"slices"
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
		{"list only", contract.SourceInfo{Capabilities: []string{"browse", "listGames"}}, true},
		{"none", contract.SourceInfo{Capabilities: []string{"browse", "updateCheck"}}, false},
	}
	for _, tc := range cases {
		if got := syncCapable(tc.source); got != tc.want {
			t.Errorf("%s: syncCapable = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// fakeSources records the source.sync calls a scheduled job makes.
type fakeSources struct {
	calls []string
	err   error
}

func (f *fakeSources) Sources() []contract.SourceInfo { return nil }

func (f *fakeSources) Call(_ context.Context, plugin, method string, _ any, _ any) error {
	f.calls = append(f.calls, plugin+" "+method)
	return f.err
}

func (f *fakeSources) StartDownload(context.Context, string, any, int64) (int64, error) {
	return 0, errors.New("not used")
}

// fakeImporter records the sources a scheduled job imports.
type fakeImporter struct {
	calls []string
	err   error
}

func (f *fakeImporter) importSource(_ context.Context, plugin, source string) error {
	f.calls = append(f.calls, plugin+" "+source)
	return f.err
}

func TestRunSourceSyncRoutesByCapability(t *testing.T) {
	cases := []struct {
		name        string
		source      contract.SourceInfo
		wantImports []string
		wantCalls   []string
	}{
		{
			name:        "listGames imports",
			source:      contract.SourceInfo{SourceID: "mine", Capabilities: []string{"listGames"}},
			wantImports: []string{"demo mine"},
		},
		{
			name:        "listGames wins over sync:games",
			source:      contract.SourceInfo{SourceID: "mine", Capabilities: []string{"sync:games", "listGames"}},
			wantImports: []string{"demo mine"},
		},
		{
			name:      "sync:games calls source.sync",
			source:    contract.SourceInfo{SourceID: "shop", Capabilities: []string{"sync:games"}},
			wantCalls: []string{"demo source.sync"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sources := &fakeSources{}
			importer := &fakeImporter{}
			err := runSourceSync(t.Context(), sources, importer.importSource, "demo", tc.source)
			if err != nil {
				t.Fatalf("runSourceSync: %v", err)
			}
			if !slices.Equal(importer.calls, tc.wantImports) {
				t.Errorf("imports = %v, want %v", importer.calls, tc.wantImports)
			}
			if !slices.Equal(sources.calls, tc.wantCalls) {
				t.Errorf("source calls = %v, want %v", sources.calls, tc.wantCalls)
			}
		})
	}
}

func TestRunSourceSyncReportsFailures(t *testing.T) {
	boom := errors.New("boom")
	listed := contract.SourceInfo{SourceID: "mine", Capabilities: []string{"listGames"}}
	browse := contract.SourceInfo{SourceID: "shop", Capabilities: []string{"sync:games"}}

	importer := &fakeImporter{err: boom}
	if err := runSourceSync(t.Context(), &fakeSources{}, importer.importSource, "demo", listed); !errors.Is(err, boom) {
		t.Errorf("import failure = %v, want it to wrap %v", err, boom)
	}
	sources := &fakeSources{err: boom}
	if err := runSourceSync(t.Context(), sources, nil, "demo", browse); !errors.Is(err, boom) {
		t.Errorf("source.sync failure = %v, want it to wrap %v", err, boom)
	}
	if err := runSourceSync(t.Context(), &fakeSources{}, nil, "demo", listed); !errors.Is(err, errImportNotWired) {
		t.Errorf("missing import = %v, want %v", err, errImportNotWired)
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
