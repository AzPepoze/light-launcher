package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type githubFake struct {
	server   *httptest.Server
	requests atomic.Int32
	latest   string
	list     string
	files    map[string]string
}

func newGithubFake(t *testing.T) *githubFake {
	t.Helper()
	fake := &githubFake{files: map[string]string{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/ada/demo/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		fake.requests.Add(1)
		if r.Header.Get("If-None-Match") == `"release-1"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		if fake.latest == "" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("ETag", `"release-1"`)
		fmt.Fprint(w, fake.latest)
	})
	mux.HandleFunc("/repos/ada/demo/releases", func(w http.ResponseWriter, r *http.Request) {
		fake.requests.Add(1)
		fmt.Fprint(w, fake.list)
	})
	mux.HandleFunc("/files/", func(w http.ResponseWriter, r *http.Request) {
		body, ok := fake.files[strings.TrimPrefix(r.URL.Path, "/files/")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, body)
	})
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://evil.invalid/SHA256SUMS", http.StatusFound)
	})
	fake.server = httptest.NewTLSServer(mux)
	t.Cleanup(fake.server.Close)
	return fake
}

func (f *githubFake) checker(t *testing.T, allowPrereleases bool, now time.Time) *Checker {
	t.Helper()
	return NewChecker(Options{
		Client:           f.server.Client(),
		Now:              func() time.Time { return now },
		AllowPrereleases: allowPrereleases,
		APIBaseURL:       f.server.URL,
	})
}

func releaseJSON(serverURL, tag string, assets ...string) string {
	var items []string
	for _, asset := range assets {
		items = append(items, asset)
	}
	return fmt.Sprintf(`{
  "tag_name": %q,
  "draft": false,
  "prerelease": false,
  "body": "Fixes a login bug.",
  "published_at": "2026-03-01T10:00:00Z",
  "assets": [%s]
}`, tag, strings.Join(items, ","))
}

func assetJSON(name, serverURL, path string) string {
	return fmt.Sprintf(`{"name": %q, "browser_download_url": %q}`, name, serverURL+path)
}

func TestCheckFindsNewerReleaseAndChecksum(t *testing.T) {
	fake := newGithubFake(t)
	sum := sha256.Sum256([]byte("package bytes"))
	digest := hex.EncodeToString(sum[:])
	fake.latest = releaseJSON(fake.server.URL, "v1.1.0",
		assetJSON("demo-1.1.0.llplugin", fake.server.URL, "/files/demo.llplugin"),
		assetJSON(checksumsAssetName, fake.server.URL, "/files/SHA256SUMS"),
	)
	fake.files["SHA256SUMS"] = digest + "  demo-1.1.0.llplugin\n"

	result, err := fake.checker(t, false, time.Unix(1700000000, 0)).Check(context.Background(), Request{
		ID: "demo", Repository: "github:ada/demo", Version: "1.0.0",
	})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !result.Available || result.Version != "1.1.0" {
		t.Fatalf("result = %+v", result)
	}
	if result.AssetURL != fake.server.URL+"/files/demo.llplugin" {
		t.Errorf("asset URL = %q", result.AssetURL)
	}
	if result.SHA256 != digest || result.SHA256URL != fake.server.URL+"/files/SHA256SUMS" {
		t.Errorf("checksum = %q from %q", result.SHA256, result.SHA256URL)
	}
	if result.Notes != "Fixes a login bug." || result.PublishedAt != "2026-03-01T10:00:00Z" {
		t.Errorf("release info = %+v", result)
	}
	if fake.requests.Load() != 1 {
		t.Errorf("requests = %d, want one API request", fake.requests.Load())
	}
}

func TestCheckWithoutAssetIsNotAvailable(t *testing.T) {
	fake := newGithubFake(t)
	fake.latest = releaseJSON(fake.server.URL, "1.1.0")
	result, err := fake.checker(t, false, time.Unix(1700000000, 0)).Check(context.Background(), Request{
		ID: "demo", Repository: "github:ada/demo", Version: "1.0.0",
	})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if result.Available {
		t.Fatalf("result = %+v, want no update", result)
	}
}

func TestCheckIgnoresSameOrOlderVersion(t *testing.T) {
	for _, tag := range []string{"1.0.0", "0.9.0"} {
		fake := newGithubFake(t)
		fake.latest = releaseJSON(fake.server.URL, tag, assetJSON("demo.llplugin", fake.server.URL, "/files/demo"))
		result, err := fake.checker(t, false, time.Unix(1700000000, 0)).Check(context.Background(), Request{
			ID: "demo", Repository: "github:ada/demo", Version: "1.0.0",
		})
		if err != nil {
			t.Fatalf("Check(%s): %v", tag, err)
		}
		if result.Available {
			t.Errorf("tag %s: result = %+v, want no update", tag, result)
		}
	}
}

func TestCheckRejectsBadVersionAndRepository(t *testing.T) {
	fake := newGithubFake(t)
	fake.latest = releaseJSON(fake.server.URL, "nightly")
	checker := fake.checker(t, false, time.Unix(1700000000, 0))
	if _, err := checker.Check(context.Background(), Request{ID: "demo", Repository: "github:ada/demo", Version: "1.0.0"}); err == nil || !strings.Contains(err.Error(), "semantic version") {
		t.Fatalf("bad tag: err = %v", err)
	}
	if _, err := checker.Check(context.Background(), Request{ID: "demo", Repository: "gitlab:ada/demo", Version: "1.0.0"}); !errors.Is(err, ErrInvalidRepository) {
		t.Fatalf("bad repository: err = %v", err)
	}
	if _, err := checker.Check(context.Background(), Request{ID: "demo", Repository: "github:ada/demo", Version: "dev"}); err == nil {
		t.Fatalf("bad installed version must fail")
	}
}

func TestCheckUsesETagAndCache(t *testing.T) {
	fake := newGithubFake(t)
	fake.latest = releaseJSON(fake.server.URL, "1.1.0", assetJSON("demo.llplugin", fake.server.URL, "/files/demo"))
	checker := fake.checker(t, false, time.Unix(1700000000, 0))
	request := Request{ID: "demo", Repository: "github:ada/demo", Version: "1.0.0"}

	first, err := checker.Check(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	second, err := checker.Check(context.Background(), request)
	if err != nil {
		t.Fatalf("second check: %v", err)
	}
	if first != second {
		t.Fatalf("cached result = %+v, want %+v", second, first)
	}
	if fake.requests.Load() != 2 {
		t.Fatalf("requests = %d, want two (second is conditional)", fake.requests.Load())
	}
}

func TestCheckRateLimitRefusesFurtherCalls(t *testing.T) {
	var requests atomic.Int32
	now := time.Unix(1700000000, 0)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", fmt.Sprint(now.Add(time.Minute).Unix()))
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()
	checker := NewChecker(Options{Client: server.Client(), Now: func() time.Time { return now }, APIBaseURL: server.URL})
	request := Request{ID: "demo", Repository: "github:ada/demo", Version: "1.0.0"}

	if _, err := checker.Check(context.Background(), request); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("first check: err = %v, want ErrRateLimited", err)
	}
	if _, err := checker.Check(context.Background(), request); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("second check: err = %v, want ErrRateLimited", err)
	}
	if requests.Load() != 1 {
		t.Fatalf("requests = %d, want one before the backoff ends", requests.Load())
	}
}

func TestCheckPrereleaseSetting(t *testing.T) {
	fake := newGithubFake(t)
	fake.latest = releaseJSON(fake.server.URL, "1.0.0", assetJSON("demo.llplugin", fake.server.URL, "/files/demo"))
	fake.list = "[" + releaseJSON(fake.server.URL, "1.1.0-beta.1", assetJSON("demo.llplugin", fake.server.URL, "/files/demo")) + "]"
	request := Request{ID: "demo", Repository: "github:ada/demo", Version: "0.9.0"}

	stable, err := fake.checker(t, false, time.Unix(1700000000, 0)).Check(context.Background(), request)
	if err != nil || stable.Version != "1.0.0" {
		t.Fatalf("stable check = %+v, %v", stable, err)
	}
	pre, err := fake.checker(t, true, time.Unix(1700000000, 0)).Check(context.Background(), request)
	if err != nil || pre.Version != "1.1.0-beta.1" {
		t.Fatalf("prerelease check = %+v, %v", pre, err)
	}
}

func TestCheckPicksAssetNamedAfterPlugin(t *testing.T) {
	fake := newGithubFake(t)
	fake.latest = releaseJSON(fake.server.URL, "1.1.0",
		assetJSON("other-1.1.0.llplugin", fake.server.URL, "/files/other"),
		assetJSON("demo-1.1.0.llplugin", fake.server.URL, "/files/demo"),
	)
	result, err := fake.checker(t, false, time.Unix(1700000000, 0)).Check(context.Background(), Request{
		ID: "demo", Repository: "github:ada/demo", Version: "1.0.0",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(result.AssetURL, "/files/demo") {
		t.Fatalf("asset URL = %q, want the demo asset", result.AssetURL)
	}
}

func TestCheckRefusesForeignHosts(t *testing.T) {
	fake := newGithubFake(t)
	fake.latest = releaseJSON(fake.server.URL, "1.1.0", assetJSON("demo.llplugin", "https://evil.invalid", "/demo.llplugin"))
	checker := fake.checker(t, false, time.Unix(1700000000, 0))
	if _, err := checker.Check(context.Background(), Request{ID: "demo", Repository: "github:ada/demo", Version: "1.0.0"}); !errors.Is(err, ErrUntrustedHost) {
		t.Fatalf("foreign asset: err = %v, want ErrUntrustedHost", err)
	}

	fake.latest = releaseJSON(fake.server.URL, "1.1.0",
		assetJSON("demo.llplugin", fake.server.URL, "/files/demo"),
		assetJSON(checksumsAssetName, fake.server.URL, "/redirect"),
	)
	if _, err := fake.checker(t, false, time.Unix(1700000000, 0)).Check(context.Background(), Request{ID: "demo", Repository: "github:ada/demo", Version: "1.0.0"}); !errors.Is(err, ErrUntrustedHost) {
		t.Fatalf("redirect: err = %v, want ErrUntrustedHost", err)
	}
}

func TestChecksumFor(t *testing.T) {
	digest := strings.Repeat("a", 64)
	cases := map[string]string{
		digest + "  demo.llplugin":   digest,
		digest + " *demo.llplugin\r": digest,
	}
	for body, want := range cases {
		got, err := checksumFor([]byte(body+"\n"), "demo.llplugin")
		if err != nil || got != want {
			t.Errorf("checksumFor(%q) = %q, %v; want %q", body, got, err, want)
		}
	}
	if _, err := checksumFor([]byte(digest+"  other.llplugin\n"), "demo.llplugin"); err == nil {
		t.Errorf("a missing entry must be an error")
	}
	if _, err := checksumFor([]byte("short  demo.llplugin\n"), "demo.llplugin"); err == nil {
		t.Errorf("a short digest must be an error")
	}
}

func TestCheckUsesLauncherUserAgent(t *testing.T) {
	var agent string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agent = r.Header.Get("User-Agent")
		fmt.Fprint(w, releaseJSON("https://example.test", "1.1.0", assetJSON("demo.llplugin", "https://example.test", "/demo")))
	}))
	defer server.Close()
	checker := NewChecker(Options{Client: server.Client(), APIBaseURL: server.URL, UserAgent: "light-launcher/2.0.0", AssetHosts: []string{"example.test"}})
	if _, err := checker.Check(context.Background(), Request{ID: "demo", Repository: "github:ada/demo", Version: "1.0.0"}); err != nil {
		t.Fatal(err)
	}
	if agent != "light-launcher/2.0.0" {
		t.Fatalf("User-Agent = %q", agent)
	}
}
