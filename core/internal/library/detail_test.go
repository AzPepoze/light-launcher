package library

import (
	"context"
	"testing"

	"light-launcher/core/internal/contract"
)

func TestDetailIncludesFilesAndChanges(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	game := upsertGame(t, lib, "detail-1", "Detail")
	candidates := []contract.FileCandidate{{Label: "Windows build", URL: "https://example.com/a.zip", Kind: "direct"}}
	if err := lib.ReplaceFiles(ctx, game.ID, candidates); err != nil {
		t.Fatal(err)
	}
	if err := lib.AddCheck(ctx, game.ID, "0.1", "0.2", "fixes"); err != nil {
		t.Fatal(err)
	}

	detail, err := lib.Detail(ctx, game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Title != "Detail" || detail.ID != game.ID {
		t.Errorf("detail must carry the game fields, got %+v", detail.Game)
	}
	if len(detail.Files) != 1 || detail.Files[0].Label != "Windows build" || detail.Files[0].GameID != game.ID {
		t.Errorf("files = %+v, want the stored candidate", detail.Files)
	}
	if len(detail.Changes) != 1 || detail.Changes[0].NewVersion == nil || *detail.Changes[0].NewVersion != "0.2" {
		t.Errorf("changes = %+v, want the 0.1 to 0.2 check", detail.Changes)
	}
}

func TestDetailMissingGameIsNotFound(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	if _, err := lib.Detail(context.Background(), 999); err != ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
