package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// newTestHost returns a host whose files live in a temp folder and whose restarts are fast.
func newTestHost(t *testing.T) *Host {
	t.Helper()
	base := t.TempDir()
	layout := Layout{Root: filepath.Join(base, "plugins"), Data: filepath.Join(base, "plugin-data")}
	host, err := NewHost(HostOptions{
		Layout:      layout,
		HostVersion: "2.0.0",
		Secrets:     newFileSecrets(filepath.Join(base, "secrets")),
	})
	if err != nil {
		t.Fatal(err)
	}
	host.tuning.backoff = 10 * time.Millisecond
	t.Cleanup(host.Shutdown)
	return host
}

// writeFixturePackage writes an unsigned, checksummed package that holds the fixture plugin at a version.
func writeFixturePackage(t *testing.T, path, version string) {
	t.Helper()
	manifest := fixtureManifest(Permissions{Network: []string{"127.0.0.1"}})
	manifest.Version = version
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.ReadFile(fixtureBinary)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{
		"plugin.json":                 manifestJSON,
		manifest.Entry[PlatformKey()]: binary,
	}
	writePackageFile(t, path, withChecksums(files))
}

func TestHostInstallRunDisableAndRemove(t *testing.T) {
	host := newTestHost(t)
	ctx := context.Background()
	source := filepath.Join(t.TempDir(), "fixture.llplugin")
	writeFixturePackage(t, source, "0.1.0")

	if _, err := host.Install(ctx, InstallRequest{Source: source}); !errors.Is(err, ErrUnsigned) {
		t.Fatalf("unsigned install without consent: err = %v", err)
	}
	info, err := host.Install(ctx, InstallRequest{Source: source, AllowUnsigned: true})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if info.ID != "fixture" || info.Status != "stopped" || !info.Unsigned {
		t.Fatalf("install info = %+v", info)
	}
	if len(host.List()) != 1 {
		t.Fatalf("List = %+v", host.List())
	}

	out, err := host.CallPlugin(ctx, "fixture", "source.parseRef", json.RawMessage(`{"url":"x"}`))
	if err != nil || !json.Valid(out) {
		t.Fatalf("CallPlugin: %s, %v", out, err)
	}
	if status := host.info("fixture").Status; status != "running" {
		t.Fatalf("status after a call = %q", status)
	}
	if _, err := host.CallPlugin(ctx, "fixture", "http.request", nil); err == nil {
		t.Fatalf("the UI must not reach host services through CallPlugin")
	}

	registry := NewRegistry(host)
	if sources := registry.Sources(); len(sources) != 1 || sources[0].SourceID != "fixture" {
		t.Fatalf("Sources = %+v", sources)
	}
	if err := registry.Call(ctx, "fixture", "source.parseRef", map[string]string{}, nil); err != nil {
		t.Fatalf("registry call: %v", err)
	}
	if err := registry.Call(ctx, "fixture", "http.request", nil, nil); err == nil {
		t.Fatalf("registry must refuse non-source methods")
	}

	if _, err := host.SetEnabled("fixture", false); err != nil {
		t.Fatal(err)
	}
	if status := host.info("fixture").Status; status != "disabled" {
		t.Fatalf("status after disable = %q", status)
	}
	if len(registry.Sources()) != 0 {
		t.Fatalf("a disabled plugin must not list sources")
	}
	if _, err := host.CallPlugin(ctx, "fixture", "source.parseRef", nil); !errors.Is(err, ErrPluginDisabled) {
		t.Fatalf("call on disabled plugin: err = %v", err)
	}

	if err := host.Remove("fixture", false); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if len(host.List()) != 0 {
		t.Fatalf("plugin still listed after Remove")
	}
}

func TestHostAutoDisablesAfterCrashes(t *testing.T) {
	host := newTestHost(t)
	ctx := context.Background()
	source := filepath.Join(t.TempDir(), "fixture.llplugin")
	writeFixturePackage(t, source, "0.1.0")
	if _, err := host.Install(ctx, InstallRequest{Source: source, AllowUnsigned: true}); err != nil {
		t.Fatal(err)
	}

	for crash := 1; crash <= 3; crash++ {
		if _, err := host.CallPlugin(ctx, "fixture", "source.parseRef", nil); err != nil {
			t.Fatalf("crash %d: call before crash: %v", crash, err)
		}
		_, _ = host.CallPlugin(ctx, "fixture", "source.crash", nil)
	}
	waitUntil(t, func() bool { return !host.install.state.get("fixture").Enabled })

	info := host.info("fixture")
	if info.Status != "disabled" || info.DisabledReason == "" {
		t.Fatalf("info after crashes = %+v", info)
	}
}

func TestHostUpdateAndRollback(t *testing.T) {
	host := newTestHost(t)
	ctx := context.Background()
	source := filepath.Join(t.TempDir(), "fixture.llplugin")
	writeFixturePackage(t, source, "0.1.0")
	if _, err := host.Install(ctx, InstallRequest{Source: source, AllowUnsigned: true}); err != nil {
		t.Fatal(err)
	}

	writeFixturePackage(t, source, "0.2.0")
	updated, err := host.Update(ctx, "fixture", true)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Version != "0.2.0" || updated.Previous != "0.1.0" {
		t.Fatalf("after update = %+v", updated)
	}

	rolled, err := host.Rollback("fixture")
	if err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if rolled.Version != "0.1.0" || rolled.Previous != "0.2.0" {
		t.Fatalf("after rollback = %+v", rolled)
	}
}

func TestHostRejectsBadIDs(t *testing.T) {
	host := newTestHost(t)
	for _, id := range []string{"../etc", "Upper", "a/b", ""} {
		if _, err := host.SetEnabled(id, false); !errors.Is(err, ErrInvalidID) {
			t.Errorf("SetEnabled(%q) err = %v, want ErrInvalidID", id, err)
		}
	}
}

func TestHostLogsKeepPluginOutput(t *testing.T) {
	host := newTestHost(t)
	ctx := context.Background()
	source := filepath.Join(t.TempDir(), "fixture.llplugin")
	writeFixturePackage(t, source, "0.1.0")
	if _, err := host.Install(ctx, InstallRequest{Source: source, AllowUnsigned: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := host.CallPlugin(ctx, "fixture", "source.echo", json.RawMessage(`{"x":"1"}`)); err != nil {
		t.Fatal(err)
	}
	lines, err := host.Logs("fixture", 0)
	if err != nil || !containsLine(lines, "echo called") {
		t.Fatalf("logs = %v, %v", lines, err)
	}
}
