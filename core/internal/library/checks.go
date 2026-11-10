package library

import (
	"context"
	"database/sql"
)

// AddCheck records a version change found by an update check.
func (s *Store) AddCheck(ctx context.Context, gameID int64, oldVersion, newVersion, changelog string) error {
	_, err := s.db.ExecContext(ctx, `insert into checks (game_id, checked_at, old_version, new_version, changelog,
		seen, applied) values (?,?,?,?,?,0,0)`,
		gameID, nowStamp(), nullIfEmpty(oldVersion), nullIfEmpty(newVersion), nullIfEmpty(changelog))
	return err
}

// ListChecks returns the newest recorded changes, optionally for one game.
func (s *Store) ListChecks(ctx context.Context, gameID int64, limit int) ([]Check, error) {
	query := `select id, game_id, checked_at, old_version, new_version, changelog, seen, applied from checks`
	args := []any{}
	if gameID != 0 {
		query += ` where game_id = ?`
		args = append(args, gameID)
	}
	query += ` order by id desc limit ?`
	args = append(args, clamp(limit, 1, 1000, 100))
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Check{}
	for rows.Next() {
		var (
			c                            Check
			oldVersion, newVersion, text sql.NullString
		)
		if err := rows.Scan(&c.ID, &c.GameID, &c.CheckedAt, &oldVersion, &newVersion, &text, &c.Seen, &c.Applied); err != nil {
			return nil, err
		}
		c.OldVersion = stringPtr(oldVersion)
		c.NewVersion = stringPtr(newVersion)
		c.Changelog = stringPtr(text)
		out = append(out, c)
	}
	return out, rows.Err()
}

// UnseenCount counts changes the user has not looked at yet.
func (s *Store) UnseenCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `select count(*) from checks where seen = 0
		and coalesce(new_version, '') <> coalesce(old_version, '')`).Scan(&count)
	return count, err
}

// MarkSeen marks the given checks as seen, or all of them when ids is empty.
func (s *Store) MarkSeen(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		_, err := s.db.ExecContext(ctx, `update checks set seen = 1`)
		return err
	}
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	_, err := s.db.ExecContext(ctx, `update checks set seen = 1 where id in (`+placeholders(len(ids))+`)`, args...)
	return err
}

// MarkApplied marks every check of a game as applied and seen.
func (s *Store) MarkApplied(ctx context.Context, gameID int64) error {
	_, err := s.db.ExecContext(ctx, `update checks set applied = 1, seen = 1 where game_id = ?`, gameID)
	return err
}
