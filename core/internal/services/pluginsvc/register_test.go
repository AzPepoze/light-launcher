package pluginsvc

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/plugins"
	"light-launcher/core/internal/plugins/updater"
	"light-launcher/core/internal/uploads"
)

type silentEvents struct{}

func (silentEvents) Emit(string, any) {}

func newRouter(t *testing.T) *api.Router {
	t.Helper()
	base := t.TempDir()
	host, err := plugins.NewHost(plugins.HostOptions{
		Layout:      plugins.Layout{Root: filepath.Join(base, "plugins"), Data: filepath.Join(base, "data")},
		HostVersion: "2.0.0",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(host.Shutdown)
	store, err := uploads.Open(filepath.Join(base, "uploads"))
	if err != nil {
		t.Fatal(err)
	}
	router := api.NewRouter()
	Register(router, silentEvents{}, host, store)
	return router
}

func TestRegisterAddsEveryPluginMethod(t *testing.T) {
	methods := newRouter(t).Methods()
	sort.Strings(methods)
	want := []string{
		"PluginCall", "PluginInstall", "PluginLogs", "PluginPermissionsPreview", "PluginRemove",
		"PluginRollback", "PluginSetEnabled", "PluginSettingsGet", "PluginSettingsSet", "PluginUpdate",
		"PluginUpdatesCheck", "PluginUpdatesStatus", "PluginsList",
	}
	if len(methods) != len(want) {
		t.Fatalf("methods = %v, want %v", methods, want)
	}
	for i := range want {
		if methods[i] != want[i] {
			t.Fatalf("methods = %v, want %v", methods, want)
		}
	}
}

func TestPluginsListIsEmptyOnFreshInstall(t *testing.T) {
	router := newRouter(t)
	out, err := router.Dispatch(context.Background(), "PluginsList", nil)
	if err != nil {
		t.Fatalf("PluginsList: %v", err)
	}
	if list, ok := out.([]plugins.PluginInfo); !ok || len(list) != 0 {
		t.Fatalf("PluginsList = %#v", out)
	}
}

func TestPluginCallRefusesHostServices(t *testing.T) {
	router := newRouter(t)
	payload := json.RawMessage(`{"plugin":"demo","method":"http.request","params":{}}`)
	_, err := router.Dispatch(context.Background(), "PluginCall", payload)
	if err == nil {
		t.Fatalf("PluginCall must not pass non-source methods to plugins")
	}
}

func TestRemoveUnknownPluginFails(t *testing.T) {
	router := newRouter(t)
	_, err := router.Dispatch(context.Background(), "PluginRemove", json.RawMessage(`{"id":"missing"}`))
	if !errors.Is(err, plugins.ErrNotInstalled) {
		t.Fatalf("err = %v, want ErrNotInstalled", err)
	}
}

func TestBadPayloadIsRejected(t *testing.T) {
	router := newRouter(t)
	_, err := router.Dispatch(context.Background(), "PluginSetEnabled", json.RawMessage(`{"id":`))
	if err == nil {
		t.Fatalf("a broken payload must fail")
	}
}

type stubChecker struct {
	result updater.Result
}

func (s stubChecker) Check(context.Context, updater.Request) (updater.Result, error) {
	return s.result, nil
}

// newRouterWithChecker returns a router whose plugin host uses the given GitHub checker.
func newRouterWithChecker(t *testing.T, checker plugins.UpdateChecker) (*api.Router, plugins.Layout) {
	t.Helper()
	base := t.TempDir()
	layout := plugins.Layout{Root: filepath.Join(base, "plugins"), Data: filepath.Join(base, "data")}
	host, err := plugins.NewHost(plugins.HostOptions{
		Layout:        layout,
		HostVersion:   "2.0.0",
		UpdateChecker: checker,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(host.Shutdown)
	store, err := uploads.Open(filepath.Join(base, "uploads"))
	if err != nil {
		t.Fatal(err)
	}
	router := api.NewRouter()
	Register(router, silentEvents{}, host, store)
	return router, layout
}

func TestPluginUpdatesCheckAndStatus(t *testing.T) {
	router, layout := newRouterWithChecker(t, stubChecker{result: updater.Result{
		Available: true,
		Version:   "2.0.0",
		AssetURL:  "https://github.test/demo.llplugin",
	}})
	pluginDir := layout.PluginDir("demo")
	if err := os.MkdirAll(filepath.Join(pluginDir, "1.0.0"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{
		"id": "demo", "name": "Demo", "version": "1.0.0", "apiVersion": "1", "minHost": "2.0.0",
		"entry": { "linux-x64": "bin/demo" }, "repository": "github:ada/demo",
		"contributes": { "sources": [] }, "permissions": { "network": [], "fs": [] }
	}`
	if err := os.WriteFile(filepath.Join(pluginDir, "1.0.0", "plugin.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "current"), []byte("1.0.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	out, err := router.Dispatch(ctx, "PluginUpdatesCheck", nil)
	if err != nil {
		t.Fatalf("PluginUpdatesCheck: %v", err)
	}
	status, ok := out.(plugins.PluginUpdatesStatus)
	if !ok || status.CheckedAt == "" || len(status.Plugins) != 1 {
		t.Fatalf("PluginUpdatesCheck = %#v", out)
	}
	if !status.Plugins[0].Available || status.Plugins[0].Version != "2.0.0" {
		t.Fatalf("update entry = %+v", status.Plugins[0])
	}

	out, err = router.Dispatch(ctx, "PluginUpdatesStatus", nil)
	if err != nil {
		t.Fatalf("PluginUpdatesStatus: %v", err)
	}
	cached, ok := out.(plugins.PluginUpdatesStatus)
	if !ok || len(cached.Plugins) != 1 || cached.Plugins[0].AssetURL != "https://github.test/demo.llplugin" {
		t.Fatalf("PluginUpdatesStatus = %#v", out)
	}
}

func TestPluginUpdatesStatusIsAnEmptyListOnFreshInstall(t *testing.T) {
	router := newRouter(t)
	out, err := router.Dispatch(context.Background(), "PluginUpdatesStatus", nil)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `{"plugins":[]}` {
		t.Fatalf("PluginUpdatesStatus = %s", encoded)
	}
	if _, err := router.Dispatch(context.Background(), "PluginUpdatesCheck", json.RawMessage(`{"id":"missing"}`)); !errors.Is(err, plugins.ErrNotInstalled) {
		t.Fatalf("check of a missing plugin: err = %v", err)
	}
}
