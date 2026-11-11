package library

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"light-launcher/core/internal/contract"
)

// imageServer serves a small image for every path and counts the requests.
func imageServer(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write([]byte("jpeg-bytes"))
	}))
	t.Cleanup(server.Close)
	return server, &hits
}

// gameNotices returns a Notify hook and a channel that receives the id of every library.game event.
func gameNotices() (func(string, any), chan any) {
	notices := make(chan any, 16)
	notify := func(channel string, data any) {
		if channel == "library.game" {
			notices <- data
		}
	}
	return notify, notices
}

func TestUpsertCachesCoverInBackground(t *testing.T) {
	server, _ := imageServer(t)
	notify, notices := gameNotices()
	lib := openTestLibrary(t, Options{Notify: notify})
	ctx := context.Background()

	id, err := lib.Upsert(ctx, contract.GameDTO{
		Source: "itch", SourceID: "bg-1", URL: "https://itch.example/bg-1", Title: "Background",
		CoverURL:    server.URL + "/cover.jpg",
		Screenshots: []string{server.URL + "/shot-1.jpg"},
	})
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-notices:
	case <-time.After(5 * time.Second):
		t.Fatal("no library.game notice after the cover was cached")
	}
	game, err := lib.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if game.CoverPath == nil {
		t.Fatal("cover path was not stored")
	}
	if _, err := os.Stat(*game.CoverPath); err != nil {
		t.Errorf("cover file missing: %v", err)
	}
	if game.CoverLocal == nil || !strings.HasPrefix(*game.CoverLocal, "covers/") {
		t.Errorf("coverLocal = %v, want a path under covers/", game.CoverLocal)
	}
}

func TestUpsertWithoutCoverDoesNotNotify(t *testing.T) {
	notify, notices := gameNotices()
	lib := openTestLibrary(t, Options{Notify: notify})

	upsertGame(t, lib, "manual-1", "No Cover")

	select {
	case data := <-notices:
		t.Errorf("unexpected library.game notice: %v", data)
	case <-time.After(200 * time.Millisecond):
	}
}

func TestUpsertSkipsCoverAlreadyStored(t *testing.T) {
	server, hits := imageServer(t)
	notify, notices := gameNotices()
	lib := openTestLibrary(t, Options{Notify: notify})
	dto := contract.GameDTO{Source: "itch", SourceID: "again", URL: "https://itch.example/again", Title: "Again", CoverURL: server.URL + "/c.jpg"}

	if _, err := lib.Upsert(context.Background(), dto); err != nil {
		t.Fatal(err)
	}
	<-notices
	if _, err := lib.Upsert(context.Background(), dto); err != nil {
		t.Fatal(err)
	}
	select {
	case <-notices:
		t.Error("a second sync with a stored cover must not notify again")
	case <-time.After(200 * time.Millisecond):
	}
	if hits.Load() != 1 {
		t.Errorf("image server hits = %d, want 1", hits.Load())
	}
}

func TestAddFromURLCachesCoverFromPlugin(t *testing.T) {
	server, _ := imageServer(t)
	notify, notices := gameNotices()
	const link = "https://itch.example/add-1"
	sources := &fakeSources{
		infos: []contract.SourceInfo{{Plugin: "itch", SourceID: "itch", Label: "itch.io"}},
		replies: map[string]func(any) (any, error){
			methodParseRef: func(any) (any, error) {
				return ParsedRef{Kind: "game", ID: "add-1", URL: link}, nil
			},
			methodFetchGame: func(any) (any, error) {
				return fetchGameReply{Game: contract.GameDTO{
					Source: "itch", SourceID: "add-1", URL: link, Title: "Added",
					CoverURL: server.URL + "/add.jpg",
				}}, nil
			},
		},
	}
	lib := openTestLibrary(t, Options{Sources: sources, Notify: notify})

	game, err := lib.AddFromURL(context.Background(), link)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-notices:
	case <-time.After(5 * time.Second):
		t.Fatal("adding a game from a link must cache its cover")
	}
	cached, err := lib.Get(context.Background(), game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if cached.CoverLocal == nil {
		t.Error("coverLocal should be set after the background download")
	}
}

func TestStopWaitsForRunningArtwork(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	started := make(chan struct{})
	done := make(chan struct{})
	lib.artwork.start(func(ctx context.Context) {
		close(started)
		<-ctx.Done()
		close(done)
	})
	<-started
	if err := lib.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	default:
		t.Error("Close must cancel and wait for running artwork jobs")
	}
}
