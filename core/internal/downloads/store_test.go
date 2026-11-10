package downloads

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func newSQLiteTestStore(t *testing.T) Store {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "downloads.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	store, err := NewSQLiteStore(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func storesUnderTest(t *testing.T) map[string]Store {
	return map[string]Store{
		"memory": NewMemoryStore(),
		"sqlite": newSQLiteTestStore(t),
	}
}

func TestStoreRoundTrip(t *testing.T) {
	for name, store := range storesUnderTest(t) {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			id, err := store.Insert(ctx, Download{
				GameID:      7,
				URL:         "https://example.test/a.zip",
				Filename:    "a.zip",
				Dest:        "/tmp/a.zip",
				Size:        100,
				Connections: 4,
				State:       StateQueued,
				Headers:     map[string]string{"Cookie": "x=1"},
				Install:     true,
			})
			if err != nil {
				t.Fatal(err)
			}
			got, err := store.Get(ctx, id)
			if err != nil {
				t.Fatal(err)
			}
			if got.ID != id || got.GameID != 7 || got.Filename != "a.zip" || !got.Install {
				t.Fatalf("unexpected row: %+v", got)
			}
			if got.Headers["Cookie"] != "x=1" || got.CreatedAt.IsZero() {
				t.Fatalf("headers or created time lost: %+v", got)
			}
		})
	}
}

func TestStorePatchKeepsUnsetFields(t *testing.T) {
	for name, store := range storesUnderTest(t) {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			id, _ := store.Insert(ctx, Download{URL: "u", Filename: "f", Dest: "d", Size: 10, Connections: 2, State: StateQueued})
			downloading := StateDownloading
			if err := store.Update(ctx, id, Patch{State: &downloading}); err != nil {
				t.Fatal(err)
			}
			done := int64(4)
			if err := store.Update(ctx, id, Patch{BytesDone: &done}); err != nil {
				t.Fatal(err)
			}
			started := time.Now()
			if err := store.Update(ctx, id, Patch{StartedAt: &started}); err != nil {
				t.Fatal(err)
			}
			got, _ := store.Get(ctx, id)
			if got.State != StateDownloading || got.BytesDone != 4 || got.Size != 10 || got.Connections != 2 {
				t.Fatalf("patch changed other fields: %+v", got)
			}
			if got.StartedAt.IsZero() {
				t.Fatal("started time not saved")
			}
		})
	}
}

func TestStoreSegmentsAndDelete(t *testing.T) {
	for name, store := range storesUnderTest(t) {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			id, _ := store.Insert(ctx, Download{URL: "u", Filename: "f", Dest: "d", State: StateQueued})
			segs := []Segment{{Index: 0, Start: 0, End: 50, Done: 20}, {Index: 1, Start: 50, End: 100}}
			if err := store.SaveSegments(ctx, id, segs); err != nil {
				t.Fatal(err)
			}
			got, err := store.Segments(ctx, id)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != 2 || got[0].Done != 20 || got[1].End != 100 {
				t.Fatalf("segments not stored: %+v", got)
			}
			if err := store.Delete(ctx, id); err != nil {
				t.Fatal(err)
			}
			if _, err := store.Get(ctx, id); err != ErrNotFound {
				t.Fatalf("want ErrNotFound after delete, got %v", err)
			}
			if err := store.Update(ctx, id, Patch{}); err != ErrNotFound {
				t.Fatalf("update of missing row: want ErrNotFound, got %v", err)
			}
		})
	}
}

func TestStoreListNewestFirst(t *testing.T) {
	for name, store := range storesUnderTest(t) {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			first, _ := store.Insert(ctx, Download{URL: "u", Filename: "a", Dest: "d", State: StateQueued})
			second, _ := store.Insert(ctx, Download{URL: "u", Filename: "b", Dest: "d", State: StateQueued})
			rows, err := store.List(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != 2 || rows[0].ID != second || rows[1].ID != first {
				t.Fatalf("want newest first, got %+v", rows)
			}
		})
	}
}
