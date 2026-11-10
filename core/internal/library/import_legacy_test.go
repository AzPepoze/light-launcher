package library

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"light-launcher/core/internal/contract"
)

// legacySchema is the games and files tables of the old interesting-launcher library.db.
const legacySchema = `
create table games (
	id integer primary key autoincrement,
	source text not null,
	source_id text not null,
	url text not null,
	title text not null,
	version text,
	developer text,
	engine text,
	tags text not null default '[]',
	cover_path text,
	cover_url text,
	screenshots text not null default '[]',
	description text,
	changelog text,
	rating real,
	status text not null default 'unknown',
	favorite integer not null default 0,
	hidden integer not null default 0,
	notes text,
	installed_version text,
	install_dir text,
	install_size integer,
	exe_path text,
	exe_args text,
	last_played_at text,
	playtime_s integer not null default 0,
	source_updated_at text,
	last_checked_at text,
	added_at text not null,
	updated_at text not null,
	extra text,
	unique(source, source_id)
);
create table files (
	id integer primary key autoincrement,
	game_id integer not null references games(id) on delete cascade,
	label text not null,
	url text not null,
	filename text,
	size integer,
	host text,
	kind text not null default 'mirror',
	platform text not null default 'unknown',
	range_hint integer,
	requires_auth integer not null default 0,
	headers text
);
`

// makeLegacyDB writes an old-style library.db with two games and two files.
func makeLegacyDB(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "library.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(legacySchema); err != nil {
		t.Fatal(err)
	}
	statements := []string{
		`insert into games (id, source, source_id, url, title, status, notes, install_dir, playtime_s, added_at, updated_at)
		 values (1, 'f95', '42', 'https://f95/42', 'Harbor', 'installed', 'great', '/games/harbor', 500, 'now', 'now')`,
		`insert into games (id, source, source_id, url, title, status, added_at, updated_at)
		 values (2, 'itch', '7', 'https://itch/7', 'Tide', 'completed', 'now', 'now')`,
		`insert into files (game_id, label, url, kind) values (1, 'Mirror A', 'https://a', 'mirror')`,
		`insert into files (game_id, label, url, kind) values (1, 'Mirror B', 'https://b', 'mirror')`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func TestImportLegacyMergesExistingAndAddsNew(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	id, err := lib.Upsert(ctx, contract.GameDTO{Source: "f95", SourceID: "42", URL: "https://f95/42", Title: "Harbor"})
	if err != nil {
		t.Fatal(err)
	}
	existing, _ := lib.Get(ctx, id)
	if _, err := lib.Patch(ctx, existing.ID, Patch{PlaytimeS: ptrInt(100)}); err != nil {
		t.Fatal(err)
	}

	report, err := lib.ImportLegacy(ctx, makeLegacyDB(t))
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if report.Games != 1 || report.Updated != 1 || report.Files != 2 {
		t.Errorf("report = %+v, want 1 added, 1 updated, 2 files", report)
	}

	harbor, _ := lib.Get(ctx, id)
	if harbor.PlaytimeS != 500 {
		t.Errorf("playtime = %d, want the larger legacy value 500", harbor.PlaytimeS)
	}
	if derefString(harbor.Notes) != "great" || derefString(harbor.InstallDir) != "/games/harbor" {
		t.Errorf("notes=%v install=%v, want legacy values merged", harbor.Notes, harbor.InstallDir)
	}
	files, _ := lib.ListFiles(ctx, id)
	if len(files) != 2 {
		t.Errorf("files = %d, want 2", len(files))
	}

	tide, err := lib.BySource(ctx, "itch", "7")
	if err != nil || tide.Title != "Tide" {
		t.Errorf("itch game = %+v err=%v", tide, err)
	}
}

func TestImportLegacyTwiceAddsNothingNew(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	path := makeLegacyDB(t)
	if _, err := lib.ImportLegacy(ctx, path); err != nil {
		t.Fatal(err)
	}
	report, err := lib.ImportLegacy(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if report.Games != 0 || report.Files != 0 {
		t.Errorf("second import report = %+v, want nothing new", report)
	}
}

func TestImportLegacyRejectsOtherDatabases(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	path := filepath.Join(t.TempDir(), "other.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`create table unrelated (x integer)`); err != nil {
		t.Fatal(err)
	}
	db.Close()

	if _, err := lib.ImportLegacy(context.Background(), path); err == nil {
		t.Error("a database without games and files must be rejected")
	}
	// the attach must be released after a failed check, so the library still works
	if _, err := lib.List(context.Background(), Filter{}); err != nil {
		t.Errorf("library unusable after failed import: %v", err)
	}
}

func ptrInt(value int64) *int64 {
	return &value
}
