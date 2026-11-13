package library

import (
	"context"
	"testing"
)

func TestUpdateHistoryCarriesCover(t *testing.T) {
	lib := openTestLibrary(t, Options{Sources: updateSources(newlyUpdated)})
	ctx := context.Background()
	game := f95Game(t, lib, "1.0")
	if _, err := lib.Patch(ctx, game.ID, Patch{CoverPath: strPtr("covers/harbor.jpg"), CoverURL: strPtr("https://f95/harbor.jpg")}); err != nil {
		t.Fatal(err)
	}
	if _, err := lib.CheckGame(ctx, game.ID); err != nil {
		t.Fatal(err)
	}

	entries, err := lib.UpdateHistory(ctx, 0, 10, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("history = %+v, want one entry", entries)
	}
	entry := entries[0]
	if entry.CoverPath == nil || *entry.CoverPath != "covers/harbor.jpg" || entry.CoverURL == nil || *entry.CoverURL != "https://f95/harbor.jpg" {
		t.Errorf("cover = %v, %v, want the game's cover", entry.CoverPath, entry.CoverURL)
	}
}

func TestUpdateHistoryFiltersByResult(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	game := f95Game(t, lib, "1.0")
	for _, result := range []string{ResultUpdate, ResultCurrent, ResultCurrent, ResultError} {
		if err := lib.AddCheckResult(ctx, game.ID, "1.0", "1.0", "", result, ""); err != nil {
			t.Fatal(err)
		}
	}

	cases := map[string]int{"": 4, ResultUpdate: 1, ResultCurrent: 2, ResultError: 1}
	for result, want := range cases {
		entries, err := lib.UpdateHistory(ctx, 0, 10, 0, result)
		if err != nil {
			t.Fatalf("filter %q: %v", result, err)
		}
		if len(entries) != want {
			t.Errorf("filter %q = %d entries, want %d", result, len(entries), want)
		}
		for _, entry := range entries {
			if result != "" && entry.Result != result {
				t.Errorf("filter %q returned a %s entry", result, entry.Result)
			}
		}
	}

	if _, err := lib.UpdateHistory(ctx, 0, 10, 0, "skipped"); err == nil {
		t.Error("an unknown result filter must be an error")
	}
}
