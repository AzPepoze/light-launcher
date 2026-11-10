package library

import (
	"context"
	"reflect"
	"testing"

	"light-launcher/core/internal/contract"
)

func TestUpsertInsertsThenMergesEmptyFields(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	dto := contract.GameDTO{Source: "f95", SourceID: "42", URL: "https://example/42", Title: "Harbor",
		Version: "0.7", Tags: []string{"adventure"}, CoverURL: "https://example/cover.png"}

	id, err := lib.Upsert(ctx, dto)
	if err != nil {
		t.Fatal(err)
	}

	// a refresh without version or tags keeps what we had
	_, err = lib.Upsert(ctx, contract.GameDTO{Source: "f95", SourceID: "42", URL: "https://example/42", Title: "Harbor 2"})
	if err != nil {
		t.Fatal(err)
	}
	game, err := lib.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if game.Title != "Harbor 2" {
		t.Errorf("title = %q, want Harbor 2", game.Title)
	}
	if derefString(game.Version) != "0.7" {
		t.Errorf("version = %v, want 0.7 kept", game.Version)
	}
	if !reflect.DeepEqual(game.Tags, []string{"adventure"}) {
		t.Errorf("tags = %v, want kept", game.Tags)
	}
	if game.Status != "unknown" {
		t.Errorf("status = %q, want unknown default", game.Status)
	}
}

func TestUpsertClearsCoverPathWhenCoverURLChanges(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	game := upsertGame(t, lib, "c1", "Cover Game")
	cover := "/cache/old.png"
	if _, err := lib.Patch(ctx, game.ID, Patch{CoverURL: strPtr("https://a/old.png"), CoverPath: &cover}); err != nil {
		t.Fatal(err)
	}
	if _, err := lib.Upsert(ctx, contract.GameDTO{Source: "manual", SourceID: "c1", URL: "u", Title: "Cover Game", CoverURL: "https://a/new.png"}); err != nil {
		t.Fatal(err)
	}
	refreshed, _ := lib.Get(ctx, game.ID)
	if refreshed.CoverPath != nil {
		t.Errorf("cover path = %v, want cleared after cover URL change", *refreshed.CoverPath)
	}
}

func TestPatchSetsAndClearsFields(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	game := upsertGame(t, lib, "p1", "Patch Me")

	notes := "good ending"
	favorite := true
	updated, err := lib.Patch(ctx, game.ID, Patch{Notes: &notes, Favorite: &favorite})
	if err != nil {
		t.Fatal(err)
	}
	if derefString(updated.Notes) != "good ending" || updated.Favorite != 1 {
		t.Errorf("after set: notes=%v favorite=%d", updated.Notes, updated.Favorite)
	}

	empty := ""
	cleared, err := lib.Patch(ctx, game.ID, Patch{Notes: &empty})
	if err != nil {
		t.Fatal(err)
	}
	if cleared.Notes != nil {
		t.Errorf("notes = %v, want NULL after empty string", *cleared.Notes)
	}
	if cleared.Favorite != 1 {
		t.Error("favorite must stay when only notes change")
	}
}

func TestListFiltersSearchesAndPages(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	harbor := upsertGame(t, lib, "h", "Moonlit Harbor")
	upsertGame(t, lib, "b", "Brass Tide")
	hidden := upsertGame(t, lib, "x", "Hidden One")
	hide := true
	if _, err := lib.Patch(ctx, hidden.ID, Patch{Hidden: &hide}); err != nil {
		t.Fatal(err)
	}
	installDir := "/games/harbor"
	if _, err := lib.Patch(ctx, harbor.ID, Patch{InstallDir: &installDir}); err != nil {
		t.Fatal(err)
	}

	all, err := lib.List(ctx, Filter{Sort: "title", Order: "asc"})
	if err != nil {
		t.Fatal(err)
	}
	if all.Total != 2 || all.Items[0].Title != "Brass Tide" {
		t.Errorf("visible list total=%d first=%q, want 2 and Brass Tide", all.Total, all.Items[0].Title)
	}

	withHidden, _ := lib.List(ctx, Filter{Hidden: true})
	if withHidden.Total != 3 {
		t.Errorf("with hidden total = %d, want 3", withHidden.Total)
	}

	search, _ := lib.List(ctx, Filter{Search: "harbor"})
	if search.Total != 1 || search.Items[0].ID != harbor.ID {
		t.Errorf("search total = %d, want the harbor game", search.Total)
	}

	installed := true
	inst, _ := lib.List(ctx, Filter{Installed: &installed})
	if inst.Total != 1 {
		t.Errorf("installed total = %d, want 1", inst.Total)
	}

	paged, _ := lib.List(ctx, Filter{Limit: 1, Offset: 1, Sort: "title", Order: "asc"})
	if paged.Total != 2 || len(paged.Items) != 1 || paged.Items[0].Title != "Moonlit Harbor" {
		t.Errorf("page = %+v, want second title with total 2", paged)
	}
}

