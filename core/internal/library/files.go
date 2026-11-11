package library

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

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

// candidate returns one saved file as a plugin sees it. It returns ErrNotFound when the game has no such file.
func (s *Store) candidate(ctx context.Context, gameID, fileID int64) (contract.FileCandidate, error) {
	var (
		file                              contract.FileCandidate
		pluginID, filename, host, headers sql.NullString
		size                              sql.NullInt64
		platform                          string
		requiresAuth                      int
	)
	err := s.db.QueryRowContext(ctx, `select plugin_file_id, label, url, filename, size, host, kind, platform,
		requires_auth, headers from game_files where game_id = ? and id = ?`, gameID, fileID).Scan(
		&pluginID, &file.Label, &file.URL, &filename, &size, &host, &file.Kind, &platform, &requiresAuth, &headers)
	if errors.Is(err, sql.ErrNoRows) {
		return contract.FileCandidate{}, ErrNotFound
	}
	if err != nil {
		return contract.FileCandidate{}, err
	}
	file.ID = pluginID.String
	file.Filename = filename.String
	file.Size = size.Int64
	file.Host = host.String
	file.Platform = contract.Platform(platform)
	file.RequiresAuth = requiresAuth == 1
	file.Headers, err = decodeHeaders(headers)
	if err != nil {
		return contract.FileCandidate{}, err
	}
	return file, nil
}

// decodeHeaders reads the headers a file was saved with. Missing headers decode to nil.
func decodeHeaders(raw sql.NullString) (map[string]string, error) {
	if !raw.Valid || raw.String == "" {
		return nil, nil
	}
	var headers map[string]string
	if err := json.Unmarshal([]byte(raw.String), &headers); err != nil {
		return nil, fmt.Errorf("read file headers: %w", err)
	}
	return headers, nil
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
		requires_auth, headers, plugin_file_id) values (?,?,?,?,?,?,?,?,?,?,?)`,
		gameID, file.Label, file.URL, nullIfEmpty(file.Filename), nullableSize(file.Size), nullIfEmpty(file.Host),
		file.Kind, platform, boolInt(file.RequiresAuth), headers, nullIfEmpty(file.ID))
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
