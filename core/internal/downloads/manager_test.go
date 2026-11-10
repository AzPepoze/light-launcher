package downloads

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"light-launcher/core/internal/contract"
)

func newTestManager(t *testing.T, store Store, dir string, settings Settings) *Manager {
	t.Helper()
	if store == nil {
		store = NewMemoryStore()
	}
	m, err := NewManager(Options{Store: store, Dir: dir, Settings: settings})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		m.Shutdown(ctx)
	})
	return m
}

func waitForState(t *testing.T, m *Manager, id int64, want State) View {
	t.Helper()
	var v View
	waitFor(t, func() bool {
		var err error
		v, err = m.Get(context.Background(), id)
		return err == nil && v.State == want
	})
	return v
}

func TestPauseKeepsProgressAndResumeFinishes(t *testing.T) {
	file := &testFile{data: patternData(2 << 20), etag: `"v1"`, slow: true}
	srv := httptest.NewServer(file)
	defer srv.Close()

	m := newTestManager(t, nil, t.TempDir(), Settings{Connections: 2, MaxActive: 1, HostConnections: 8, Retries: 1})
	id, err := m.Enqueue(context.Background(), contract.DownloadRequest{URL: srv.URL + "/pause.bin", Filename: "pause.bin"})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool {
		v, _ := m.Get(context.Background(), id)
		return v.Received > 0
	})
	if err := m.Pause(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	paused := waitForState(t, m, id, StatePaused)
	if paused.Received == 0 || paused.Received >= int64(len(file.data)) {
		t.Fatalf("paused with %d bytes, want a partial download", paused.Received)
	}

	file.setSlow(false)
	if err := m.Resume(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	done := waitForState(t, m, id, StateDone)
	if done.Total != int64(len(file.data)) {
		t.Fatalf("total = %d", done.Total)
	}
	mustMatchFile(t, done.Dest, file.data)
}

func TestRestartResumesSavedSegments(t *testing.T) {
	file := &testFile{data: patternData(2 << 20), etag: `"v1"`, slow: true}
	srv := httptest.NewServer(file)
	defer srv.Close()

	store := NewMemoryStore()
	dir := t.TempDir()
	first := newTestManager(t, store, dir, Settings{Connections: 2, MaxActive: 1, HostConnections: 8, Retries: 1})
	id, err := first.Enqueue(context.Background(), contract.DownloadRequest{URL: srv.URL + "/restart.bin", Filename: "restart.bin"})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool {
		v, _ := first.Get(context.Background(), id)
		return v.Received > 0
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := first.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	saved, err := store.Segments(context.Background(), id)
	if err != nil || len(saved) == 0 {
		t.Fatalf("segments not saved before shutdown: %v", err)
	}

	file.setSlow(false)
	before := len(file.rangeLog())
	second := newTestManager(t, store, dir, Settings{Connections: 2, MaxActive: 1, HostConnections: 8, Retries: 1})
	if err := second.Restore(context.Background()); err != nil {
		t.Fatal(err)
	}
	done := waitForState(t, second, id, StateDone)
	mustMatchFile(t, done.Dest, file.data)

	resumed := file.rangeLog()[before:]
	for _, seg := range saved {
		if remaining(seg) == 0 {
			continue
		}
		want := fmt.Sprintf("bytes=%d-%d", seg.Start+seg.Done, seg.End-1)
		if !slices.Contains(resumed, want) {
			t.Fatalf("restart did not resume at %s; requests %v", want, resumed)
		}
	}
}

func TestRestoreMarksMissingPartAsError(t *testing.T) {
	store := NewMemoryStore()
	dir := t.TempDir()
	id, _ := store.Insert(context.Background(), Download{
		URL:   "https://example.test/x.bin",
		Dest:  filepath.Join(dir, "x.bin"),
		State: StateDownloading,
	})

	m := newTestManager(t, store, dir, Settings{})
	if err := m.Restore(context.Background()); err != nil {
		t.Fatal(err)
	}
	v := waitForState(t, m, id, StateError)
	if v.Error != "partial file is missing" {
		t.Fatalf("error = %q", v.Error)
	}
}

func TestCancelThenRetryFinishes(t *testing.T) {
	file := &testFile{data: patternData(1 << 20), etag: `"v1"`, slow: true}
	srv := httptest.NewServer(file)
	defer srv.Close()

	m := newTestManager(t, nil, t.TempDir(), Settings{Connections: 1, MaxActive: 1, HostConnections: 8, Retries: 1})
	id, err := m.Enqueue(context.Background(), contract.DownloadRequest{URL: srv.URL + "/cancel.bin"})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool {
		v, _ := m.Get(context.Background(), id)
		return v.Received > 0
	})
	if err := m.Cancel(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	waitForState(t, m, id, StateCanceled)

	file.setSlow(false)
	if err := m.Retry(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	done := waitForState(t, m, id, StateDone)
	mustMatchFile(t, done.Dest, file.data)
}

func TestRemoveDeletesRowAndFiles(t *testing.T) {
	file := &testFile{data: patternData(20_000), etag: `"v1"`}
	srv := httptest.NewServer(file)
	defer srv.Close()

	m := newTestManager(t, nil, t.TempDir(), Settings{})
	id, err := m.Enqueue(context.Background(), contract.DownloadRequest{URL: srv.URL + "/remove.bin"})
	if err != nil {
		t.Fatal(err)
	}
	done := waitForState(t, m, id, StateDone)
	if err := m.Remove(context.Background(), id, true); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Get(context.Background(), id); !errors.Is(err, ErrNotFound) {
		t.Fatalf("removed row still found: %v", err)
	}
	if _, err := os.Stat(done.Dest); !os.IsNotExist(err) {
		t.Fatal("finished file should be deleted")
	}
}

func TestEnqueueRejectsNonWebLinks(t *testing.T) {
	m := newTestManager(t, nil, t.TempDir(), Settings{})
	for _, link := range []string{"ftp://example.test/a.zip", "file:///etc/passwd", "not a link"} {
		if _, err := m.Enqueue(context.Background(), contract.DownloadRequest{URL: link}); err == nil {
			t.Errorf("link %q should be rejected", link)
		}
	}
}

func TestEnqueueRejectsWebPageWithoutQueueing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html>check</html>"))
	}))
	defer srv.Close()

	store := NewMemoryStore()
	m := newTestManager(t, store, t.TempDir(), Settings{})
	_, err := m.Enqueue(context.Background(), contract.DownloadRequest{URL: srv.URL + "/page"})
	var webPage *WebPageError
	if !errors.As(err, &webPage) {
		t.Fatalf("want WebPageError, got %v", err)
	}
	rows, _ := store.List(context.Background())
	if len(rows) != 0 {
		t.Fatalf("a web page must not create a row, got %d", len(rows))
	}
}

