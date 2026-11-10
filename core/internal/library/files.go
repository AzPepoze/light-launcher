package library

import (
	"context"
	"database/sql"

	"light-launcher/core/internal/contract"
)

// ListFiles returns the stored download candidates of a game.
func (s *Store) ListFiles(ctx context.Context, gameID int64) ([]GameFile, error) {
	rows, err := s.db.QueryContext(ctx, `select id, game_id, label, url, filename, size, host, kind, platform,
		range_hint, requires_auth from game_files where game_id = ? order by id`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	files := []GameFile{}
	for rows.Next() {
		file, err := scanFile(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, rows.Err()
}

// ReplaceFiles swaps the stored download candidates of a game in one transaction.
func (s *Store) ReplaceFiles(ctx context.Context, gameID int64, candidates []contract.FileCandidate) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `delete from game_files where game_id = ?`, gameID); err != nil {
		return err
	}
	for _, file := range candidates {
		if err := insertFile(ctx, tx, gameID, file); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func insertFile(ctx context.Context, tx *sql.Tx, gameID int64, file contract.FileCandidate) error {
	headers, err := encodeOptionalMap(file.Headers)
	if err != nil {
		return err
	}
	platform := string(file.Platform)
	if platform == "" {
		platform = "unknown"
	}
	_, err = tx.ExecContext(ctx, `insert into game_files (game_id, label, url, filename, size, host, kind, platform,
		requires_auth, headers) values (?,?,?,?,?,?,?,?,?,?)`,
		gameID, file.Label, file.URL, nullIfEmpty(file.Filename), nullableSize(file.Size), nullIfEmpty(file.Host),
		file.Kind, platform, boolInt(file.RequiresAuth), headers)
	return err
}

func encodeOptionalMap(headers map[string]string) (any, error) {
	if headers == nil {
		return nil, nil
	}
	return encodeJSON(headers)
}

func nullableSize(size int64) any {
	if size == 0 {
		return nil
	}
	return size
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
