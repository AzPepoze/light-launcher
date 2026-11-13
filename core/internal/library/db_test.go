package library

import (
	"context"
	"database/sql"
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
	if totalOf(page) != 1 {
		t.Errorf("total after reopen = %d, want 1", totalOf(page))
	}

	var version int
	if err := second.db.QueryRowContext(context.Background(), `pragma user_version`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != len(migrations) {
		t.Errorf("user_version = %d, want %d", version, len(migrations))
	}
}

func TestMigrationFromV2KeepsSavedFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "library.db")
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for _, m := range migrations[:2] {
		for _, statement := range m.statements {
			if _, err := db.ExecContext(ctx, statement); err != nil {
				t.Fatal(err)
			}
		}
	}
	if _, err := db.ExecContext(ctx, `pragma user_version = 2`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `insert into games
		(source, source_id, url, title, tags, screenshots, status, added_at, updated_at)
		values ('f95', '9', 'https://f95.example/9', 'Old Game', '[]', '[]', 'unknown', '2026-01-01', '2026-01-01')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `insert into game_files (game_id, label, url, kind, platform, headers)
		values (1, 'Saved', 'https://files.example/a.zip', 'direct', 'unknown', '{"X-Token":"t"}')`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	lib, err := Open(ctx, Options{DBPath: path})
	if err != nil {
		t.Fatalf("upgrade v2 database: %v", err)
	}
	defer lib.Close()
	files, err := lib.ListFiles(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Label != "Saved" {
		t.Fatalf("files after upgrade = %+v, want the saved file", files)
	}
	saved, err := lib.candidate(ctx, 1, files[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID != "" || saved.Headers["X-Token"] != "t" {
		t.Fatalf("candidate = %+v, want no plugin id and the saved headers", saved)
	}
}

func TestMigrationFromV1KeepsChecks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "library.db")
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range migrations[0].statements {
		if _, err := db.ExecContext(context.Background(), statement); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.ExecContext(context.Background(), `pragma user_version = 1`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(context.Background(), `insert into games
		(source, source_id, url, title, tags, screenshots, status, added_at, updated_at)
		values ('manual', 'old-1', 'file://old-1', 'Old Game', '[]', '[]', 'unknown', '2026-01-01', '2026-01-01')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(context.Background(), `insert into checks
		(game_id, checked_at, old_version, new_version, changelog, seen, applied)
		values (1, '2026-02-01', '1.0', '2.0', '', 0, 0)`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	lib, err := Open(context.Background(), Options{DBPath: path})
	if err != nil {
		t.Fatalf("upgrade v1 database: %v", err)
	}
	defer lib.Close()
	entries, err := lib.UpdateHistory(context.Background(), 0, 10, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Title != "Old Game" || entries[0].Result != ResultUpdate {
		t.Errorf("history after upgrade = %+v, want the old check as an update", entries)
	}
}

func TestOpenRequiresDBPath(t *testing.T) {
	if _, err := Open(context.Background(), Options{}); err == nil {
		t.Error("Open without DBPath should fail")
	}
}
