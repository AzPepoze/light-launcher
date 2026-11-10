package library

import (
	"context"
	"database/sql"
	"errors"
)

// deletedHistoryLimit keeps the history useful without letting it grow forever.
const deletedHistoryLimit = 500

// DeletedInput describes a game that is being removed from the library.
type DeletedInput struct {
	GameID       int64
	Source       string
	SourceID     string
	URL          string
	Title        string
	Version      string
	Developer    string
	Tags         []string
	InstallDir   string
	DeletedFiles bool
}

// AddDeleted records a removed game and trims the history to its limit.
func (s *Store) AddDeleted(ctx context.Context, in DeletedInput) (DeletedGame, error) {
	tags, err := encodeJSON(nonNilStrings(in.Tags))
	if err != nil {
		return DeletedGame{}, err
	}
	result, err := s.db.ExecContext(ctx, `insert into deleted_games (game_id, source, source_id, url, title, version,
		developer, tags, install_dir, deleted_files, deleted_at) values (?,?,?,?,?,?,?,?,?,?,?)`,
		in.GameID, in.Source, nullIfEmpty(in.SourceID), in.URL, in.Title, nullIfEmpty(in.Version),
		nullIfEmpty(in.Developer), tags, nullIfEmpty(in.InstallDir), boolInt(in.DeletedFiles), nowStamp())
	if err != nil {
		return DeletedGame{}, err
	}
	if _, err := s.db.ExecContext(ctx, `delete from deleted_games where id not in
		(select id from deleted_games order by deleted_at desc, id desc limit ?)`, deletedHistoryLimit); err != nil {
		return DeletedGame{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return DeletedGame{}, err
	}
	return s.GetDeleted(ctx, id)
}

// ListDeleted returns the newest removed games first.
func (s *Store) ListDeleted(ctx context.Context) ([]DeletedGame, error) {
	rows, err := s.db.QueryContext(ctx, `select id, game_id, source, source_id, url, title, version, developer, tags,
		install_dir, deleted_files, deleted_at from deleted_games order by deleted_at desc, id desc limit ?`,
		deletedHistoryLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []DeletedGame{}
	for rows.Next() {
		entry, err := scanDeleted(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, entry)
	}
	return out, rows.Err()
}

// GetDeleted returns one history entry, or ErrNotFound.
func (s *Store) GetDeleted(ctx context.Context, id int64) (DeletedGame, error) {
	entry, err := scanDeleted(s.db.QueryRowContext(ctx, `select id, game_id, source, source_id, url, title, version,
		developer, tags, install_dir, deleted_files, deleted_at from deleted_games where id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return DeletedGame{}, ErrNotFound
	}
	return entry, err
}

// RemoveDeleted forgets one history entry.
func (s *Store) RemoveDeleted(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `delete from deleted_games where id = ?`, id)
	return err
}

// ClearDeleted forgets the whole history.
func (s *Store) ClearDeleted(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `delete from deleted_games`)
	return err
}

func scanDeleted(row rowScanner) (DeletedGame, error) {
	var (
		entry                             DeletedGame
		sourceID, version, developer, dir sql.NullString
		tags                              string
		deletedFiles                      int
	)
	err := row.Scan(&entry.ID, &entry.GameID, &entry.Source, &sourceID, &entry.URL, &entry.Title, &version,
		&developer, &tags, &dir, &deletedFiles, &entry.DeletedAt)
	if err != nil {
		return DeletedGame{}, err
	}
	entry.SourceID = stringPtr(sourceID)
	entry.Version = stringPtr(version)
	entry.Developer = stringPtr(developer)
	entry.InstallDir = stringPtr(dir)
	entry.Tags = decodeStrings(tags)
	entry.DeletedFiles = deletedFiles != 0
	return entry, nil
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
