package library

import (
	"context"
	"database/sql"
	"fmt"
)

// migration is one versioned step. Versions are stored in PRAGMA user_version.
type migration struct {
	version    int
	statements []string
}

// The first migration creates the v1 schema. Later steps are appended.
var migrations = []migration{
	{
		version: 1,
		statements: []string{
			`create table if not exists games (
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
			)`,
			`create table if not exists game_files (
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
			)`,
			`create index if not exists game_files_game on game_files(game_id)`,
			`create table if not exists checks (
				id integer primary key autoincrement,
				game_id integer not null references games(id) on delete cascade,
				checked_at text not null,
				old_version text,
				new_version text,
				changelog text,
				seen integer not null default 0,
				applied integer not null default 0
			)`,
			`create index if not exists checks_game on checks(game_id, checked_at desc)`,
			`create table if not exists deleted_games (
				id integer primary key autoincrement,
				game_id integer not null,
				source text not null,
				source_id text,
				url text not null,
				title text not null,
				version text,
				developer text,
				tags text not null default '[]',
				install_dir text,
				deleted_files integer not null default 0,
				deleted_at text not null
			)`,
			`create index if not exists deleted_games_deleted on deleted_games(deleted_at desc)`,
			`create table if not exists meta (key text primary key, value text)`,
			`create table if not exists http_cache (url text primary key, etag text, last_modified text, fetched_at text)`,
			`create table if not exists thread_index (
				url text primary key,
				source text not null,
				slug text not null,
				title text not null,
				norm text not null,
				updated_at text
			)`,
			`create index if not exists thread_index_source on thread_index(source)`,
			`create index if not exists thread_index_norm on thread_index(norm)`,
			`create table if not exists thread_cover (url text primary key, cover_path text not null)`,
		},
	},
	{
		version: 2,
		statements: []string{
			`alter table checks add column result text not null default 'update'`,
			`alter table checks add column error text`,
		},
	},
	{
		// plugin_file_id keeps the id the plugin gave the file, which source.prepareDownload needs.
		version: 3,
		statements: []string{
			`alter table game_files add column plugin_file_id text`,
		},
	},
}

func migrate(ctx context.Context, db *sql.DB) error {
	var current int
	if err := db.QueryRowContext(ctx, `pragma user_version`).Scan(&current); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	for _, m := range migrations {
		if m.version <= current {
			continue
		}
		if err := applyMigration(ctx, db, m); err != nil {
			return fmt.Errorf("migration %d: %w", m.version, err)
		}
	}
	return nil
}

func applyMigration(ctx context.Context, db *sql.DB, m migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, statement := range m.statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`pragma user_version = %d`, m.version)); err != nil {
		return err
	}
	return tx.Commit()
}
