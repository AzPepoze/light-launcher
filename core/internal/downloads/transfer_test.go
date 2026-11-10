package downloads

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"
	"time"
)

func newTestTransferer(retries int) *transferer {
	return &transferer{
		client:  &http.Client{},
		limiter: &RateLimiter{},
		hosts:   newHostGates(8),
		retries: retries,
	}
}

// startJob runs one transfer in the background. The returned job's done channel closes when it ends.
func startJob(tr *transferer, row Download, resume []Segment) (*job, <-chan error) {
	j := newJob(row, resume)
	ctx, cancel := context.WithCancel(context.Background())
	j.cancel = cancel
	errs := make(chan error, 1)
	go func() {
		err := tr.run(ctx, j)
		cancel()
		close(j.done)
		errs <- err
	}()
	return j, errs
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for condition")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func newTestRow(t *testing.T, url string, connections int) Download {
	return Download{
		URL:         url,
		Filename:    "file.bin",
		Dest:        filepath.Join(t.TempDir(), "file.bin"),
		Connections: connections,
	}
}

func mustMatchFile(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("file content differs: got %d bytes, want %d", len(got), len(want))
	}
}

func TestRangedTransferWritesExactFile(t *testing.T) {
	file := &testFile{data: patternData(1 << 20), etag: `"v1"`}
	srv := httptest.NewServer(file)
	defer srv.Close()

	row := newTestRow(t, srv.URL+"/game.bin", 4)
	j, errs := startJob(newTestTransferer(3), row, nil)
	if err := <-errs; err != nil {
		t.Fatal(err)
	}
	mustMatchFile(t, row.Dest, file.data)
	if got := len(j.snapshotSegments()); got != 4 {
		t.Fatalf("segments = %d, want 4", got)
	}
	if _, err := os.Stat(row.Dest + ".part"); !os.IsNotExist(err) {
		t.Fatal("part file should be renamed away")
	}
}

func TestStreamTransferWhenServerHasNoRanges(t *testing.T) {
	file := &testFile{data: patternData(300_000), noRanges: true}
	srv := httptest.NewServer(file)
	defer srv.Close()

	row := newTestRow(t, srv.URL+"/plain.bin", 4)
	j, errs := startJob(newTestTransferer(3), row, nil)
	if err := <-errs; err != nil {
		t.Fatal(err)
	}
	mustMatchFile(t, row.Dest, file.data)
	if got := len(j.snapshotSegments()); got != 1 {
		t.Fatalf("segments = %d, want 1 for a stream", got)
	}
}

func TestRetriesServerErrorsThenSucceeds(t *testing.T) {
	file := &testFile{data: patternData(200_000), etag: `"v1"`, failGets: 2}
	srv := httptest.NewServer(file)
	defer srv.Close()

	row := newTestRow(t, srv.URL+"/flaky.bin", 1)
	_, errs := startJob(newTestTransferer(3), row, nil)
	if err := <-errs; err != nil {
		t.Fatal(err)
	}
	mustMatchFile(t, row.Dest, file.data)
	if file.failures != 2 {
		t.Fatalf("injected failures used = %d, want 2", file.failures)
	}
}

func TestGivesUpAfterRetryLimit(t *testing.T) {
	file := &testFile{data: patternData(50_000), etag: `"v1"`, failGets: 100}
	srv := httptest.NewServer(file)
	defer srv.Close()

	row := newTestRow(t, srv.URL+"/broken.bin", 1)
	_, errs := startJob(newTestTransferer(1), row, nil)
	if err := <-errs; err == nil {
		t.Fatal("expected an error after the retry limit")
	}
}

func TestRateLimitedRequestWaitsForRetryAfter(t *testing.T) {
	file := &testFile{data: patternData(64_000), etag: `"v1"`, limitGets: 1}
	srv := httptest.NewServer(file)
	defer srv.Close()

	row := newTestRow(t, srv.URL+"/limited.bin", 1)
	started := time.Now()
	_, errs := startJob(newTestTransferer(0), row, nil)
	if err := <-errs; err != nil {
		t.Fatal(err)
	}
	if waited := time.Since(started); waited < 900*time.Millisecond {
		t.Fatalf("returned after %v, want to wait for Retry-After", waited)
	}
	mustMatchFile(t, row.Dest, file.data)
}

