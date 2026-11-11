package library

import (
	"context"
	"testing"
	"time"

	"light-launcher/core/internal/contract"
)

const batchStamp = "2026-01-01T00:00:00.000Z"

// batchGame is a plugin game with a fixed update stamp and version, so the batch can skip it when it repeats.
func batchGame(sourceID, title string) contract.GameDTO {
	return contract.GameDTO{
		Source: "f95", SourceID: sourceID, URL: "https://f95/" + sourceID, Title: title,
		Version: "1.0", UpdatedAt: batchStamp,
	}
}

func TestUpsertManyReturnsIDsInOrderAndKeepsTheLastDuplicate(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()

	ids, err := lib.UpsertMany(ctx, []contract.GameDTO{
		batchGame("a", "Alpha"), batchGame("b", "Beta"), batchGame("a", "Alpha again"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 3 || ids[0] != ids[2] || ids[0] == ids[1] {
		t.Fatalf("ids = %v, want the first and third game to share an id", ids)
	}
	game, err := lib.Get(ctx, ids[0])
	if err != nil {
		t.Fatal(err)
	}
	if game.Title != "Alpha again" {
		t.Errorf("title = %q, want the last duplicate", game.Title)
	}
	if page, _ := lib.List(ctx, Filter{}); totalOf(page) != 2 {
		t.Errorf("library total = %d, want 2 rows", totalOf(page))
	}
}

func TestUpsertManyEmptyBatchIsNoOp(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ids, err := lib.UpsertMany(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Errorf("ids = %v, want none", ids)
	}
}

func TestUpsertManyRollsBackTheWholeBatchOnError(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	broken := batchGame("bad", "Broken")
	broken.Extra = map[string]any{"callback": func() {}}

	if _, err := lib.UpsertMany(ctx, []contract.GameDTO{batchGame("ok", "Fine"), broken}); err == nil {
		t.Fatal("a game that cannot be encoded must fail the batch")
	}
	if page, _ := lib.List(ctx, Filter{}); totalOf(page) != 0 {
		t.Errorf("library total = %d, want 0: earlier games of a failed batch must roll back", totalOf(page))
	}
}

func TestSaveBatchSkipsGamesWithTheSameUpdateStampAndVersion(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	first, err := lib.saveBatch(ctx, []contract.GameDTO{batchGame("a", "Alpha"), batchGame("b", "Beta")})
	if err != nil {
		t.Fatal(err)
	}
	if first.Written != 2 || first.Skipped != 0 {
		t.Fatalf("first batch = %+v, want 2 written", first)
	}

	moved := batchGame("b", "Beta changed")
	moved.UpdatedAt = "2026-02-01T00:00:00.000Z"
	again, err := lib.saveBatch(ctx, []contract.GameDTO{batchGame("a", "Alpha"), moved})
	if err != nil {
		t.Fatal(err)
	}
	if again.Written != 1 || again.Skipped != 1 {
		t.Errorf("repeat batch = %+v, want the same game skipped and the changed one written", again)
	}
	if again.IDs[1] != first.IDs[1] {
		t.Errorf("skipped and written games keep their ids: %v vs %v", again.IDs, first.IDs)
	}
}

func TestSaveBatchNeverSkipsGamesWithoutAnUpdateStamp(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	game := batchGame("a", "Alpha")
	game.UpdatedAt = ""
	if _, err := lib.saveBatch(ctx, []contract.GameDTO{game}); err != nil {
		t.Fatal(err)
	}
	again, err := lib.saveBatch(ctx, []contract.GameDTO{game})
	if err != nil {
		t.Fatal(err)
	}
	if again.Skipped != 0 || again.Written != 1 {
		t.Errorf("batch = %+v, want the game written again", again)
	}
}

func TestUpsertManyKeepsHiddenAndFavoriteFlags(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	ids, err := lib.UpsertMany(ctx, []contract.GameDTO{batchGame("a", "Alpha")})
	if err != nil {
		t.Fatal(err)
	}
	hidden, favorite := true, true
	if _, err := lib.Patch(ctx, ids[0], Patch{Hidden: &hidden, Favorite: &favorite}); err != nil {
		t.Fatal(err)
	}

	changed := batchGame("a", "Alpha changed")
	changed.UpdatedAt = "2026-02-01T00:00:00.000Z"
	if _, err := lib.UpsertMany(ctx, []contract.GameDTO{changed}); err != nil {
		t.Fatal(err)
	}
	game, err := lib.Get(ctx, ids[0])
	if err != nil {
		t.Fatal(err)
	}
	if game.Title != "Alpha changed" || game.Hidden != 1 || game.Favorite != 1 {
		t.Errorf("game = title %q hidden %d favorite %d, want the new title and both flags kept",
			game.Title, game.Hidden, game.Favorite)
	}
}

func TestUpsertManyKeepsStoredFieldsTheSourceDoesNotSend(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	full := batchGame("a", "Alpha")
	full.Developer = "Blue Fox"
	full.Tags = []string{"rpg"}
	if _, err := lib.UpsertMany(ctx, []contract.GameDTO{full}); err != nil {
		t.Fatal(err)
	}

	sparse := contract.GameDTO{Source: "f95", SourceID: "a", URL: "https://f95/a", Title: "Alpha", UpdatedAt: "2026-02-01T00:00:00.000Z"}
	ids, err := lib.UpsertMany(ctx, []contract.GameDTO{sparse})
	if err != nil {
		t.Fatal(err)
	}
	game, err := lib.Get(ctx, ids[0])
	if err != nil {
		t.Fatal(err)
	}
	if game.Developer == nil || *game.Developer != "Blue Fox" || len(game.Tags) != 1 || game.Tags[0] != "rpg" {
		t.Errorf("developer = %v, tags = %v; the stored values must stay", game.Developer, game.Tags)
	}
}

func TestSaveBatchReportsCoversThatChanged(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	withCover := batchGame("a", "Alpha")
	withCover.CoverURL = "https://f95/a.jpg"

	first, err := lib.saveBatch(ctx, []contract.GameDTO{withCover})
	if err != nil {
		t.Fatal(err)
	}
	if len(first.CoverIDs) != 1 || first.CoverIDs[0] != first.IDs[0] {
		t.Errorf("new cover ids = %v, want the new game", first.CoverIDs)
	}

	again, err := lib.saveBatch(ctx, []contract.GameDTO{withCover})
	if err != nil {
		t.Fatal(err)
	}
	if len(again.CoverIDs) != 0 {
		t.Errorf("unchanged cover ids = %v, want none", again.CoverIDs)
	}

	moved := withCover
	moved.UpdatedAt = "2026-02-01T00:00:00.000Z"
	moved.CoverURL = "https://f95/a-2.jpg"
	changed, err := lib.saveBatch(ctx, []contract.GameDTO{moved})
	if err != nil {
		t.Fatal(err)
	}
	if len(changed.CoverIDs) != 1 {
		t.Errorf("changed cover ids = %v, want the game", changed.CoverIDs)
	}
}

func TestUpsertManyCachesChangedCoversInTheBackground(t *testing.T) {
	server, _ := imageServer(t)
	notify, notices := gameNotices()
	lib := openTestLibrary(t, Options{Notify: notify})
	game := batchGame("a", "Alpha")
	game.CoverURL = server.URL + "/cover.jpg"

	ids, err := lib.UpsertMany(context.Background(), []contract.GameDTO{game})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-notices:
	case <-time.After(5 * time.Second):
		t.Fatal("the cover of a changed game was not cached")
	}
	stored, err := lib.Get(context.Background(), ids[0])
	if err != nil {
		t.Fatal(err)
	}
	if stored.CoverPath == nil {
		t.Error("cover path not stored")
	}
}
