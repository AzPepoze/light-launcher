package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// writeDevFolder writes a source folder that holds the fixture plugin at this version, with no checksums.
func writeDevFolder(t *testing.T, dir, version string) {
	t.Helper()
	manifest := fixtureManifest(Permissions{Network: []string{"127.0.0.1"}})
	manifest.Version = version
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plugin.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	binary, err := os.ReadFile(fixtureBinary)
	if err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(dir, filepath.FromSlash(manifest.Entry[PlatformKey()]))
	if err := os.MkdirAll(filepath.Dir(entry), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entry, binary, 0o755); err != nil {
		t.Fatal(err)
	}
}

// installFixture installs the fixture plugin at version from an unsigned package.
func installFixture(t *testing.T, host *Host, version string) {
	t.Helper()
	source := filepath.Join(t.TempDir(), "fixture.llplugin")
	writeFixturePackage(t, source, version)
	if _, err := host.Install(context.Background(), InstallRequest{Source: source, AllowUnsigned: true, AcceptedPermissions: testConsent}); err != nil {
		t.Fatalf("Install: %v", err)
	}
}

func TestDevLinkRefusesFolderWithoutManifest(t *testing.T) {
	host := newTestHost(t)
	_, err := host.DevLink(t.TempDir())
	if !errors.Is(err, ErrNoManifest) {
		t.Fatalf("DevLink without plugin.json: err = %v", err)
	}
	if host.install.state.get("fixture").Dev {
		t.Fatalf("a refused link must not be recorded")
	}
}

func TestDevLinkRefusesInvalidManifest(t *testing.T) {
	host := newTestHost(t)
	dir := t.TempDir()
	manifest := fixtureManifest(Permissions{})
	manifest.MinHost = "9.0.0"
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plugin.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := host.DevLink(dir); err == nil {
		t.Fatalf("DevLink must refuse a manifest that needs a newer launcher")
	}
	if len(host.List()) != 0 {
		t.Fatalf("a refused link must not list a plugin")
	}
}

func TestDevLinkRefusesFile(t *testing.T) {
	host := newTestHost(t)
	file := filepath.Join(t.TempDir(), "plugin.json")
	if err := os.WriteFile(file, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := host.DevLink(file); !errors.Is(err, ErrNotFolder) {
		t.Fatalf("DevLink on a file: err = %v", err)
	}
}

func TestDevLinkRunsFromSourceFolder(t *testing.T) {
	host := newTestHost(t)
	installFixture(t, host, "0.1.0")
	dir := filepath.Join(t.TempDir(), "fixture-src")
	writeDevFolder(t, dir, "0.2.0")

	info, err := host.DevLink(dir)
	if err != nil {
		t.Fatalf("DevLink: %v", err)
	}
	if !info.Dev || info.DevPath != dir || info.Version != "0.2.0" {
		t.Fatalf("linked info = %+v", info)
	}
	if _, err := host.CallPlugin(context.Background(), "fixture", "source.parseRef", json.RawMessage(`{"url":"x"}`)); err != nil {
		t.Fatalf("call on linked plugin: %v", err)
	}
	if !host.info("fixture").Running {
		t.Fatalf("linked plugin should be running after a call")
	}
}

func TestDevUnlinkReturnsToInstalledCopy(t *testing.T) {
	host := newTestHost(t)
	installFixture(t, host, "0.1.0")
	dir := filepath.Join(t.TempDir(), "fixture-src")
	writeDevFolder(t, dir, "0.2.0")
	if _, err := host.DevLink(dir); err != nil {
		t.Fatal(err)
	}

	if err := host.DevUnlink("fixture"); err != nil {
		t.Fatalf("DevUnlink: %v", err)
	}
	info := host.info("fixture")
	if info.Dev || info.DevPath != "" || info.Version != "0.1.0" {
		t.Fatalf("info after unlink = %+v", info)
	}
	if host.running("fixture") {
		t.Fatalf("unlinking must stop the linked plugin")
	}
}

func TestDevOnlyPluginIsListedAndRemovedByUnlink(t *testing.T) {
	host := newTestHost(t)
	dir := filepath.Join(t.TempDir(), "fixture-src")
	writeDevFolder(t, dir, "0.2.0")
	if _, err := host.DevLink(dir); err != nil {
		t.Fatal(err)
	}
	if infos := host.List(); len(infos) != 1 || infos[0].ID != "fixture" || !infos[0].Dev {
		t.Fatalf("List after link = %+v", infos)
	}

	req := InstallRequest{Source: dir, AllowUnsigned: true, AcceptedPermissions: testConsent}
	if _, err := host.Install(context.Background(), req); !errors.Is(err, ErrDevLinked) {
		t.Fatalf("Install over a linked id: err = %v", err)
	}
	if err := host.DevUnlink("fixture"); err != nil {
		t.Fatalf("DevUnlink: %v", err)
	}
	if infos := host.List(); len(infos) != 0 {
		t.Fatalf("List after unlink = %+v", infos)
	}
}

func TestDevUnlinkRefusesPluginThatIsNotLinked(t *testing.T) {
	host := newTestHost(t)
	installFixture(t, host, "0.1.0")
	if err := host.DevUnlink("fixture"); !errors.Is(err, ErrNotDevLinked) {
		t.Fatalf("DevUnlink on an installed plugin: err = %v", err)
	}
}
