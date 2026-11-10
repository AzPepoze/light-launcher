package downloads

import (
	"context"
	"testing"
)

func TestPersistKeepsPlannedConnectionsNotRateLimitedLimit(t *testing.T) {
	store := NewMemoryStore()
	id, err := store.Insert(context.Background(), Download{URL: "https://example.test/a.zip", Dest: "a.zip", Connections: 4, State: StateDownloading})
	if err != nil {
		t.Fatal(err)
	}
	m := newTestManager(t, store, t.TempDir(), Settings{})

	j := newJob(Download{ID: id, URL: "https://example.test/a.zip", Dest: "a.zip", Connections: 4}, nil)
	j.setLayout(1000, `"v1"`, planSegments(1000, 4))
	j.gateRef().OnRateLimit()
	if limit := j.gateRef().Limit(); limit != 2 {
		t.Fatalf("limit after a rate limit = %d, want 2", limit)
	}

	m.persist(j)
	got, _ := store.Get(context.Background(), id)
	if got.Connections != 4 {
		t.Fatalf("stored connections = %d, want the planned 4", got.Connections)
	}
	if got.ETag != `"v1"` {
		t.Fatalf("stored etag = %q", got.ETag)
	}
}
