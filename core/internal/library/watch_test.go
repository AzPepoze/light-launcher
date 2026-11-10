package library

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// recordingSink remembers what the watcher handed over.
type recordingSink struct {
	gameIDs []int64
	paths   []string
}

func (s *recordingSink) Accept(_ context.Context, gameID int64, path string) error {
	s.gameIDs = append(s.gameIDs, gameID)
	s.paths = append(s.paths, path)
	return nil
}

type watchFixture struct {
	lib       *Library
	sink      *recordingSink
	watchDir  string
	downloads string
}

func newWatchFixture(t *testing.T) watchFixture {
	t.Helper()
	fx := watchFixture{sink: &recordingSink{}, watchDir: t.TempDir(), downloads: t.TempDir()}
	fx.lib = openTestLibrary(t, Options{
		WatchFolder:  func() string { return fx.watchDir },
		DownloadsDir: fx.downloads,
		Sink:         fx.sink,
	})
	return fx
}

func (fx watchFixture) drop(t *testing.T, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(fx.watchDir, name), []byte("archive bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestWatchImportsStableMatchAfterTwoPolls(t *testing.T) {
	fx := newWatchFixture(t)
	harbor := upsertGame(t, fx.lib, "h", "Moonlit Harbor")
	fx.drop(t, "Moonlit-Harbor-v1.1-PC.zip")
	ctx := context.Background()

	first, err := fx.lib.ScanWatchFolder(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if first != 0 {
		t.Errorf("first poll imported %d, want 0 while the size is still new", first)
	}

	second, err := fx.lib.ScanWatchFolder(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if second != 1 {
		t.Fatalf("second poll imported %d, want 1", second)
	}
	if len(fx.sink.gameIDs) != 1 || fx.sink.gameIDs[0] != harbor.ID {
		t.Errorf("sink got %v, want game %d", fx.sink.gameIDs, harbor.ID)
	}
	if _, err := os.Stat(filepath.Join(fx.downloads, "Moonlit-Harbor-v1.1-PC.zip")); err != nil {
		t.Errorf("file was not moved to downloads: %v", err)
	}
	if _, err := os.Stat(filepath.Join(fx.watchDir, "Moonlit-Harbor-v1.1-PC.zip")); err == nil {
		t.Error("file is still in the watch folder")
	}
}

func TestWatchLeavesUnmatchedFilesAlone(t *testing.T) {
	fx := newWatchFixture(t)
	upsertGame(t, fx.lib, "h", "Moonlit Harbor")
	fx.drop(t, "readme.txt")
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		if _, err := fx.lib.ScanWatchFolder(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if len(fx.sink.gameIDs) != 0 {
		t.Errorf("unmatched file was imported: %v", fx.sink.gameIDs)
	}
}

func TestIncomingListsSuggestionsAndIgnoreHidesFile(t *testing.T) {
	fx := newWatchFixture(t)
	upsertGame(t, fx.lib, "h", "Moonlit Harbor")
	fx.drop(t, "Moonlit Harbor patch v1.zip")
	ctx := context.Background()

	report, err := fx.lib.Incoming(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Configured || len(report.Items) != 1 {
		t.Fatalf("report = %+v", report)
	}
	item := report.Items[0]
	if item.Auto || len(item.Suggestions) == 0 || item.Suggestions[0].Title != "Moonlit Harbor" {
		t.Errorf("item = %+v, want a suggestion and no auto import", item)
	}
	if report.Count != 1 {
		t.Errorf("count = %d, want 1", report.Count)
	}

	if err := fx.lib.IncomingIgnore(ctx, "Moonlit Harbor patch v1.zip"); err != nil {
		t.Fatal(err)
	}
	report, _ = fx.lib.Incoming(ctx)
	if report.Count != 0 || !report.Items[0].Ignored {
		t.Errorf("after ignore: count=%d ignored=%v", report.Count, report.Items[0].Ignored)
	}
}

func TestIncomingRejectsPathsOutsideWatchFolder(t *testing.T) {
	fx := newWatchFixture(t)
	if err := fx.lib.IncomingIgnore(context.Background(), "../secret.zip"); err == nil {
		t.Error("a path outside the watch folder must be rejected")
	}
}

func TestIncomingCreateLocalMakesManualGame(t *testing.T) {
	fx := newWatchFixture(t)
	game, err := fx.lib.IncomingCreateLocal(context.Background(), "Ember_Vow_v0.5.zip")
	if err != nil {
		t.Fatal(err)
	}
	if game.Source != "manual" || game.Title != "Ember Vow v0.5" || game.Status != "completed" {
		t.Errorf("game = %+v", game)
	}
}

func TestWatchWithoutFolderDoesNothing(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	imported, err := lib.ScanWatchFolder(context.Background())
	if err != nil || imported != 0 {
		t.Errorf("imported=%d err=%v, want 0 and nil", imported, err)
	}
	report, _ := lib.Incoming(context.Background())
	if report.Configured {
		t.Error("report should not be configured without a folder")
	}
}
