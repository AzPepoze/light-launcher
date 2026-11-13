package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"light-launcher/core/internal/contract"
)

func TestMarkSourceCollisionsKeepsTheFirstOwner(t *testing.T) {
	infos := []PluginInfo{
		{ID: "alpha", Enabled: true, Status: "running", Sources: []contract.SourceInfo{{SourceID: "shared"}}},
		{ID: "beta", Enabled: true, Status: "running", Sources: []contract.SourceInfo{{SourceID: "shared"}, {SourceID: "own"}}},
	}
	got := markSourceCollisions(infos)
	if got[0].Status == "broken" {
		t.Fatalf("first plugin was marked broken: %+v", got[0])
	}
	if got[1].Status != "broken" || !strings.Contains(got[1].Error, "already provided by plugin alpha") {
		t.Fatalf("second plugin = %+v, want broken with the owner in the error", got[1])
	}
}

func TestHostInstallOrUpdateInstallsNewPlugin(t *testing.T) {
	host := newTestHost(t)
	source := filepath.Join(t.TempDir(), "fixture.llplugin")
	writeFixturePackage(t, source, "0.1.0")

	info, err := host.InstallOrUpdate(context.Background(), InstallRequest{
		Source: source, AllowUnsigned: true, AcceptedPermissions: testConsent,
	})
	if err != nil {
		t.Fatalf("InstallOrUpdate: %v", err)
	}
	if info.ID != "fixture" || info.Version != "0.1.0" {
		t.Fatalf("installed info = %+v", info)
	}
	if version := host.info("fixture").Version; version != "0.1.0" {
		t.Fatalf("installed version = %q", version)
	}
}

func TestHostInstallOrUpdateUpdatesInPlaceAndKeepsData(t *testing.T) {
	store := newFakeSecrets(nil)
	host := newTestHostWithSecrets(t, store)
	ctx := context.Background()
	source := filepath.Join(t.TempDir(), "fixture.llplugin")
	writeFixturePackage(t, source, "0.1.0")
	if _, err := host.Install(ctx, InstallRequest{Source: source, AllowUnsigned: true, AcceptedPermissions: testConsent}); err != nil {
		t.Fatal(err)
	}
	dataFile := filepath.Join(host.layout.DataDir("fixture"), "save.json")
	if err := os.MkdirAll(filepath.Dir(dataFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dataFile, []byte(`{"level":3}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := host.deps.secrets.Set(ctx, "fixture", "token", "abc"); err != nil {
		t.Fatal(err)
	}

	writeFixturePackage(t, source, "0.2.0")
	info, err := host.InstallOrUpdate(ctx, InstallRequest{
		Source: source, AllowUnsigned: true, AcceptedPermissions: testConsent,
	})
	if err != nil {
		t.Fatalf("InstallOrUpdate: %v", err)
	}
	if info.Version != "0.2.0" || info.Previous != "0.1.0" {
		t.Fatalf("updated info = %+v", info)
	}
	data, err := os.ReadFile(dataFile)
	if err != nil || string(data) != `{"level":3}` {
		t.Fatalf("data after update = %q, %v", data, err)
	}
	value, found, err := host.deps.secrets.Get(ctx, "fixture", "token")
	if err != nil || !found || value != "abc" {
		t.Fatalf("secret after update = %q, found %v, err %v", value, found, err)
	}
}

func TestHostInstallOrUpdateRefusesSameVersion(t *testing.T) {
	host := newTestHost(t)
	ctx := context.Background()
	source := filepath.Join(t.TempDir(), "fixture.llplugin")
	writeFixturePackage(t, source, "0.1.0")
	if _, err := host.Install(ctx, InstallRequest{Source: source, AllowUnsigned: true, AcceptedPermissions: testConsent}); err != nil {
		t.Fatal(err)
	}

	_, err := host.InstallOrUpdate(ctx, InstallRequest{
		Source: source, AllowUnsigned: true, AcceptedPermissions: testConsent,
	})
	if !errors.Is(err, ErrNoUpdate) {
		t.Fatalf("err = %v, want ErrNoUpdate", err)
	}
	if !strings.Contains(err.Error(), "fixture 0.1.0") {
		t.Fatalf("err = %v, want the plugin id and version in the message", err)
	}
	if version := host.info("fixture").Version; version != "0.1.0" {
		t.Fatalf("version after a refused same-version package = %q", version)
	}
	if previous, err := readPointer(host.layout.PluginDir("fixture"), previousPointer); err != nil || previous != "" {
		t.Fatalf("same-version package changed the previous pointer: %q, %v", previous, err)
	}
}

func TestHostInstallOrUpdateKeepsURLUpdatesStrict(t *testing.T) {
	host := newTestHost(t)
	ctx := context.Background()
	source := filepath.Join(t.TempDir(), "fixture.llplugin")
	writeFixturePackage(t, source, "0.1.0")
	if _, err := host.Install(ctx, InstallRequest{Source: source, AllowUnsigned: true, AcceptedPermissions: testConsent}); err != nil {
		t.Fatal(err)
	}

	packaged := filepath.Join(t.TempDir(), "fixture-0.2.0.llplugin")
	writeFixturePackage(t, packaged, "0.2.0")
	body, err := os.ReadFile(packaged)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	defer server.Close()
	host.install.client = server.Client()
	link := server.URL + "/fixture-0.2.0.llplugin"
	if _, err := host.PermissionsPreview(ctx, link, ""); err != nil {
		t.Fatalf("PermissionsPreview: %v", err)
	}

	if _, err := host.InstallOrUpdate(ctx, InstallRequest{
		Source: link, AllowUnsigned: true, AcceptedPermissions: testConsent,
	}); !errors.Is(err, ErrUpdateNotAvailable) {
		t.Fatalf("err = %v, want ErrUpdateNotAvailable for an unchecked https source", err)
	}
	if version := host.info("fixture").Version; version != "0.1.0" {
		t.Fatalf("version after a refused link update = %q", version)
	}
}

func TestHostInstallOrUpdateRollsBackFailedUpdate(t *testing.T) {
	store := newFakeSecrets(nil)
	host := newTestHostWithSecrets(t, store)
	ctx := context.Background()
	source := filepath.Join(t.TempDir(), "fixture.llplugin")
	writeFixturePackage(t, source, "0.1.0")
	if _, err := host.Install(ctx, InstallRequest{Source: source, AllowUnsigned: true, AcceptedPermissions: testConsent}); err != nil {
		t.Fatal(err)
	}
	dataFile := filepath.Join(host.layout.DataDir("fixture"), "save.json")
	if err := os.MkdirAll(filepath.Dir(dataFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dataFile, []byte(`{"level":3}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := host.deps.secrets.Set(ctx, "fixture", "token", "abc"); err != nil {
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

	_, err = host.InstallOrUpdate(ctx, InstallRequest{
		Source: broken, AllowUnsigned: true, AcceptedPermissions: testConsent,
	})
	if err == nil || !strings.Contains(err.Error(), "rolled back") {
		t.Fatalf("err = %v, want a rollback error", err)
	}
	if version := host.info("fixture").Version; version != "0.1.0" {
		t.Fatalf("version after a failed update = %q, want 0.1.0", version)
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
	data, err := os.ReadFile(dataFile)
	if err != nil || string(data) != `{"level":3}` {
		t.Fatalf("data after a failed update = %q, %v", data, err)
	}
	value, found, err := host.deps.secrets.Get(ctx, "fixture", "token")
	if err != nil || !found || value != "abc" {
		t.Fatalf("secret after a failed update = %q, found %v, err %v", value, found, err)
	}
}