func TestTagCountsAndStats(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	_, err := lib.Upsert(ctx, contract.GameDTO{Source: "manual", SourceID: "t1", URL: "u", Title: "One", Tags: []string{"rpg", "dark"}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lib.Upsert(ctx, contract.GameDTO{Source: "manual", SourceID: "t2", URL: "u", Title: "Two", Tags: []string{"rpg"}})
	if err != nil {
		t.Fatal(err)
	}

	tags, err := lib.TagCounts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 2 || tags[0] != (TagCount{Tag: "rpg", Count: 2}) {
		t.Errorf("tag counts = %+v", tags)
	}

	stats, err := lib.Stats(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != 2 || stats.Installed != 0 {
		t.Errorf("stats = %+v", stats)
	}
}

func TestDeletedGamesKeepSourceAndLink(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()

	entry, err := lib.AddDeleted(ctx, DeletedInput{
		GameID:       42,
		Source:       "f95",
		SourceID:     "thread-42",
		URL:          "https://f95zone.to/threads/42/",
		Title:        "Deleted Game",
		Version:      "1.0",
		Developer:    "Dev",
		Tags:         []string{"tag"},
		InstallDir:   "/games/deleted",
		DeletedFiles: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if entry.GameID != 42 || entry.Source != "f95" || entry.URL != "https://f95zone.to/threads/42/" {
		t.Errorf("entry = %+v", entry)
	}
	if !reflect.DeepEqual(entry.Tags, []string{"tag"}) || !entry.DeletedFiles {
		t.Errorf("tags or deletedFiles wrong: %+v", entry)
	}
	list, err := lib.ListDeleted(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || !reflect.DeepEqual(list[0], entry) {
		t.Errorf("list = %+v, want the same entry", list)
	}
}

func TestDeletedGamesCanBeForgottenAndCleared(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()

	entry, err := lib.AddDeleted(ctx, DeletedInput{GameID: 7, Source: "itch", SourceID: "7", URL: "https://example.itch.io/game", Title: "Itch Game"})
	if err != nil {
		t.Fatal(err)
	}
	if err := lib.RemoveDeleted(ctx, entry.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := lib.GetDeleted(ctx, entry.ID); err != ErrNotFound {
		t.Errorf("GetDeleted after remove = %v, want ErrNotFound", err)
	}

	if _, err := lib.AddDeleted(ctx, DeletedInput{GameID: 8, Source: "manual", SourceID: "local", URL: "/games/local", Title: "Local Game"}); err != nil {
		t.Fatal(err)
	}
	if err := lib.ClearDeleted(ctx); err != nil {
		t.Fatal(err)
	}
	list, _ := lib.ListDeleted(ctx)
	if len(list) != 0 {
		t.Errorf("list after clear = %d entries", len(list))
	}
}

func TestReplaceFilesAndListFiles(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	game := upsertGame(t, lib, "f1", "With Files")

	err := lib.ReplaceFiles(ctx, game.ID, []contract.FileCandidate{
		{Label: "Mirror A", URL: "https://a/x.zip", Kind: "mirror", Size: 100, RequiresAuth: true},
		{Label: "Mirror B", URL: "https://b/x.zip", Kind: "mirror"},
	})
	if err != nil {
		t.Fatal(err)
	}
	files, err := lib.ListFiles(ctx, game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || files[0].Label != "Mirror A" || files[0].RequiresAuth != 1 || files[1].Platform != "unknown" {
		t.Errorf("files = %+v", files)
	}

	if err := lib.ReplaceFiles(ctx, game.ID, nil); err != nil {
		t.Fatal(err)
	}
	files, _ = lib.ListFiles(ctx, game.ID)
	if len(files) != 0 {
		t.Errorf("files after replace with nothing = %d", len(files))
	}
}

func TestChecksCountUnseenAndMarkSeen(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	game := upsertGame(t, lib, "k1", "Checked")

	if err := lib.AddCheck(ctx, game.ID, "1.0", "1.1", "fixes"); err != nil {
		t.Fatal(err)
	}
	if err := lib.AddCheck(ctx, game.ID, "1.1", "1.1", ""); err != nil {
		t.Fatal(err)
	}
	unseen, err := lib.UnseenCount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if unseen != 1 {
		t.Errorf("unseen = %d, want 1 (same-version checks do not count)", unseen)
	}

	if err := lib.MarkSeen(ctx, nil); err != nil {
		t.Fatal(err)
	}
	unseen, _ = lib.UnseenCount(ctx)
	if unseen != 0 {
		t.Errorf("unseen after mark = %d, want 0", unseen)
	}
}

func TestMetaJSONRoundTrip(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	if err := lib.setMetaJSON(ctx, "k", []string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	var got []string
	if err := lib.getMetaJSON(ctx, "k", &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Errorf("meta = %v", got)
	}
}