func TestResumesFromBytesWrittenAfterConnectionDrop(t *testing.T) {
	file := &testFile{data: patternData(400_000), etag: `"v1"`, dropAfter: 100_000}
	srv := httptest.NewServer(file)
	defer srv.Close()

	row := newTestRow(t, srv.URL+"/drop.bin", 1)
	_, errs := startJob(newTestTransferer(3), row, nil)
	if err := <-errs; err != nil {
		t.Fatal(err)
	}
	mustMatchFile(t, row.Dest, file.data)

	ranges := file.rangeLog()
	if len(ranges) < 2 || ranges[0] != "bytes=0-399999" {
		t.Fatalf("unexpected range log %v", ranges)
	}
	if ranges[1] == "bytes=0-399999" {
		t.Fatal("retry restarted at byte 0 instead of resuming")
	}
}

func TestSpeedLimitSlowsTransfer(t *testing.T) {
	file := &testFile{data: patternData(256 << 10), etag: `"v1"`}
	srv := httptest.NewServer(file)
	defer srv.Close()

	tr := newTestTransferer(0)
	tr.limiter.SetRate(512 << 10)
	row := newTestRow(t, srv.URL+"/throttled.bin", 1)
	started := time.Now()
	_, errs := startJob(tr, row, nil)
	if err := <-errs; err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(started); elapsed < 350*time.Millisecond {
		t.Fatalf("256 KiB at 512 KiB/s finished in %v, limit not applied", elapsed)
	}
	mustMatchFile(t, row.Dest, file.data)
}

func TestPauseKeepsPartAndResumeFinishes(t *testing.T) {
	file := &testFile{data: patternData(2 << 20), etag: `"v1"`, slow: true}
	srv := httptest.NewServer(file)
	defer srv.Close()

	tr := newTestTransferer(0)
	row := newTestRow(t, srv.URL+"/paused.bin", 2)
	first, firstErrs := startJob(tr, row, nil)
	waitFor(t, func() bool { return first.progress(time.Now()).received > 0 })
	first.requestStop(stopPause)
	if err := <-firstErrs; err == nil {
		t.Fatal("paused run should report the stop")
	}
	saved := first.snapshotSegments()
	var savedBytes int64
	for _, seg := range saved {
		savedBytes += seg.Done
	}
	if savedBytes == 0 || savedBytes >= int64(len(file.data)) {
		t.Fatalf("saved %d bytes, want a partial transfer", savedBytes)
	}

	row.ETag = `"v1"`
	file.setSlow(false)
	before := len(file.rangeLog())
	_, errs := startJob(tr, row, saved)
	if err := <-errs; err != nil {
		t.Fatal(err)
	}
	mustMatchFile(t, row.Dest, file.data)

	resumed := file.rangeLog()[before:]
	for _, seg := range saved {
		if remaining(seg) == 0 {
			continue
		}
		want := fmt.Sprintf("bytes=%d-%d", seg.Start+seg.Done, seg.End-1)
		if !slices.Contains(resumed, want) {
			t.Fatalf("resumed run did not request %s; requests were %v", want, resumed)
		}
	}
}

func TestChangedETagRestartsDownload(t *testing.T) {
	file := &testFile{data: patternData(1 << 20), etag: `"new"`, slow: true}
	srv := httptest.NewServer(file)
	defer srv.Close()

	tr := newTestTransferer(0)
	row := newTestRow(t, srv.URL+"/changed.bin", 2)
	first, firstErrs := startJob(tr, row, nil)
	waitFor(t, func() bool { return first.progress(time.Now()).received > 0 })
	first.requestStop(stopPause)
	<-firstErrs
	saved := first.snapshotSegments()

	row.ETag = `"old"`
	file.setSlow(false)
	before := len(file.rangeLog())
	_, errs := startJob(tr, row, saved)
	if err := <-errs; err != nil {
		t.Fatal(err)
	}
	mustMatchFile(t, row.Dest, file.data)
	restarted := file.rangeLog()[before:]
	want := "bytes=0-" + strconv.FormatInt(saved[0].End-1, 10)
	if !slices.Contains(restarted, want) {
		t.Fatalf("a changed file must restart at byte 0; wanted %s in %v", want, restarted)
	}
}
