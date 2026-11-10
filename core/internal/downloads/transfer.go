package downloads

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// transferer moves bytes for one download run. It owns no state between runs.
type transferer struct {
	client  *http.Client // must not set a total timeout: bodies are long-lived
	limiter *RateLimiter
	hosts   *hostGates
	retries int
}

// run probes the file, prepares the part file, downloads it and renames it into place.
func (t *transferer) run(ctx context.Context, j *job) error {
	probe, err := t.probeFor(ctx, j)
	if err != nil {
		return err
	}
	ranged := probe.Ranges && probe.Size > 0
	connections := 1
	if ranged {
		connections = ClampConnections(j.row.Connections)
	}

	part := j.row.Dest + ".part"
	if err := os.MkdirAll(filepath.Dir(j.row.Dest), 0o755); err != nil {
		return err
	}
	file, segments, err := openPart(part, probe, connections, ranged, j)
	if err != nil {
		return err
	}
	defer file.Close()
	j.setLayout(probe.Size, probe.ETag, segments)

	if ranged {
		err = t.transferRanges(ctx, j, file)
	} else {
		err = t.transferStream(ctx, j, file)
	}
	if err != nil {
		return err
	}
	if err := verifyComplete(j, ranged); err != nil {
		return err
	}
	// a stop that landed after the last byte keeps the part file for the next run
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(part, j.row.Dest)
}

func (t *transferer) probeFor(ctx context.Context, j *job) (ProbeResult, error) {
	if j.probe != nil {
		probe := *j.probe
		j.probe = nil
		return probe, nil
	}
	return Probe(ctx, t.client, j.row.URL, j.row.Headers)
}

// openPart reuses the part file when the saved segments still match this file. Otherwise it starts over.
func openPart(part string, probe ProbeResult, connections int, ranged bool, j *job) (*os.File, []Segment, error) {
	if ranged && probe.ETag == j.row.ETag {
		if file, segments, ok := reusePart(part, probe.Size, j.resume); ok {
			return file, segments, nil
		}
	}
	file, err := os.OpenFile(part, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, nil, err
	}
	if !ranged {
		return file, []Segment{{Index: 0}}, nil
	}
	if err := file.Truncate(probe.Size); err != nil {
		file.Close()
		return nil, nil, err
	}
	return file, planSegments(probe.Size, connections), nil
}

// reusePart opens an existing part file if its saved layout and size still fit.
func reusePart(part string, size int64, saved []Segment) (*os.File, []Segment, bool) {
	segments, ok := reuseSegments(size, saved)
	if !ok {
		return nil, nil, false
	}
	file, err := os.OpenFile(part, os.O_RDWR, 0o644)
	if err != nil {
		return nil, nil, false
	}
	info, err := file.Stat()
	if err != nil || info.Size() != size {
		file.Close()
		return nil, nil, false
	}
	return file, segments, true
}

// transferRanges downloads every pending segment in parallel. The first permanent error stops the rest.
func (t *transferer) transferRanges(ctx context.Context, j *job, file *os.File) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	var (
		wg       sync.WaitGroup
		once     sync.Once
		firstErr error
	)
	for _, idx := range j.pendingSegments() {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if err := t.fetchSegment(ctx, j, file, idx); err != nil {
				once.Do(func() {
					firstErr = err
					cancel(err)
				})
			}
		}(idx)
	}
	wg.Wait()
	return firstErr
}

// transferStream downloads a file that has no ranges in one request. A failed attempt starts over.
func (t *transferer) transferStream(ctx context.Context, j *job, file *os.File) error {
	for attempt := 0; ; attempt++ {
		err := t.streamOnce(ctx, j, file)
		if err == nil {
			return nil
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		if attempt >= t.retries {
			return err
		}
		if err := sleepContext(ctx, retryDelay(attempt)); err != nil {
			return err
		}
		if err := file.Truncate(0); err != nil {
			return err
		}
		j.resetStream()
	}
}

func (t *transferer) streamOnce(ctx context.Context, j *job, file *os.File) error {
	gate := j.gateRef()
	if err := gate.Acquire(ctx); err != nil {
		return err
	}
	defer gate.Release()
	host := t.hosts.forURL(j.row.URL)
	if err := host.Acquire(ctx); err != nil {
		return err
	}
	defer host.Release()

	return t.fetch(ctx, request{url: j.row.URL, headers: j.row.Headers}, file, func(n int64) {
		j.addBytes(0, n)
	})
}

// verifyComplete checks that every byte arrived before the file is renamed.
func verifyComplete(j *job, ranged bool) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if !ranged {
		if j.size > 0 && j.received != j.size {
			return fmt.Errorf("size mismatch (%d/%d)", j.received, j.size)
		}
		j.size = j.received
		return nil
	}
	if j.received != j.size {
		return fmt.Errorf("size mismatch (%d/%d)", j.received, j.size)
	}
	return nil
}
