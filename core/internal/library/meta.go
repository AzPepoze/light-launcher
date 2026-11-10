package library

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

// GetMeta reads a key from the meta table. The bool is false when the key is missing.
func (s *Store) GetMeta(ctx context.Context, key string) (string, bool, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `select value from meta where key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	return value, err == nil, err
}

// SetMeta writes a key to the meta table.
func (s *Store) SetMeta(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx,
		`insert into meta (key, value) values (?, ?) on conflict(key) do update set value = excluded.value`,
		key, value)
	return err
}

// getMetaJSON reads a JSON value into out. A missing or broken value leaves out untouched.
func (s *Store) getMetaJSON(ctx context.Context, key string, out any) error {
	raw, found, err := s.GetMeta(ctx, key)
	if err != nil || !found {
		return err
	}
	_ = json.Unmarshal([]byte(raw), out)
	return nil
}

// setMetaJSON stores a value as JSON in the meta table.
func (s *Store) setMetaJSON(ctx context.Context, key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.SetMeta(ctx, key, string(raw))
}
