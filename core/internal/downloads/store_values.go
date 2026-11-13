package downloads

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"light-launcher/core/internal/contract"
)

// encodeHeaders stores request headers as JSON. No headers means NULL.
func encodeHeaders(h map[string]string) (any, error) {
	if len(h) == 0 {
		return nil, nil
	}
	data, err := json.Marshal(h)
	if err != nil {
		return nil, fmt.Errorf("encode headers: %w", err)
	}
	return string(data), nil
}

// decodeHeaders ignores a damaged value rather than failing the whole row.
func decodeHeaders(v sql.NullString) map[string]string {
	if !v.Valid || v.String == "" {
		return nil
	}
	var h map[string]string
	if err := json.Unmarshal([]byte(v.String), &h); err != nil {
		return nil
	}
	return h
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// encodeDecrypt stores the decrypt spec as JSON. No spec means NULL. The key is stored as it is,
// because the download can only be decrypted after a restart if the key comes back with the row.
func encodeDecrypt(d *contract.DecryptSpec) (any, error) {
	if d == nil {
		return nil, nil
	}
	data, err := json.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("encode decrypt: %w", err)
	}
	return string(data), nil
}

// decodeDecrypt reads a stored decrypt spec. A damaged value fails the row instead of silently
// installing an encrypted archive.
func decodeDecrypt(v sql.NullString) (*contract.DecryptSpec, error) {
	if !v.Valid || v.String == "" {
		return nil, nil
	}
	var d contract.DecryptSpec
	if err := json.Unmarshal([]byte(v.String), &d); err != nil {
		return nil, fmt.Errorf("decode decrypt: %w", err)
	}
	return &d, nil
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// nullBoolInt stores an optional flag. Nil means NULL, which keeps the settings' choice.
func nullBoolInt(b *bool) any {
	if b == nil {
		return nil
	}
	return boolInt(*b)
}

// decodeBoolInt reads a flag stored by nullBoolInt.
func decodeBoolInt(v sql.NullInt64) *bool {
	if !v.Valid {
		return nil
	}
	b := v.Int64 != 0
	return &b
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func optionalTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return formatTime(t)
}

func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, s)
}

func parseOptionalTime(v sql.NullString) (time.Time, error) {
	if !v.Valid || v.String == "" {
		return time.Time{}, nil
	}
	return parseTime(v.String)
}

func stateArg(s *State) any {
	if s == nil {
		return nil
	}
	return string(*s)
}

func timeArg(t *time.Time) any {
	if t == nil {
		return nil
	}
	return formatTime(*t)
}

// ptrArg turns an optional value into a SQL argument. Nil keeps the column as it is.
func ptrArg[T any](p *T) any {
	if p == nil {
		return nil
	}
	return *p
}
