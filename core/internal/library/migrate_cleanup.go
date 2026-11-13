package library

import (
	"context"
	"database/sql"
	"fmt"
)

// staleItchErrorMeta guards the one-time cleanup of the update history rows older itch builds
// wrote for every game when no API key was configured. The cleanup runs on the first start after
// the update and is skipped by every later start.
const staleItchErrorMeta = "cleanup-stale-itch-errors-v1"

// staleItchAPIKeyMessages are the exact missing-key errors old itch builds stored in the update
// history. The plain text is the message the source reported; the longer forms are how the plugin
// host rendered that error, with the itch prefix and the JSON-RPC code suffix. Only these known
// rows go: other itch failures, such as a rate limit or a login that expired, stay in the history.
var staleItchAPIKeyMessages = []string{
	"API key is not configured",
	"itch.io API key is not configured",
	"itch.io API key is not configured (code -32603)",
}

// cleanupStaleItchChecks removes the stale missing-key rows of the itch source from the update
// history. It deletes checks rows only, and it records the run in the meta table, so calling it
// again is a no-op.
func cleanupStaleItchChecks(ctx context.Context, db *sql.DB) error {
	var done int
	if err := db.QueryRowContext(ctx, `select count(*) from meta where key = ?`, staleItchErrorMeta).Scan(&done); err != nil {
		return fmt.Errorf("read stale itch cleanup marker: %w", err)
	}
	if done > 0 {
		return nil
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	args := make([]any, 0, len(staleItchAPIKeyMessages)+2)
	args = append(args, ResultError)
	for _, message := range staleItchAPIKeyMessages {
		args = append(args, message)
	}
	args = append(args, "itch")
	query := fmt.Sprintf(`delete from checks where result = ? and error in (%s)
		and game_id in (select id from games where source = ?)`, placeholders(len(staleItchAPIKeyMessages)))
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("delete stale itch checks: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `insert into meta (key, value) values (?, ?)
		on conflict(key) do update set value = excluded.value`,
		staleItchErrorMeta, nowStamp()); err != nil {
		return fmt.Errorf("mark stale itch cleanup: %w", err)
	}
	return tx.Commit()
}
