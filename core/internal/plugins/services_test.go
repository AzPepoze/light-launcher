package plugins

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"light-launcher/core/internal/contract"
)

// testEnv returns a plugin environment with the given grants and a temp data folder.
func testEnv(t *testing.T, perms Permissions, folders map[string]string) *pluginEnv {
	t.Helper()
	manifest := Manifest{
		ID:          "demo",
		Permissions: perms,
		Contributes: Contributes{Sources: []SourceContribution{{ID: "demo", Label: "Demo", Capabilities: []string{"browse"}}}},
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	deps := &serviceDeps{
		secrets:   newFileSecrets(t.TempDir()),
		folders:   folders,
		transport: transport,
		limiter:   newHostLimiter(0),
		cache:     newETagCache(),
	}
	return newPluginEnv(manifest, t.TempDir(), deps, newLineRing(10))
}

func callService(t *testing.T, env *pluginEnv, method string, params any) (any, error) {
	t.Helper()
	raw, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	return env.call(context.Background(), method, raw)
}

func expectCode(t *testing.T, err error, code int) {
	t.Helper()
	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Code != code {
		t.Fatalf("err = %v, want RPC code %d", err, code)
	}
}

func TestHostMatchesWildcards(t *testing.T) {
	cases := []struct {
		entry, host string
		want        bool
	}{
		{"example.com", "example.com", true},
		{"example.com", "a.example.com", false},
		{"*.example.com", "a.example.com", true},
		{"*.example.com", "example.com", false},
		{"*.example.com", "evilexample.com", false},
		{"Example.COM", "example.com", true},
	}
	for _, tc := range cases {
		if got := hostMatches(tc.entry, tc.host); got != tc.want {
			t.Errorf("hostMatches(%q, %q) = %v, want %v", tc.entry, tc.host, got, tc.want)
		}
	}
}

func TestServicesDeniedWithoutGrant(t *testing.T) {
	env := testEnv(t, Permissions{}, nil)
	_, err := callService(t, env, "secrets.get", map[string]string{"key": "token"})
	expectCode(t, err, CodePermissionDenied)

	_, err = callService(t, env, "browser.assist", map[string]string{"url": "https://example.com"})
	expectCode(t, err, CodePermissionDenied)

	_, err = callService(t, env, "http.request", map[string]string{"url": "https://denied.example/"})
	expectCode(t, err, CodePermissionDenied)

	_, err = callService(t, env, "no.such.method", nil)
	expectCode(t, err, CodeMethodNotFound)
}

func TestHTTPRequestFetchesAllowedHost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "hello page")
	}))
	defer server.Close()
	env := testEnv(t, Permissions{Network: []string{"127.0.0.1"}}, nil)

	result, err := callService(t, env, "http.request", map[string]string{"url": server.URL + "/page"})
	if err != nil {
		t.Fatalf("http.request: %v", err)
	}
	got := result.(httpRequestResult)
	if got.Status != http.StatusOK || got.Text != "hello page" {
		t.Fatalf("result = %+v", got)
	}
}

func TestHTTPRedirectToDeniedHostIsBlocked(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://localhost:"+portOf(server.URL)+"/other", http.StatusFound)
	}))
	defer server.Close()
	env := testEnv(t, Permissions{Network: []string{"127.0.0.1"}}, nil)

	_, err := callService(t, env, "http.request", map[string]string{"url": server.URL})
	expectCode(t, err, CodePermissionDenied)
}

func TestHTTPNotModifiedWithETag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") == `"v1"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `"v1"`)
		fmt.Fprint(w, "fresh")
	}))
	defer server.Close()
	env := testEnv(t, Permissions{Network: []string{"127.0.0.1"}}, nil)
	params := map[string]string{"url": server.URL, "cacheKey": "list"}

	first, err := callService(t, env, "http.request", params)
	if err != nil || first.(httpRequestResult).Text != "fresh" {
		t.Fatalf("first request: %v %+v", err, first)
	}
	second, err := callService(t, env, "http.request", params)
	if err != nil || !second.(httpRequestResult).NotModified {
		t.Fatalf("second request should be not modified: %v %+v", err, second)
	}
}

func TestHTTPResponseSizeLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, strings.Repeat("x", 100))
	}))
	defer server.Close()
	env := testEnv(t, Permissions{Network: []string{"127.0.0.1"}}, nil)

	_, err := callService(t, env, "http.request", map[string]any{"url": server.URL, "maxBytes": 10})
	expectCode(t, err, CodeRemoteError)
}

func TestHTTPRateLimitedAfterRetries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()
	env := testEnv(t, Permissions{Network: []string{"127.0.0.1"}}, nil)

	_, err := callService(t, env, "http.request", map[string]any{"url": server.URL, "retries": 1})
	expectCode(t, err, CodeRateLimited)
}

func TestStorageRoundTrip(t *testing.T) {
	env := testEnv(t, Permissions{}, nil)
	if _, err := callService(t, env, "storage.set", map[string]any{"key": "cursor", "value": map[string]int{"page": 3}}); err != nil {
		t.Fatalf("storage.set: %v", err)
	}
	value, err := callService(t, env, "storage.get", map[string]string{"key": "cursor"})
	if err != nil {
		t.Fatalf("storage.get: %v", err)
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, value.(json.RawMessage)); err != nil || compact.String() != `{"page":3}` {
		t.Fatalf("value = %s, %v", value, err)
	}
}

func TestSecretsRoundTripThroughService(t *testing.T) {
	env := testEnv(t, Permissions{Secrets: true}, nil)
	if _, err := callService(t, env, "secrets.set", map[string]string{"key": "token", "value": "abc"}); err != nil {
		t.Fatalf("secrets.set: %v", err)
	}
	value, err := callService(t, env, "secrets.get", map[string]string{"key": "token"})
	if err != nil || value != "abc" {
		t.Fatalf("secrets.get = %v, %v", value, err)
	}
	if _, err := callService(t, env, "secrets.delete", map[string]string{"key": "token"}); err != nil {
		t.Fatalf("secrets.delete: %v", err)
	}
	if value, _ := callService(t, env, "secrets.get", map[string]string{"key": "token"}); value != nil {
		t.Fatalf("deleted secret is still set: %v", value)
	}
}

func TestExtractStaysInGrantedFolders(t *testing.T) {
	base := t.TempDir()
	games := filepath.Join(base, "games")
	if err := os.MkdirAll(games, 0o755); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(games, "pack.zip")
	writePackageFile(t, archive, map[string][]byte{"readme.txt": []byte("hi")})
	env := testEnv(t, Permissions{FS: []string{"$GAMES"}}, map[string]string{"$GAMES": games})

	result, err := callService(t, env, "extract", map[string]string{
		"archive":     archive,
		"destination": filepath.Join(games, "out"),
	})
	if err != nil {
		t.Fatalf("extract inside grant: %v", err)
	}
	if result.(map[string]string)["destination"] == "" {
		t.Fatalf("extract returned no destination")
	}

	_, err = callService(t, env, "extract", map[string]string{
		"archive":     archive,
		"destination": filepath.Join(base, "elsewhere"),
	})
	expectCode(t, err, CodePermissionDenied)

	_, err = callService(t, env, "extract", map[string]string{"archive": "pack.zip", "destination": games})
	expectCode(t, err, CodeInvalidParams)
}

func TestLibraryRejectsSourceOfAnotherPlugin(t *testing.T) {
	env := testEnv(t, Permissions{}, nil)
	_, err := callService(t, env, "library.upsert", map[string]any{
		"games": []contract.GameDTO{{Source: "other", SourceID: "1", URL: "https://x.example", Title: "A"}},
	})
	expectCode(t, err, CodeInvalidParams)
}

func TestDownloadsNeedAllowedHostAndPlainName(t *testing.T) {
	env := testEnv(t, Permissions{Network: []string{"files.example"}}, nil)
	_, err := callService(t, env, "downloads.enqueue", map[string]string{
		"url": "https://other.example/file.zip", "filename": "file.zip",
	})
	expectCode(t, err, CodePermissionDenied)

	_, err = callService(t, env, "downloads.enqueue", map[string]string{
		"url": "https://files.example/file.zip", "filename": "../file.zip",
	})
	expectCode(t, err, CodeInvalidParams)
}

func portOf(rawURL string) string {
	return rawURL[strings.LastIndex(rawURL, ":")+1:]
}
