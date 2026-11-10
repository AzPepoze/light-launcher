package library

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
)

// ImportReport counts what a legacy import changed.
type ImportReport struct {
	Games   int64 `json:"games"`
	Updated int64 `json:"updated"`
	Files   int64 `json:"files"`
}

// ImportLegacy copies games, files, playtime and notes from an old interesting-launcher library.db.
// Games are matched by source and source id. Existing rows keep their own data, and only empty
// fields and playtime get the legacy values. The legacy file is only read.
func (l *Library) ImportLegacy(ctx context.Context, path string) (ImportReport, error) {
	if _, err := os.Stat(path); err != nil {
		return ImportReport{}, fmt.Errorf("legacy database: %w", err)
	}
	// One connection holds ATTACH for the whole import, and blocks other queries until it is released.
	conn, err := l.db.Conn(ctx)
	if err != nil {
		return ImportReport{}, err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `attach database ? as legacy`, path); err != nil {
		return ImportReport{}, fmt.Errorf("open legacy database: %w", err)
	}
	defer conn.ExecContext(context.Background(), `detach database legacy`)

	if err := checkLegacySchema(ctx, conn); err != nil {
		return ImportReport{}, err
	}
	report, err := importLegacyRows(ctx, conn)
	if err != nil {
		return ImportReport{}, err
	}
	return report, nil
}

func checkLegacySchema(ctx context.Context, conn *sql.Conn) error {
	var tables int
	err := conn.QueryRowContext(ctx,
		`select count(*) from legacy.sqlite_master where type = 'table' and name in ('games', 'files')`).Scan(&tables)
	if err != nil {
		return err
	}
	if tables != 2 {
		return errors.New("this is not an interesting-launcher library database")
	}
	return nil
}

func importLegacyRows(ctx context.Context, conn *sql.Conn) (ImportReport, error) {
	var report ImportReport
	var legacyGames int64
	if err := conn.QueryRowContext(ctx, `select count(*) from legacy.games`).Scan(&legacyGames); err != nil {
		return report, err
	}
	before, err := countGames(ctx, conn)
	if err != nil {
		return report, err
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return report, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, legacyGamesSQL); err != nil {
		return report, fmt.Errorf("import games: %w", err)
	}
	files, err := tx.ExecContext(ctx, legacyFilesSQL)
	if err != nil {
		return report, fmt.Errorf("import files: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return report, err
	}

	after, err := countGames(ctx, conn)
	if err != nil {
		return report, err
	}
	report.Games = after - before
	report.Updated = legacyGames - report.Games
	report.Files, err = files.RowsAffected()
	return report, err
}

func countGames(ctx context.Context, conn *sql.Conn) (int64, error) {
	var n int64
	err := conn.QueryRowContext(ctx, `select count(*) from main.games`).Scan(&n)
	return n, err
}

// legacyGamesSQL inserts new games and merges install data and playtime into existing ones.
// The legacy cover_path points into the old data folder, so it is not copied.
const legacyGamesSQL = `insert into main.games (source, source_id, url, title, version, developer, engine, tags,
	cover_url, screenshots, description, changelog, rating, status, favorite, hidden, notes, installed_version,
	install_dir, install_size, exe_path, exe_args, last_played_at, playtime_s, source_updated_at, last_checked_at,
	added_at, updated_at, extra)
select source, source_id, url, title, version, developer, engine, tags,
	cover_url, screenshots, description, changelog, rating, status, favorite, hidden, notes, installed_version,
	install_dir, install_size, exe_path, exe_args, last_played_at, playtime_s, source_updated_at, last_checked_at,
	added_at, updated_at, extra
from legacy.games where true
on conflict(source, source_id) do update set
	playtime_s = max(games.playtime_s, excluded.playtime_s),
	favorite = max(games.favorite, excluded.favorite),
	notes = coalesce(games.notes, excluded.notes),
	install_dir = coalesce(games.install_dir, excluded.install_dir),
	install_size = coalesce(games.install_size, excluded.install_size),
	exe_path = coalesce(games.exe_path, excluded.exe_path),
	exe_args = coalesce(games.exe_args, excluded.exe_args),
	installed_version = coalesce(games.installed_version, excluded.installed_version)`

// legacyFilesSQL copies download candidates, but only for games that have none yet.
const legacyFilesSQL = `insert into main.game_files (game_id, label, url, filename, size, host, kind, platform,
	range_hint, requires_auth, headers)
select g.id, f.label, f.url, f.filename, f.size, f.host, f.kind, f.platform, f.range_hint, f.requires_auth, f.headers
from legacy.files f
join legacy.games lg on lg.id = f.game_id
join main.games g on g.source = lg.source and g.source_id = lg.source_id
where not exists (select 1 from main.game_files x where x.game_id = g.id)`
