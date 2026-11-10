package library

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOpenCreatesSchemaInWALMode(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()

	for _, table := range []string{"games", "game_files", "meta", "deleted_games", "checks", "http_cache", "thread_index", "thread_cover"} {
		var found string
		err := lib.db.QueryRowContext(ctx, `select name from sqlite_master where type = 'table' and name = ?`, table).Scan(&found)
		if err != nil {
			t.Errorf("table %s missing: %v", table, err)
		}
	}

	var mode string
	if err := lib.db.QueryRowContext(ctx, `pragma journal_mode`).Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if mode != "wal" {
		t.Errorf("journal mode = %q, want wal", mode)
	}
}

func TestReopenKeepsDataAndVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "library.db")
	first, err := Open(context.Background(), Options{DBPath: path})
	if err != nil {
		t.Fatal(err)
	}
	upsertGame(t, first, "keep", "Keep Me")
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	second, err := Open(context.Background(), Options{DBPath: path})
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer second.Close()
	page, err := second.List(context.Background(), Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 {
		t.Errorf("total after reopen = %d, want 1", page.Total)
	}

	var version int
	if err := second.db.QueryRowContext(context.Background(), `pragma user_version`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != len(migrations) {
		t.Errorf("user_version = %d, want %d", version, len(migrations))
	}
}

func TestOpenRequiresDBPath(t *testing.T) {
	if _, err := Open(context.Background(), Options{}); err == nil {
		t.Error("Open without DBPath should fail")
	}
}
