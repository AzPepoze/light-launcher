package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"light-launcher/core/internal/plugins/updater"
)

type fakeUpdateChecker struct {
	result   updater.Result
	err      error
	requests []updater.Request
}

func (f *fakeUpdateChecker) Check(_ context.Context, req updater.Request) (updater.Result, error) {
	f.requests = append(f.requests, req)
	return f.result, f.err
}

// installFixtureWithRepository installs the fixture plugin with a repository in its manifest.
func installFixtureWithRepository(t *testing.T, host *Host, version, repository string) {
	t.Helper()
	manifest := fixtureManifest(Permissions{Network: []string{"127.0.0.1"}})
	manifest.Version = version
	manifest.Repository = repository
	source := filepath.Join(t.TempDir(), "fixture.llplugin")
	writeFixtureManifest(t, source, manifest)
	if _, err := host.Install(context.Background(), InstallRequest{
		Source: source, AllowUnsigned: true, AcceptedPermissions: testConsent,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestHostCheckUpdatesCachesResults(t *testing.T) {
	checker := &fakeUpdateChecker{result: updater.Result{
		Available:   true,
		Version:     "0.2.0",
		AssetURL:    "https://github.test/demo.llplugin",
		SHA256:      strings.Repeat("a", 64),
		Notes:       "Fixes a bug.",
		PublishedAt: "2026-03-01T10:00:00Z",
	}}
	host := newTestHostWithChecker(t, checker)
	installFixtureWithRepository(t, host, "0.1.0", "github:ada/fixture")
	ctx := context.Background()

	status, err := host.CheckUpdates(ctx, "")
	if err != nil {
		t.Fatalf("CheckUpdates: %v", err)
	}
	if status.CheckedAt == "" || len(status.Plugins) != 1 {
		t.Fatalf("status = %+v", status)
	}
	entry := status.Plugins[0]
	if entry.ID != "fixture" || entry.InstalledVersion != "0.1.0" || !entry.Available || entry.Version != "0.2.0" {
		t.Fatalf("entry = %+v", entry)
	}
	if entry.Notes != "Fixes a bug." || entry.SHA256 != strings.Repeat("a", 64) {
		t.Fatalf("release info = %+v", entry)
	}
	if len(checker.requests) != 1 || checker.requests[0].Repository != "github:ada/fixture" || checker.requests[0].Version != "0.1.0" {
		t.Fatalf("checker requests = %+v", checker.requests)
	}
	if cached := host.UpdatesStatus(); cached.CheckedAt != status.CheckedAt || len(cached.Plugins) != 1 {
		t.Fatalf("cached status = %+v", cached)
	}
}

func TestHostCheckUpdatesSkipsPluginsWithoutRepository(t *testing.T) {
	checker := &fakeUpdateChecker{}
	host := newTestHostWithChecker(t, checker)
	installFixtureWithRepository(t, host, "0.1.0", "")

	status, err := host.CheckUpdates(context.Background(), "fixture")
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Plugins) != 1 || status.Plugins[0].Available || status.Plugins[0].Error != "" {
		t.Fatalf("status = %+v", status)
	}
	if len(checker.requests) != 0 {
		t.Fatalf("checker was called for a plugin without a repository")
	}
}

func TestHostUpdatesFromCheckedRelease(t *testing.T) {
	host := newTestHostWithChecker(t, &fakeUpdateChecker{})
	installFixtureWithRepository(t, host, "0.1.0", "github:ada/fixture")
	ctx := context.Background()

	source := filepath.Join(t.TempDir(), "fixture-0.2.0.llplugin")
	manifest := fixtureManifest(Permissions{Network: []string{"127.0.0.1"}})
	manifest.Version = "0.2.0"
	manifest.Repository = "github:ada/fixture"
	writeFixtureManifest(t, source, manifest)
	digest, err := fileSHA256(source)
	if err != nil {
		t.Fatal(err)
	}
	checker := &fakeUpdateChecker{result: updater.Result{Available: true, Version: "0.2.0", AssetURL: source, SHA256: digest}}
	host.updates.checker = checker

	if _, err := host.CheckUpdates(ctx, "fixture"); err != nil {
		t.Fatal(err)
	}
	info, err := host.Update(ctx, UpdateRequest{ID: "fixture", Source: source, AllowUnsigned: true, AcceptedPermissions: testConsent})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if info.Version != "0.2.0" {
		t.Fatalf("version = %q", info.Version)
	}
	if status := host.UpdatesStatus(); len(status.Plugins) != 0 {
		t.Fatalf("stale update status after update: %+v", status)
	}
}

func TestHostUpdateRefusesSourceWithoutCheck(t *testing.T) {
	host := newTestHostWithChecker(t, &fakeUpdateChecker{})
	installFixtureWithRepository(t, host, "0.1.0", "github:ada/fixture")

	source := filepath.Join(t.TempDir(), "fixture-0.2.0.llplugin")
	manifest := fixtureManifest(Permissions{Network: []string{"127.0.0.1"}})
	manifest.Version = "0.2.0"
	manifest.Repository = "github:ada/fixture"
	writeFixtureManifest(t, source, manifest)

	if _, err := host.Update(context.Background(), UpdateRequest{ID: "fixture", Source: source, AllowUnsigned: true, AcceptedPermissions: testConsent}); !errors.Is(err, ErrUpdateNotAvailable) {
		t.Fatalf("err = %v, want ErrUpdateNotAvailable", err)
	}
	if version := host.info("fixture").Version; version != "0.1.0" {
		t.Fatalf("version = %q after a refused update", version)
	}
}

func TestHostUpdateRejectsBadReleaseChecksum(t *testing.T) {
	host := newTestHostWithChecker(t, &fakeUpdateChecker{})
	installFixtureWithRepository(t, host, "0.1.0", "github:ada/fixture")
	ctx := context.Background()

	source := filepath.Join(t.TempDir(), "fixture-0.2.0.llplugin")
	manifest := fixtureManifest(Permissions{Network: []string{"127.0.0.1"}})
	manifest.Version = "0.2.0"
	manifest.Repository = "github:ada/fixture"
	writeFixtureManifest(t, source, manifest)
	checker := &fakeUpdateChecker{result: updater.Result{Available: true, Version: "0.2.0", AssetURL: source, SHA256: strings.Repeat("0", 64)}}
	host.updates.checker = checker
	if _, err := host.CheckUpdates(ctx, "fixture"); err != nil {
		t.Fatal(err)
	}

	_, err := host.Update(ctx, UpdateRequest{ID: "fixture", Source: source, AllowUnsigned: true, AcceptedPermissions: testConsent})
	if err == nil || !strings.Contains(err.Error(), "release SHA256SUMS") {
		t.Fatalf("err = %v, want a checksum error", err)
	}
	if version := host.info("fixture").Version; version != "0.1.0" {
		t.Fatalf("version = %q after a bad checksum", version)
	}
}

func TestHostPreviewFlagsRepositoryChange(t *testing.T) {
	host := newTestHostWithChecker(t, &fakeUpdateChecker{})
	installFixtureWithRepository(t, host, "0.1.0", "github:ada/fixture")
	ctx := context.Background()

	next := fixtureManifest(Permissions{Network: []string{"127.0.0.1"}})
	next.Version = "0.2.0"
	next.Repository = "github:mallory/fixture"
	source := filepath.Join(t.TempDir(), "fixture-0.2.0.llplugin")
	writeFixtureManifest(t, source, next)

	preview, err := host.PermissionsPreview(ctx, source, "")
	if err != nil {
		t.Fatal(err)
	}
	if !preview.RepositoryChanged || preview.InstalledRepository != "github:ada/fixture" {
		t.Fatalf("preview = %+v, want a repository change warning", preview)
	}

	next.Repository = "github:ada/fixture"
	writeFixtureManifest(t, source, next)
	preview, err = host.PermissionsPreview(ctx, source, "")
	if err != nil {
		t.Fatal(err)
	}
	if preview.RepositoryChanged {
		t.Fatalf("preview = %+v, want no warning for the same repository", preview)
	}
}

func TestHostUpdateRollsBackWhenNewVersionCannotStart(t *testing.T) {
	host := newTestHost(t)
	ctx := context.Background()
	source := filepath.Join(t.TempDir(), "fixture.llplugin")
	writeFixturePackage(t, source, "0.1.0")
	if _, err := host.Install(ctx, InstallRequest{Source: source, AllowUnsigned: true, AcceptedPermissions: testConsent}); err != nil {
		t.Fatal(err)
	}

	broken := filepath.Join(t.TempDir(), "broken.llplugin")
	brokenManifest := fixtureManifest(Permissions{Network: []string{"127.0.0.1"}})
	brokenManifest.Version = "0.2.0"
	manifestJSON, err := json.Marshal(brokenManifest)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{
		"plugin.json":                       manifestJSON,
		brokenManifest.Entry[PlatformKey()]: []byte("this is not a program"),
	}
	writePackageFile(t, broken, withChecksums(files))
	if err := host.install.state.update("fixture", func(state *PluginState) { state.Source = broken }); err != nil {
		t.Fatal(err)
	}

	_, err = host.Update(ctx, UpdateRequest{ID: "fixture", AllowUnsigned: true, AcceptedPermissions: testConsent})
	if err == nil || !strings.Contains(err.Error(), "rolled back") {
		t.Fatalf("err = %v, want a rollback error", err)
	}
	if version := host.info("fixture").Version; version != "0.1.0" {
		t.Fatalf("version = %q, want the previous version", version)
	}
	current, err := readPointer(host.layout.PluginDir("fixture"), currentPointer)
	if err != nil {
		t.Fatal(err)
	}
	previous, err := readPointer(host.layout.PluginDir("fixture"), previousPointer)
	if err != nil {
		t.Fatal(err)
	}
	if current != "0.1.0" || previous != "0.2.0" {
		t.Fatalf("current = %q previous = %q, want 0.1.0 and 0.2.0", current, previous)
	}
}

func TestHostCheckUpdatesRecordsErrors(t *testing.T) {
	checker := &fakeUpdateChecker{err: updater.ErrRateLimited}
	host := newTestHostWithChecker(t, checker)
	installFixtureWithRepository(t, host, "0.1.0", "github:ada/fixture")

	status, err := host.CheckUpdates(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Plugins) != 1 || !strings.Contains(status.Plugins[0].Error, "rate limit") {
		t.Fatalf("status = %+v", status)
	}
}
