package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func validRegistryEntry() RegistryIndexEntry {
	return RegistryIndexEntry{
		ID:          "demo",
		Name:        "Demo",
		Description: "Finds demo games.",
		Author:      "Ada",
		Repository:  "github:ada/demo",
		Version:     "1.2.3",
		DownloadURL: "https://plugins.example/demo.llplugin",
		SHA256:      strings.Repeat("a", 64),
		Homepage:    "https://example.test/demo",
	}
}

func registryFixture(t *testing.T, entries ...RegistryIndexEntry) string {
	t.Helper()
	data, err := json.Marshal(RegistryIndex{Version: RegistryIndexVersion, Plugins: entries})
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func fetchRegistryFixture(t *testing.T, body string) (RegistryIndex, error) {
	t.Helper()
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
	t.Cleanup(server.Close)
	client := RegistryIndexClient{Client: server.Client(), Timeout: 2 * time.Second}
	return client.Fetch(context.Background(), server.URL+"/plugins.json")
}

func TestRegistryIndexFixtureParses(t *testing.T) {
	entry := validRegistryEntry()
	entry.SHA256 = strings.ToUpper(entry.SHA256)

	index, err := fetchRegistryFixture(t, registryFixture(t, entry))
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if index.Version != RegistryIndexVersion || len(index.Plugins) != 1 {
		t.Fatalf("index = %+v", index)
	}
	got := index.Plugins[0]
	if got.ID != entry.ID || got.Name != entry.Name || got.Description != entry.Description {
		t.Errorf("identity fields = %+v", got)
	}
	if got.Author != entry.Author || got.Repository != entry.Repository || got.Version != entry.Version {
		t.Errorf("info fields = %+v", got)
	}
	if got.DownloadURL != entry.DownloadURL || got.Homepage != entry.Homepage {
		t.Errorf("URL fields = %+v", got)
	}
	if got.SHA256 != strings.Repeat("a", 64) {
		t.Errorf("SHA256 = %q, want lowercase hex", got.SHA256)
	}
}

func TestRegistryIndexAllowsEmptyAndOptionalFields(t *testing.T) {
	index, err := fetchRegistryFixture(t, `{"version":1,"plugins":[]}`)
	if err != nil {
		t.Fatalf("empty index: %v", err)
	}
	if len(index.Plugins) != 0 {
		t.Fatalf("Plugins = %+v", index.Plugins)
	}

	entry := validRegistryEntry()
	entry.Author = ""
	entry.Description = ""
	entry.Homepage = ""
	entry.Repository = ""
	if _, err := fetchRegistryFixture(t, registryFixture(t, entry)); err != nil {
		t.Fatalf("entry without optional fields: %v", err)
	}

	entry.Repository = "https://github.com/ada/demo"
	if _, err := fetchRegistryFixture(t, registryFixture(t, entry)); err != nil {
		t.Fatalf("repository as https URL: %v", err)
	}
}

func TestRegistryIndexIgnoresUnknownFields(t *testing.T) {
	entry := validRegistryEntry()
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	body := fmt.Sprintf(`{"version":1,"generatedAt":"2026-01-01","plugins":[%s]}`, data)
	if _, err := fetchRegistryFixture(t, body); err != nil {
		t.Fatalf("unknown fields: %v", err)
	}
}

func TestRegistryIndexRejectsUnknownVersion(t *testing.T) {
	for name, body := range map[string]string{
		"unknown": `{"version":2,"plugins":[]}`,
		"missing": `{"plugins":[]}`,
		"zero":    `{"version":0,"plugins":[]}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := fetchRegistryFixture(t, body)
			if !errors.Is(err, ErrRegistryVersion) {
				t.Fatalf("err = %v, want ErrRegistryVersion", err)
			}
			if !strings.Contains(err.Error(), fmt.Sprintf("want %d", RegistryIndexVersion)) {
				t.Fatalf("err = %q, want the accepted version in the message", err)
			}
		})
	}
}

func TestRegistryIndexRejectsBadEntries(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*RegistryIndexEntry)
		want   string
	}{
		{"uppercase id", func(e *RegistryIndexEntry) { e.ID = "Demo" }, "id"},
		{"short id", func(e *RegistryIndexEntry) { e.ID = "d" }, "id"},
		{"empty name", func(e *RegistryIndexEntry) { e.Name = "  " }, "name"},
		{"long name", func(e *RegistryIndexEntry) { e.Name = strings.Repeat("a", maxRegistryNameBytes+1) }, "name"},
		{"long description", func(e *RegistryIndexEntry) {
			e.Description = strings.Repeat("a", maxRegistryDescriptionBytes+1)
		}, "description"},
		{"long author", func(e *RegistryIndexEntry) { e.Author = strings.Repeat("a", maxRegistryAuthorBytes+1) }, "author"},
		{"short version", func(e *RegistryIndexEntry) { e.Version = "1.2" }, "version"},
		{"leading zero version", func(e *RegistryIndexEntry) { e.Version = "01.2.3" }, "version"},
		{"http download", func(e *RegistryIndexEntry) { e.DownloadURL = "http://plugins.example/demo.llplugin" }, "downloadUrl"},
		{"missing download", func(e *RegistryIndexEntry) { e.DownloadURL = "" }, "downloadUrl"},
		{"short sha", func(e *RegistryIndexEntry) { e.SHA256 = "abc" }, "sha256"},
		{"non-hex sha", func(e *RegistryIndexEntry) { e.SHA256 = strings.Repeat("z", 64) }, "sha256"},
		{"missing sha", func(e *RegistryIndexEntry) { e.SHA256 = "" }, "sha256"},
		{"http homepage", func(e *RegistryIndexEntry) { e.Homepage = "http://example.test/demo" }, "homepage"},
		{"unknown repository", func(e *RegistryIndexEntry) { e.Repository = "gitlab:ada/demo" }, "repository"},
		{"http repository", func(e *RegistryIndexEntry) { e.Repository = "http://github.com/ada/demo" }, "repository"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entry := validRegistryEntry()
			tc.mutate(&entry)
			_, err := fetchRegistryFixture(t, registryFixture(t, entry))
			if !errors.Is(err, ErrRegistryInvalid) {
				t.Fatalf("err = %v, want ErrRegistryInvalid", err)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %q, want %q in the message", err, tc.want)
			}
		})
	}
}

func TestRegistryIndexRejectsDuplicateIDs(t *testing.T) {
	entry := validRegistryEntry()
	_, err := fetchRegistryFixture(t, registryFixture(t, entry, entry))
	if !errors.Is(err, ErrRegistryInvalid) {
		t.Fatalf("err = %v, want ErrRegistryInvalid", err)
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("err = %q, want a duplicate hint", err)
	}
}

func TestRegistryIndexRejectsTooManyEntries(t *testing.T) {
	entries := make([]RegistryIndexEntry, maxRegistryEntries+1)
	for i := range entries {
		entries[i] = validRegistryEntry()
		entries[i].ID = fmt.Sprintf("plugin-%d", i)
	}
	_, err := fetchRegistryFixture(t, registryFixture(t, entries...))
	if !errors.Is(err, ErrRegistryInvalid) {
		t.Fatalf("err = %v, want ErrRegistryInvalid", err)
	}
	if !strings.Contains(err.Error(), "at most") {
		t.Fatalf("err = %q, want the entry limit in the message", err)
	}
}

func TestRegistryIndexRejectsNonHTTPSURLBeforeAnyRequest(t *testing.T) {
	calls := 0
	client := httpDoerFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("unexpected request")
	})
	_, err := (RegistryIndexClient{Client: client}).Fetch(context.Background(), "http://plugins.example/index.json")
	if !errors.Is(err, ErrRegistryURL) {
		t.Fatalf("err = %v, want ErrRegistryURL", err)
	}
	if calls != 0 {
		t.Fatalf("made %d requests for an http URL, want none", calls)
	}
}

func TestRegistryIndexRefusesRedirectToHTTP(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://plugins.example/index.json", http.StatusFound)
	}))
	t.Cleanup(server.Close)

	client := RegistryIndexClient{Client: server.Client()}
	_, err := client.Fetch(context.Background(), server.URL+"/plugins.json")
	if !errors.Is(err, ErrRegistryURL) {
		t.Fatalf("err = %v, want ErrRegistryURL", err)
	}
}

func TestRegistryIndexFollowsHTTPSRedirect(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/plugins.json", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/index.json", http.StatusFound)
	})
	mux.HandleFunc("/index.json", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, registryFixture(t, validRegistryEntry()))
	})
	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	client := RegistryIndexClient{Client: server.Client()}
	index, err := client.Fetch(context.Background(), server.URL+"/plugins.json")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(index.Plugins) != 1 || index.Plugins[0].ID != "demo" {
		t.Fatalf("index = %+v", index)
	}
}

func TestRegistryIndexRefusesRedirectLoop(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/plugins.json", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/plugins.json", http.StatusFound)
	})
	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	client := RegistryIndexClient{Client: server.Client()}
	_, err := client.Fetch(context.Background(), server.URL+"/plugins.json")
	if !errors.Is(err, ErrRegistryFetch) || !strings.Contains(err.Error(), "too many redirects") {
		t.Fatalf("err = %v, want a too many redirects error", err)
	}
}

func TestRegistryIndexRefusesOversizedReply(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(make([]byte, maxRegistryIndexBytes+1))
	}))
	t.Cleanup(server.Close)

	client := RegistryIndexClient{Client: server.Client()}
	_, err := client.Fetch(context.Background(), server.URL+"/plugins.json")
	if !errors.Is(err, ErrRegistryTooLarge) {
		t.Fatalf("err = %v, want ErrRegistryTooLarge", err)
	}
}

func TestRegistryIndexStopsAtTimeout(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)

	client := RegistryIndexClient{Client: server.Client(), Timeout: 50 * time.Millisecond}
	start := time.Now()
	_, err := client.Fetch(context.Background(), server.URL+"/plugins.json")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want a deadline", err)
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("Fetch took %s, want the 50ms timeout to end it", elapsed)
	}
}

func TestRegistryIndexReportsHTTPAndJSONFailures(t *testing.T) {
	t.Run("status", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "down", http.StatusInternalServerError)
		}))
		t.Cleanup(server.Close)
		client := RegistryIndexClient{Client: server.Client()}
		_, err := client.Fetch(context.Background(), server.URL+"/plugins.json")
		if !errors.Is(err, ErrRegistryFetch) || !strings.Contains(err.Error(), "HTTP 500") {
			t.Fatalf("err = %v, want an HTTP 500 fetch error", err)
		}
	})

	for name, body := range map[string]string{
		"broken json": `{"version":`,
		"trailing":    `{"version":1,"plugins":[]} extra`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := fetchRegistryFixture(t, body)
			if !errors.Is(err, ErrRegistryInvalid) {
				t.Fatalf("err = %v, want ErrRegistryInvalid", err)
			}
		})
	}
}

func TestRegistryIndexReportsNetworkFailure(t *testing.T) {
	client := httpDoerFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})
	_, err := (RegistryIndexClient{Client: client}).Fetch(context.Background(), "https://plugins.example/index.json")
	if !errors.Is(err, ErrRegistryFetch) || !strings.Contains(err.Error(), "network down") {
		t.Fatalf("err = %v, want a fetch error with the cause", err)
	}
}

type httpDoerFunc func(*http.Request) (*http.Response, error)

func (f httpDoerFunc) Do(request *http.Request) (*http.Response, error) {
	return f(request)
}
