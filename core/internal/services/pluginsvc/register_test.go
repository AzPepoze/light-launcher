package pluginsvc

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"sort"
	"testing"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/plugins"
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
	router := api.NewRouter()
	Register(router, silentEvents{}, host)
	return router
}

func TestRegisterAddsEveryPluginMethod(t *testing.T) {
	methods := newRouter(t).Methods()
	sort.Strings(methods)
	want := []string{
		"PluginCall", "PluginInstall", "PluginLogs", "PluginPermissionsPreview", "PluginRemove",
		"PluginRollback", "PluginSetEnabled", "PluginSettingsGet", "PluginSettingsSet", "PluginUpdate",
		"PluginsList",
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
