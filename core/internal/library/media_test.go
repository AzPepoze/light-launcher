package library

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"light-launcher/core/internal/contract"
)

func TestCacheImageDownloadsOnceAndKeepsBytes(t *testing.T) {
	var hits atomic.Int32
	payload := []byte("\x89PNG\r\n\x1a\nfake-png-bytes")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "image/png")
		w.Write(payload)
	}))
	defer server.Close()

	media := newMediaCache(t.TempDir())
	url := server.URL + "/cover"
	first, err := media.cacheImage(context.Background(), url, "covers")
	if err != nil {
		t.Fatal(err)
	}
	second, err := media.cacheImage(context.Background(), url, "covers")
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Errorf("paths differ: %s and %s", first, second)
	}
	if hits.Load() != 1 {
		t.Errorf("server hits = %d, want 1", hits.Load())
	}
	if !strings.HasSuffix(first, ".png") {
		t.Errorf("path %s should end in .png", first)
	}
	stored, err := os.ReadFile(first)
	if err != nil {
		t.Fatal(err)
	}
	if string(stored) != string(payload) {
		t.Error("stored bytes were changed")
	}
}

func TestCacheImageRejectsNonImages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html>login</html>"))
	}))
	defer server.Close()

	media := newMediaCache(t.TempDir())
	if _, err := media.cacheImage(context.Background(), server.URL+"/page", "covers"); err == nil {
		t.Error("an HTML page must not be cached as an image")
	}
}

func TestCacheImageRejectsNonHTTPURLs(t *testing.T) {
	media := newMediaCache(t.TempDir())
	if _, err := media.cacheImage(context.Background(), "file:///etc/passwd", "covers"); err == nil {
		t.Error("file URLs must be rejected")
	}
}

func TestCachedScreenshotsGetLocalPathsAndFiles(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte("png-bytes"))
	}))
	defer server.Close()

	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	id, err := lib.Upsert(ctx, contract.GameDTO{
		Source: "manual", SourceID: "shots-1", URL: "file://shots-1", Title: "Shots",
		Screenshots: []string{server.URL + "/1.png", server.URL + "/2.png"},
	})
	if err != nil {
		t.Fatal(err)
	}
	game, err := lib.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	lib.CacheScreenshots(ctx, game)

	game, err = lib.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if len(game.ScreenshotsLocal) != 2 {
		t.Fatalf("screenshotsLocal = %v, want two entries", game.ScreenshotsLocal)
	}
	for i, local := range game.ScreenshotsLocal {
		if local == nil || !strings.HasPrefix(*local, "screenshots/") {
			t.Errorf("screenshotsLocal[%d] = %v, want a path in screenshots/", i, local)
		}
	}
	for i := range game.Screenshots {
		if _, ok := lib.ScreenshotFile(id, i); !ok {
			t.Errorf("ScreenshotFile(%d) = missing, want the cached file", i)
		}
	}
	if _, ok := lib.ScreenshotFile(id, 2); ok {
		t.Error("an index past the last screenshot must not resolve")
	}
}

func TestUncachedScreenshotHasNoLocalPath(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	id, err := lib.Upsert(ctx, contract.GameDTO{
		Source: "manual", SourceID: "shots-2", URL: "file://shots-2", Title: "No shots",
		Screenshots: []string{"https://example.test/not-downloaded.png"},
	})
	if err != nil {
		t.Fatal(err)
	}
	game, err := lib.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if len(game.ScreenshotsLocal) != 1 || game.ScreenshotsLocal[0] != nil {
		t.Errorf("screenshotsLocal = %v, want one nil entry", game.ScreenshotsLocal)
	}
	if _, ok := lib.ScreenshotFile(id, 0); ok {
		t.Error("an uncached screenshot must not resolve")
	}
}

func TestAttachCoverStoresLocalPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write([]byte("jpeg-bytes"))
	}))
	defer server.Close()

	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	game := upsertGame(t, lib, "cov", "Cover Game")
	coverURL := server.URL + "/c.jpg"
	if _, err := lib.Patch(ctx, game.ID, Patch{CoverURL: &coverURL}); err != nil {
		t.Fatal(err)
	}
	game, _ = lib.Get(ctx, game.ID)

	updated, err := lib.AttachCover(ctx, game)
	if err != nil {
		t.Fatal(err)
	}
	if updated.CoverPath == nil || filepath.Dir(*updated.CoverPath) != filepath.Join(lib.opts.MediaDir, "covers") {
		t.Errorf("cover path = %v, want a file in the covers folder", updated.CoverPath)
	}
}
