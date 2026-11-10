package library

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"light-launcher/core/internal/contract"
)

func TestGetReportsRunningState(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	game := upsertGame(t, lib, "run-1", "Harbor")

	if !lib.launches.begin(game.ID) {
		t.Fatal("begin should succeed for a free game")
	}
	got, err := lib.Get(context.Background(), game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Running {
		t.Error("a game with an active launch must report running")
	}
	lib.launches.end(game.ID)
	got, _ = lib.Get(context.Background(), game.ID)
	if got.Running {
		t.Error("a game that exited must not report running")
	}
}

func TestDescriptionTextIsPlainAndSafe(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	id, err := lib.Upsert(ctx, contract.GameDTO{
		Source: "manual", SourceID: "desc-1", URL: "file://x", Title: "Desc",
		DescriptionHTML: "<p>Hello&nbsp;<b>there</b></p><script>alert(1)</script><br>Bye &amp; co",
	})
	if err != nil {
		t.Fatal(err)
	}
	game, err := lib.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if game.Description == nil || *game.Description != "<p>Hello&nbsp;<b>there</b></p><script>alert(1)</script><br>Bye &amp; co" {
		t.Errorf("description must keep the html, got %v", game.Description)
	}
	want := "Hello there\n\nBye & co"
	if game.DescriptionText == nil || *game.DescriptionText != want {
		t.Errorf("descriptionText = %q, want %q", deref(game.DescriptionText), want)
	}
}

func TestCoverLocalAndCoverFile(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	game := upsertGame(t, lib, "cover-1", "Covered")
	if game.CoverLocal != nil || !isMissingCover(lib, game.ID) {
		t.Fatal("a game without a cached cover must not report one")
	}

	cover := filepath.Join(lib.opts.MediaDir, "covers", "abc.jpg")
	if err := os.MkdirAll(filepath.Dir(cover), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cover, []byte("jpg"), 0o644); err != nil {
		t.Fatal(err)
	}
	updated, err := lib.Patch(ctx, game.ID, Patch{CoverPath: &cover})
	if err != nil {
		t.Fatal(err)
	}
	if updated.CoverLocal == nil || *updated.CoverLocal != "covers/abc.jpg" {
		t.Errorf("coverLocal = %v, want covers/abc.jpg", deref(updated.CoverLocal))
	}
	path, ok := lib.CoverFile(game.ID)
	if !ok || path != cover {
		t.Errorf("CoverFile = %q, %v; want %q, true", path, ok, cover)
	}

	outside := filepath.Join(t.TempDir(), "elsewhere.jpg")
	if err := os.WriteFile(outside, []byte("jpg"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := lib.Patch(ctx, game.ID, Patch{CoverPath: &outside}); err != nil {
		t.Fatal(err)
	}
	if _, ok := lib.CoverFile(game.ID); ok {
		t.Error("a cover outside the media folder must not be served")
	}
}

func isMissingCover(lib *Library, id int64) bool {
	_, ok := lib.CoverFile(id)
	return !ok
}

func deref(value *string) string {
	if value == nil {
		return "<nil>"
	}
	return *value
}
