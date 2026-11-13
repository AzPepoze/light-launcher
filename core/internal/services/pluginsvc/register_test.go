package pluginsvc

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	router, _ := newRouterAndLayout(t)
	return router
}

// newRouterAndLayout returns the router and its plugin layout, so a test can inspect trusted_keys.
func newRouterAndLayout(t *testing.T) (*api.Router, plugins.Layout) {
	t.Helper()
	base := t.TempDir()
	layout := plugins.Layout{Root: filepath.Join(base, "plugins"), Data: filepath.Join(base, "data")}
	host, err := plugins.NewHost(plugins.HostOptions{
		Layout:      layout,
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
	register(router, silentEvents{}, host, store, staticSettings{}, plugins.FetchRegistryIndex)
	return router, layout
}

func TestRegisterAddsEveryPluginMethod(t *testing.T) {
	methods := newRouter(t).Methods()
	sort.Strings(methods)
	want := []string{
		"PluginCall", "PluginInstall", "PluginInstallOrUpdate", "PluginLogs", "PluginPermissionsPreview",
		"PluginRegistry", "PluginRemove", "PluginRollback", "PluginSetEnabled", "PluginSettingsGet",
		"PluginSettingsSet", "PluginTrustedKeyAdd", "PluginTrustedKeyRemove", "PluginTrustedKeys",
		"PluginUpdate", "PluginUpdatesCheck", "PluginUpdatesStatus", "PluginsList",
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

// trustedKeyPayload builds the add/remove payload for one public key.
func trustedKeyPayload(t *testing.T, key ed25519.PublicKey) json.RawMessage {
	t.Helper()
	payload, err := json.Marshal(map[string]string{"key": base64.StdEncoding.EncodeToString(key)})
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

func TestPluginTrustedKeysIsEmptyOnFreshInstall(t *testing.T) {
	router := newRouter(t)
	out, err := router.Dispatch(context.Background(), "PluginTrustedKeys", nil)
	if err != nil {
		t.Fatalf("PluginTrustedKeys: %v", err)
	}
	keys, ok := out.([]plugins.TrustedKey)
	if !ok || len(keys) != 0 {
		t.Fatalf("PluginTrustedKeys = %#v", out)
	}
	encoded, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != "[]" {
		t.Fatalf("PluginTrustedKeys = %s, want []", encoded)
	}
}

func TestPluginTrustedKeyAddListAndRemove(t *testing.T) {
	router, layout := newRouterAndLayout(t)
	public, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(layout.Root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(layout.TrustedKeysPath(), []byte("# keep this comment\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	payload := trustedKeyPayload(t, public)
	encodedKey := base64.StdEncoding.EncodeToString(public)

	out, err := router.Dispatch(ctx, "PluginTrustedKeyAdd", payload)
	if err != nil {
		t.Fatalf("PluginTrustedKeyAdd: %v", err)
	}
	keys, ok := out.([]plugins.TrustedKey)
	if !ok || len(keys) != 1 || keys[0].Base64 != encodedKey || len(keys[0].Fingerprint) != 16 {
		t.Fatalf("PluginTrustedKeyAdd = %#v", out)
	}
	data, err := os.ReadFile(layout.TrustedKeysPath())
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "# keep this comment\n"+encodedKey+"\n" {
		t.Fatalf("trusted_keys = %q, want the comment and the new key", data)
	}
	info, err := os.Stat(layout.TrustedKeysPath())
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o644 {
		t.Fatalf("trusted_keys mode = %o, want 644", perm)
	}

	out, err = router.Dispatch(ctx, "PluginTrustedKeys", nil)
	if err != nil {
		t.Fatalf("PluginTrustedKeys: %v", err)
	}
	if listed, ok := out.([]plugins.TrustedKey); !ok || len(listed) != 1 || listed[0].Base64 != encodedKey {
		t.Fatalf("PluginTrustedKeys = %#v", out)
	}

	if _, err := router.Dispatch(ctx, "PluginTrustedKeyAdd", payload); !errors.Is(err, plugins.ErrTrustedKeyDuplicate) {
		t.Fatalf("duplicate add: err = %v, want ErrTrustedKeyDuplicate", err)
	}

	out, err = router.Dispatch(ctx, "PluginTrustedKeyRemove", payload)
	if err != nil {
		t.Fatalf("PluginTrustedKeyRemove: %v", err)
	}
	if remaining, ok := out.([]plugins.TrustedKey); !ok || len(remaining) != 0 {
		t.Fatalf("PluginTrustedKeyRemove = %#v", out)
	}
	data, err = os.ReadFile(layout.TrustedKeysPath())
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "# keep this comment\n" {
		t.Fatalf("trusted_keys = %q, want the comment kept", data)
	}

	if _, err := router.Dispatch(ctx, "PluginTrustedKeyRemove", payload); !errors.Is(err, plugins.ErrTrustedKeyNotFound) {
		t.Fatalf("removing a missing key: err = %v, want ErrTrustedKeyNotFound", err)
	}
}

func TestPluginTrustedKeyMethodsReportValidationErrors(t *testing.T) {
	router := newRouter(t)
	ctx := context.Background()
	for _, method := range []string{"PluginTrustedKeyAdd", "PluginTrustedKeyRemove"} {
		payload, err := json.Marshal(map[string]string{"key": "not a key!"})
		if err != nil {
			t.Fatal(err)
		}
		_, err = router.Dispatch(ctx, method, payload)
		if !errors.Is(err, plugins.ErrTrustedKeyInvalid) {
			t.Fatalf("%s: err = %v, want ErrTrustedKeyInvalid", method, err)
		}
		if !strings.Contains(err.Error(), "32-byte ed25519 public key") {
			t.Fatalf("%s: err = %q, want the key size in the message", method, err)
		}
	}

	_, err := router.Dispatch(ctx, "PluginTrustedKeyAdd", json.RawMessage(`{}`))
	if !errors.Is(err, plugins.ErrTrustedKeyInvalid) || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("missing key field: err = %v, want an empty key error", err)
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

func TestFailedInstallRemovesItsUploadFolder(t *testing.T) {
	for _, method := range []string{"PluginInstall", "PluginInstallOrUpdate"} {
		t.Run(method, func(t *testing.T) {
			base := t.TempDir()
			layout := plugins.Layout{Root: filepath.Join(base, "plugins"), Data: filepath.Join(base, "data")}
			host, err := plugins.NewHost(plugins.HostOptions{Layout: layout, HostVersion: "2.0.0"})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(host.Shutdown)
			store, err := uploads.Open(filepath.Join(base, "uploads"))
			if err != nil {
				t.Fatal(err)
			}
			router := api.NewRouter()
			register(router, silentEvents{}, host, store, staticSettings{}, plugins.FetchRegistryIndex)

			folder, err := store.NewFolder()
			if err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(folder, "broken")
			if err := os.MkdirAll(source, 0o755); err != nil {
				t.Fatal(err)
			}
			payload, err := json.Marshal(map[string]string{"source": source})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := router.Dispatch(context.Background(), method, payload); err == nil {
				t.Fatalf("an install without plugin.json must fail")
			}
			if _, err := os.Stat(folder); !os.IsNotExist(err) {
				t.Fatalf("upload folder still exists after a failed install: %v", err)
			}
		})
	}
}

// writeLocalPluginPackage writes a folder package that installs without a running plugin.
func writeLocalPluginPackage(t *testing.T, dir, id, version string) string {
	t.Helper()
	source := filepath.Join(dir, id+"-"+version)
	entry := "bin/plugin-" + plugins.PlatformKey()
	manifest := fmt.Sprintf(`{
		"id": %q, "name": "Test", "version": %q, "apiVersion": "1", "minHost": "2.0.0",
		"entry": { %q: %q },
		"contributes": { "sources": [] },
		"permissions": { "network": [], "fs": [] }
	}`, id, version, plugins.PlatformKey(), entry)
	files := map[string][]byte{
		"plugin.json": []byte(manifest),
		entry:         []byte("test entry"),
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	var sums strings.Builder
	for _, name := range names {
		sum := sha256.Sum256(files[name])
		fmt.Fprintf(&sums, "%s  %s\n", hex.EncodeToString(sum[:]), name)
	}
	files["SHA256SUMS"] = []byte(sums.String())
	for name, data := range files {
		path := filepath.Join(source, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return source
}

func TestPluginInstallOrUpdateInstallsThenRefusesSameVersion(t *testing.T) {
	router := newRouter(t)
	ctx := context.Background()
	source := writeLocalPluginPackage(t, t.TempDir(), "demo", "1.0.0")
	payload, err := json.Marshal(map[string]any{"source": source, "allowUnsigned": true})
	if err != nil {
		t.Fatal(err)
	}

	out, err := router.Dispatch(ctx, "PluginInstallOrUpdate", payload)
	if err != nil {
		t.Fatalf("PluginInstallOrUpdate: %v", err)
	}
	info, ok := out.(plugins.PluginInfo)
	if !ok || info.ID != "demo" || info.Version != "1.0.0" {
		t.Fatalf("PluginInstallOrUpdate = %#v", out)
	}

	_, err = router.Dispatch(ctx, "PluginInstallOrUpdate", payload)
	if !errors.Is(err, plugins.ErrNoUpdate) || !strings.Contains(err.Error(), "demo 1.0.0") {
		t.Fatalf("same version: err = %v, want ErrNoUpdate with the id and version", err)
	}

	if _, err := router.Dispatch(ctx, "PluginInstall", payload); !errors.Is(err, plugins.ErrAlreadyInstalled) {
		t.Fatalf("PluginInstall on an installed id: err = %v, want ErrAlreadyInstalled", err)
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
	register(router, silentEvents{}, host, store, staticSettings{}, plugins.FetchRegistryIndex)
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
