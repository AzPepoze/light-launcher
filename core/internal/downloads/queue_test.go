package downloads

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"light-launcher/core/internal/contract"
)

// inFlightTracker records how many distinct paths are being served at the same time.
type inFlightTracker struct {
	next    http.Handler
	mu      sync.Mutex
	current map[string]bool
	peak    int
}

func (t *inFlightTracker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		t.mark(r.URL.Path, true)
		defer t.mark(r.URL.Path, false)
	}
	t.next.ServeHTTP(w, r)
}

func (t *inFlightTracker) mark(path string, on bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current == nil {
		t.current = map[string]bool{}
	}
	if on {
		t.current[path] = true
		t.peak = max(t.peak, len(t.current))
	} else {
		delete(t.current, path)
	}
}

func (t *inFlightTracker) peakInFlight() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.peak
}

func TestQueueRunsNoMoreThanMaxActive(t *testing.T) {
	tracker := &inFlightTracker{}
	mux := http.NewServeMux()
	for i := 0; i < 4; i++ {
		mux.Handle(fmt.Sprintf("/part%d.bin", i), &testFile{data: patternData(600_000), etag: `"v1"`, slow: true})
	}
	tracker.next = mux
	srv := httptest.NewServer(tracker)
	defer srv.Close()

	m := newTestManager(t, nil, t.TempDir(), Settings{Connections: 1, MaxActive: 2, HostConnections: 8, Retries: 1})
	var ids []int64
	for i := 0; i < 4; i++ {
		id, err := m.Enqueue(context.Background(), contract.DownloadRequest{URL: fmt.Sprintf("%s/part%d.bin", srv.URL, i)})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	for _, id := range ids {
		waitForState(t, m, id, StateDone)
	}
	if peak := tracker.peakInFlight(); peak != 2 {
		t.Fatalf("peak concurrent downloads = %d, want 2", peak)
	}
}
