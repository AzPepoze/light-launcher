package library

import (
	"context"
	"path/filepath"
	"testing"

	"light-launcher/core/internal/contract"
)

// addGame saves one plugin game and returns its library id.
func cleanupAddGame(t *testing.T, lib *Library, source, sourceID, title string) int64 {
	t.Helper()
	id, err := lib.Upsert(context.Background(), contract.GameDTO{
		Source: source, SourceID: sourceID, URL: "https://" + source + ".example/" + sourceID, Title: title,
	})
	if err != nil {
		t.Fatalf("upsert %s: %v", title, err)
	}
	return id
}

// addErrorCheck records one failed update check for a game.
func cleanupAddErrorCheck(t *testing.T, lib *Library, gameID int64, message string) {
	t.Helper()
	if err := lib.AddCheckResult(context.Background(), gameID, "1.0", "", "", ResultError, message); err != nil {
		t.Fatalf("add error check %q: %v", message, err)
	}
}

func TestCleanupStaleItchUpdateErrorsRunsOnce(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "library.db")

	first, err := Open(ctx, Options{DBPath: path})
	if err != nil {
		t.Fatal(err)
	}
	itchID := cleanupAddGame(t, first, "itch", "1", "Itch Game")
	f95ID := cleanupAddGame(t, first, "f95", "9", "F95 Game")

	// A fresh database is already marked clean. Drop the marker to emulate a database written
	// before this cleanup existed, then seed the stale rows it must remove.
	if _, err := first.db.ExecContext(ctx, `delete from meta where key = ?`, staleItchErrorMeta); err != nil {
		t.Fatal(err)
	}
	for _, message := range staleItchAPIKeyMessages {
		cleanupAddErrorCheck(t, first, itchID, message)
	}
	cleanupAddErrorCheck(t, first, itchID, "itch.io rate limited (code -32004)")
	cleanupAddErrorCheck(t, first, f95ID, "API key is not configured")
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	// The next open runs the cleanup once: only the known itch missing-key rows go.
	second, err := Open(ctx, Options{DBPath: path})
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	assertStaleItchChecksRemoved(t, second)
	if _, found, err := second.GetMeta(ctx, staleItchErrorMeta); err != nil || !found {
		t.Errorf("cleanup marker = found %v, err %v, want the cleanup recorded", found, err)
	}
	if err := second.Close(); err != nil {
		t.Fatal(err)
	}

	// A later open skips the cleanup and must not delete the real failure a second time.
	third, err := Open(ctx, Options{DBPath: path})
	if err != nil {
		t.Fatalf("third open: %v", err)
	}
	defer third.Close()
	assertStaleItchChecksRemoved(t, third)
}

// assertStaleItchChecksRemoved checks that the known stale itch rows are gone, while the real itch
// failure and the same text on another source stay.
func assertStaleItchChecksRemoved(t *testing.T, lib *Library) {
	t.Helper()
	entries, err := lib.UpdateHistory(context.Background(), 0, 20, 0, ResultError)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("error history = %+v, want the two rows that are not stale itch missing-key errors", entries)
	}
	bySource := map[string]string{}
	for _, entry := range entries {
		bySource[entry.Source] = derefString(entry.Error)
	}
	if got := bySource["itch"]; got != "itch.io rate limited (code -32004)" {
		t.Errorf("itch error = %q, want the real rate-limit failure", got)
	}
	if got := bySource["f95"]; got != "API key is not configured" {
		t.Errorf("f95 error = %q, want the same text on another source to stay", got)
	}
}
