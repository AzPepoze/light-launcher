package library

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestUpdatesSummaryReportsLastCheck(t *testing.T) {
	sources := updateSources(func(any) (any, error) {
		return map[string]any{"update": map[string]string{"version": "2.0"}}, nil
	})
	lib := openTestLibrary(t, Options{Sources: sources})
	ctx := context.Background()

	before, err := lib.UpdatesSummary(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if before.Last != nil || before.Unseen != 0 {
		t.Errorf("before any check = %+v, want no run and no updates", before)
	}

	f95Game(t, lib, "1.0")
	if _, err := lib.CheckAll(ctx, nil, false); err != nil {
		t.Fatal(err)
	}
	summary, err := lib.UpdatesSummary(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Unseen != 1 {
		t.Errorf("unseen = %d, want 1", summary.Unseen)
	}
	if summary.Last == nil || summary.Last.CheckedAt == "" || len(summary.Last.Results) != 1 {
		t.Fatalf("last = %+v, want the check with one result", summary.Last)
	}
	if !summary.Last.Results[0].HasUpdate {
		t.Errorf("last result = %+v, want an update", summary.Last.Results[0])
	}
}

func TestForcedCheckIgnoresRecheckInterval(t *testing.T) {
	calls := atomic.Int32{}
	sources := updateSources(func(any) (any, error) {
		calls.Add(1)
		return nil, nil
	})
	lib := openTestLibrary(t, Options{Sources: sources, UpdateRule: recheckEvery(time.Hour)})
	ctx := context.Background()
	game := f95Game(t, lib, "1.0")
	stamp := nowStamp()
	if _, err := lib.Patch(ctx, game.ID, Patch{LastCheckedAt: &stamp}); err != nil {
		t.Fatal(err)
	}

	if _, err := lib.CheckAll(ctx, nil, true); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 0 {
		t.Errorf("a normal check called the source %d times, want 0 for a recent game", calls.Load())
	}
	if _, err := lib.CheckAll(ctx, nil, false); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Errorf("a forced check called the source %d times, want 1", calls.Load())
	}
}
