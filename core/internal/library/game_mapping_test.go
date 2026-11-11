package library

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"light-launcher/core/internal/contract"
)

// sdkGameJSON has every field of the plugin SDK GameDTO, as a plugin sends it over the wire.
const sdkGameJSON = `{
	"source": "f95",
	"sourceId": "4242",
	"url": "https://f95.example/threads/4242/",
	"title": "Harbor Lights",
	"version": "0.9.1",
	"developer": "Studio Pine",
	"engine": "Ren'Py",
	"tags": ["adventure", "pixel art"],
	"coverUrl": "https://f95.example/covers/4242.jpg",
	"screenshots": ["https://f95.example/shots/1.png", "https://f95.example/shots/2.png"],
	"descriptionHtml": "<p>A quiet harbor.</p>",
	"changelog": "Fixed the ferry.",
	"rating": 4.5,
	"status": "ongoing",
	"updatedAt": "2026-05-01T10:00:00Z",
	"extra": {"threadId": "4242", "language": "English"}
}`

// decodeSDKGame decodes the SDK shape strictly, so a field the Go side does not know fails the test.
func decodeSDKGame(t *testing.T, raw string) contract.GameDTO {
	t.Helper()
	var dto contract.GameDTO
	decoder := json.NewDecoder(bytes.NewReader([]byte(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&dto); err != nil {
		t.Fatalf("decode SDK game: %v", err)
	}
	return dto
}

func TestSDKGameFieldsAreStored(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	dto := decodeSDKGame(t, sdkGameJSON)

	id, err := lib.Upsert(ctx, dto)
	if err != nil {
		t.Fatal(err)
	}
	game, err := lib.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	checkString(t, "version", game.Version, "0.9.1")
	checkString(t, "developer", game.Developer, "Studio Pine")
	checkString(t, "engine", game.Engine, "Ren'Py")
	checkString(t, "coverUrl", game.CoverURL, "https://f95.example/covers/4242.jpg")
	checkString(t, "description", game.Description, "<p>A quiet harbor.</p>")
	checkString(t, "changelog", game.Changelog, "Fixed the ferry.")
	checkString(t, "sourceUpdatedAt", game.SourceUpdatedAt, "2026-05-01T10:00:00Z")
	checkString(t, "status", &game.Status, "ongoing")
	if game.Rating == nil || *game.Rating != 4.5 {
		t.Errorf("rating = %v, want 4.5", game.Rating)
	}
	if len(game.Tags) != 2 || game.Tags[1] != "pixel art" {
		t.Errorf("tags = %v", game.Tags)
	}
	if len(game.Screenshots) != 2 || game.Screenshots[0] != "https://f95.example/shots/1.png" {
		t.Errorf("screenshots = %v", game.Screenshots)
	}
	if game.Extra["language"] != "English" {
		t.Errorf("extra = %v", game.Extra)
	}
}

func TestUpsertKeepsRatingWhenSourceSendsNone(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	dto := decodeSDKGame(t, sdkGameJSON)
	if _, err := lib.Upsert(ctx, dto); err != nil {
		t.Fatal(err)
	}

	dto.Rating = nil
	id, err := lib.Upsert(ctx, dto)
	if err != nil {
		t.Fatal(err)
	}
	game, err := lib.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if game.Rating == nil || *game.Rating != 4.5 {
		t.Errorf("rating = %v, want the stored 4.5", game.Rating)
	}
}

func TestSDKGameWithNullFieldsIsAccepted(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	raw := `{"source":"itch","sourceId":"9","url":"https://itch.example/9","title":"Nulls",
		"version":null,"developer":null,"engine":null,"coverUrl":null,"descriptionHtml":null,
		"changelog":null,"rating":null,"updatedAt":null,"extra":null}`

	id, err := lib.Upsert(context.Background(), decodeSDKGame(t, raw))
	if err != nil {
		t.Fatal(err)
	}
	game, err := lib.Get(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if game.Rating != nil || game.CoverURL != nil || game.Version != nil {
		t.Errorf("null fields must stay empty, got rating=%v coverUrl=%v version=%v", game.Rating, game.CoverURL, game.Version)
	}
}

func checkString(t *testing.T, name string, got *string, want string) {
	t.Helper()
	if got == nil || *got != want {
		t.Errorf("%s = %v, want %q", name, deref(got), want)
	}
}
