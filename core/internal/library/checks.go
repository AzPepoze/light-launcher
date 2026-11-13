package library

import (
	"context"
	"database/sql"
	"fmt"
)

// The result values of one update check, as stored in the checks table.
const (
	ResultUpdate  = "update"
	ResultCurrent = "current"
	ResultError   = "error"
)

// AddCheck records a version change found by an update check.
func (s *Store) AddCheck(ctx context.Context, gameID int64, oldVersion, newVersion, changelog string) error {
	return s.AddCheckResult(ctx, gameID, oldVersion, newVersion, changelog, ResultUpdate, "")
}

// AddCheckResult records one update check, including checks that found nothing or failed.
func (s *Store) AddCheckResult(ctx context.Context, gameID int64, oldVersion, newVersion, changelog, result, failure string) error {
	_, err := s.db.ExecContext(ctx, `insert into checks (game_id, checked_at, old_version, new_version, changelog,
		seen, applied, result, error) values (?,?,?,?,?,0,0,?,?)`,
		gameID, nowStamp(), nullIfEmpty(oldVersion), nullIfEmpty(newVersion), nullIfEmpty(changelog),
		result, nullIfEmpty(failure))
	return err
}

// ListChecks returns the newest recorded changes, optionally for one game.
func (s *Store) ListChecks(ctx context.Context, gameID int64, limit int) ([]Check, error) {
	query := `select id, game_id, checked_at, old_version, new_version, changelog, seen, applied
		from checks where result = ?`
	args := []any{ResultUpdate}
	if gameID != 0 {
		query += ` and game_id = ?`
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
	err := s.db.QueryRowContext(ctx, `select count(*) from checks where seen = 0 and result = ?
		and coalesce(new_version, '') <> coalesce(old_version, '')`, ResultUpdate).Scan(&count)
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

// UpdateHistoryEntry is one recorded update check, with the game it belongs to.
type UpdateHistoryEntry struct {
	ID        int64   `json:"id"`
	GameID    int64   `json:"gameId"`
	Title     string  `json:"title"`
	CheckedAt string  `json:"checkedAt"`
	From      *string `json:"from"`
	To        *string `json:"to"`
	Result    string  `json:"result"`
	Error     *string `json:"error"`
	CoverPath *string `json:"coverPath"`
	CoverURL  *string `json:"coverUrl"`
	// Source is the game's source id, so a source tab can page the server-side history.
	Source string `json:"source"`
}

// UpdateHistory returns recent checks, newest first. before is the id to page from, or zero for the first page.
// result keeps only one result value (update, current or error). Empty keeps all of them.
func (s *Store) UpdateHistory(ctx context.Context, gameID int64, limit int, before int64, result string) ([]UpdateHistoryEntry, error) {
	return s.UpdateHistorySource(ctx, gameID, limit, before, result, "")
}

// UpdateHistorySource is UpdateHistory with an optional source filter. Paging a source on the
// server keeps its pages full instead of filtering already-loaded rows in the browser.
func (s *Store) UpdateHistorySource(ctx context.Context, gameID int64, limit int, before int64, result, source string) ([]UpdateHistoryEntry, error) {
	if err := checkHistoryResult(result); err != nil {
		return nil, err
	}
	query := `select c.id, c.game_id, g.title, c.checked_at, c.old_version, c.new_version, c.result, c.error,
		g.cover_path, g.cover_url, g.source from checks c join games g on g.id = c.game_id where 1 = 1`
	args := []any{}
	if gameID != 0 {
		query += ` and c.game_id = ?`
		args = append(args, gameID)
	}
	if before > 0 {
		query += ` and c.id < ?`
		args = append(args, before)
	}
	if result != "" {
		query += ` and c.result = ?`
		args = append(args, result)
	}
	if source != "" {
		query += ` and g.source = ?`
		args = append(args, source)
	}
	query += ` order by c.id desc limit ?`
	args = append(args, clamp(limit, 1, 200, 50))
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []UpdateHistoryEntry{}
	for rows.Next() {
		var (
			entry                          UpdateHistoryEntry
			from, to, failure, cover, link sql.NullString
		)
		if err := rows.Scan(&entry.ID, &entry.GameID, &entry.Title, &entry.CheckedAt, &from, &to,
			&entry.Result, &failure, &cover, &link, &entry.Source); err != nil {
			return nil, err
		}
		entry.From = stringPtr(from)
		entry.To = stringPtr(to)
		entry.Error = stringPtr(failure)
		entry.CoverPath = stringPtr(cover)
		entry.CoverURL = stringPtr(link)
		out = append(out, entry)
	}
	return out, rows.Err()
}

// checkHistoryResult accepts an empty result filter or one of the three result values.
func checkHistoryResult(result string) error {
	switch result {
	case "", ResultUpdate, ResultCurrent, ResultError:
		return nil
	}
	return fmt.Errorf("unknown update result %q", result)
}

// MarkApplied marks every check of a game as applied and seen.
func (s *Store) MarkApplied(ctx context.Context, gameID int64) error {
	_, err := s.db.ExecContext(ctx, `update checks set applied = 1, seen = 1 where game_id = ?`, gameID)
	return err
}