func TestEnqueueAvoidsFileNameCollisions(t *testing.T) {
	file := &testFile{data: patternData(10_000), etag: `"v1"`}
	srv := httptest.NewServer(file)
	defer srv.Close()

	m := newTestManager(t, nil, t.TempDir(), Settings{})
	first, _ := m.Enqueue(context.Background(), contract.DownloadRequest{URL: srv.URL + "/game.zip"})
	second, _ := m.Enqueue(context.Background(), contract.DownloadRequest{URL: srv.URL + "/game.zip"})
	a := waitForState(t, m, first, StateDone)
	b := waitForState(t, m, second, StateDone)
	if a.Dest == b.Dest {
		t.Fatalf("both downloads saved to %s", a.Dest)
	}
	if filepath.Base(b.Dest) != "game (2).zip" {
		t.Fatalf("second name = %q, want game (2).zip", filepath.Base(b.Dest))
	}
}

func TestSettingsAreClampedOnSet(t *testing.T) {
	m := newTestManager(t, nil, t.TempDir(), Settings{})
	got := m.SetSettings(Settings{Connections: 99, MaxActive: 0, HostConnections: -1, Retries: 50, SpeedLimitKBps: -5})
	want := Settings{Connections: 32, MaxActive: 1, HostConnections: 1, Retries: 10, SpeedLimitKBps: 0}
	if got != want {
		t.Fatalf("settings = %+v, want %+v", got, want)
	}
	if s := m.SetSpeedLimit(256); s.SpeedLimitKBps != 256 {
		t.Fatalf("speed limit = %d", s.SpeedLimitKBps)
	}
}
