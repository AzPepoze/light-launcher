package pluginsvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/plugins"
	"light-launcher/core/internal/types"
	"light-launcher/core/internal/uploads"
)

// staticSettings is a settings view that does not read settings.json.
type staticSettings struct {
	settings types.AppSettings
}

func (s staticSettings) Load() types.AppSettings {
	return s.settings
}

// newRegistryRouter registers the plugin methods with a fixed setting and a stubbed fetch.
func newRegistryRouter(t *testing.T, registryURL string, fetch registryFetch) *api.Router {
	t.Helper()
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
	register(router, silentEvents{}, host, store, staticSettings{
		settings: types.AppSettings{PluginRegistryURL: registryURL},
	}, fetch)
	return router
}

func registryEntry() plugins.RegistryIndexEntry {
	return plugins.RegistryIndexEntry{
		ID:          "demo",
		Name:        "Demo",
		Description: "Finds demo games.",
		Author:      "Ada",
		Repository:  "github:ada/demo",
		Version:     "1.2.3",
		DownloadURL: "https://plugins.example/demo.llplugin",
		SHA256:      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Homepage:    "https://example.test/demo",
	}
}

func TestPluginRegistryWithoutURLMakesNoRequest(t *testing.T) {
	called := false
	fetch := func(context.Context, string) (plugins.RegistryIndex, error) {
		called = true
		return plugins.RegistryIndex{}, errors.New("must not be called")
	}
	router := newRegistryRouter(t, "", fetch)

	out, err := router.Dispatch(context.Background(), "PluginRegistry", nil)
	if err != nil {
		t.Fatalf("PluginRegistry: %v", err)
	}
	result, ok := out.(plugins.RegistryIndexResult)
	if !ok {
		t.Fatalf("PluginRegistry = %#v", out)
	}
	if result.Configured || result.URL != "" || len(result.Entries) != 0 || result.Error != "" {
		t.Fatalf("result = %+v, want the unconfigured result", result)
	}
	encoded, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `{"configured":false,"url":"","entries":[]}` {
		t.Fatalf("PluginRegistry = %s", encoded)
	}
	if called {
		t.Fatal("an empty URL started a request")
	}
}

func TestPluginRegistryTrimsTheConfiguredURL(t *testing.T) {
	var gotURL string
	fetch := func(_ context.Context, url string) (plugins.RegistryIndex, error) {
		gotURL = url
		return plugins.RegistryIndex{}, nil
	}
	router := newRegistryRouter(t, "  https://plugins.example/index.json  ", fetch)

	out, err := router.Dispatch(context.Background(), "PluginRegistry", nil)
	if err != nil {
		t.Fatal(err)
	}
	if gotURL != "https://plugins.example/index.json" {
		t.Fatalf("fetched %q, want the trimmed URL", gotURL)
	}
	if result := out.(plugins.RegistryIndexResult); result.URL != gotURL || !result.Configured {
		t.Fatalf("result = %+v", result)
	}
}

func TestPluginRegistryReturnsFetchedEntries(t *testing.T) {
	entry := registryEntry()
	fetch := func(_ context.Context, url string) (plugins.RegistryIndex, error) {
		if url != "https://plugins.example/index.json" {
			t.Fatalf("url = %q", url)
		}
		return plugins.RegistryIndex{Version: plugins.RegistryIndexVersion, Plugins: []plugins.RegistryIndexEntry{entry}}, nil
	}
	router := newRegistryRouter(t, "https://plugins.example/index.json", fetch)

	out, err := router.Dispatch(context.Background(), "PluginRegistry", nil)
	if err != nil {
		t.Fatalf("PluginRegistry: %v", err)
	}
	result, ok := out.(plugins.RegistryIndexResult)
	if !ok || !result.Configured || result.Error != "" {
		t.Fatalf("result = %+v", out)
	}
	if len(result.Entries) != 1 || result.Entries[0].ID != entry.ID || result.Entries[0].SHA256 != entry.SHA256 {
		t.Fatalf("entries = %+v", result.Entries)
	}
}

func TestPluginRegistryReportsLoadErrorsInTheResult(t *testing.T) {
	fetch := func(context.Context, string) (plugins.RegistryIndex, error) {
		return plugins.RegistryIndex{}, fmt.Errorf("%w: HTTP 503", plugins.ErrRegistryFetch)
	}
	router := newRegistryRouter(t, "https://plugins.example/index.json", fetch)

	out, err := router.Dispatch(context.Background(), "PluginRegistry", nil)
	if err != nil {
		t.Fatalf("a load error must not fail the call: %v", err)
	}
	result, ok := out.(plugins.RegistryIndexResult)
	if !ok || !result.Configured || result.URL != "https://plugins.example/index.json" {
		t.Fatalf("result = %+v", out)
	}
	if result.Error == "" || len(result.Entries) != 0 {
		t.Fatalf("result = %+v, want an error and no entries", result)
	}
}

// TestPluginRegistryParsesAFixtureIndex runs the real fetcher against an httptest TLS fixture.
func TestPluginRegistryParsesAFixtureIndex(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"version":1,"plugins":[
			{"id":"demo","name":"Demo","description":"Finds demo games.","author":"Ada",
			 "repository":"github:ada/demo","version":"1.2.3",
			 "downloadUrl":"https://plugins.example/demo.llplugin",
			 "sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			 "homepage":"https://example.test/demo"}]}`)
	}))
	t.Cleanup(server.Close)

	fetch := plugins.RegistryIndexClient{Client: server.Client()}.Fetch
	router := newRegistryRouter(t, server.URL+"/plugins.json", fetch)

	out, err := router.Dispatch(context.Background(), "PluginRegistry", nil)
	if err != nil {
		t.Fatalf("PluginRegistry: %v", err)
	}
	result := out.(plugins.RegistryIndexResult)
	if result.Error != "" || len(result.Entries) != 1 {
		t.Fatalf("result = %+v", result)
	}
	if entry := result.Entries[0]; entry.Name != "Demo" || entry.Version != "1.2.3" {
		t.Fatalf("entry = %+v", entry)
	}

	// A broken fixture must come back in the result, not as a failed RPC.
	broken := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"version":2,"plugins":[]}`)
	}))
	t.Cleanup(broken.Close)
	router = newRegistryRouter(t, broken.URL+"/plugins.json", plugins.RegistryIndexClient{Client: broken.Client()}.Fetch)
	out, err = router.Dispatch(context.Background(), "PluginRegistry", nil)
	if err != nil {
		t.Fatalf("PluginRegistry: %v", err)
	}
	result = out.(plugins.RegistryIndexResult)
	if result.Error == "" || len(result.Entries) != 0 {
		t.Fatalf("result = %+v, want the version error", result)
	}
}
