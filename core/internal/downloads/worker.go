package downloads

import (
	"context"
	"errors"
	"math/rand/v2"
	"os"
	"time"
)

// stallGiveUp ends a download that has been rate limited this long without receiving a byte.
const stallGiveUp = 10 * time.Minute

// rateLimitError means the host answered 429 or 503.
type rateLimitError struct {
	retryAfter time.Duration
}

func (e *rateLimitError) Error() string {
	return "rate limited by host"
}

// fetchSegment downloads one segment, retrying failed requests with backoff. Each retry resumes at the bytes already written.
func (t *transferer) fetchSegment(ctx context.Context, j *job, file *os.File, idx int) error {
	attempt := 0
	for {
		err := t.fetchRange(ctx, j, file, idx)
		if err == nil {
			return nil
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		var limited *rateLimitError
		if errors.As(err, &limited) {
			if err := waitRateLimited(ctx, j, limited); err != nil {
				return err
			}
			continue
		}
		if attempt >= t.retries {
			return err
		}
		if err := sleepContext(ctx, retryDelay(attempt)); err != nil {
			return err
		}
		attempt++
	}
}

// fetchRange makes one ranged request for the bytes a segment still needs.
func (t *transferer) fetchRange(ctx context.Context, j *job, file *os.File, idx int) error {
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

	seg := j.segment(idx)
	from := seg.Start + seg.Done
	if from >= seg.End {
		return nil
	}
	rq := request{url: j.row.URL, headers: j.row.Headers, ranged: true, from: from, end: seg.End}
	return t.fetch(ctx, rq, file, func(n int64) {
		j.addBytes(idx, n)
	})
}

// waitRateLimited halves the connection limit on a new congestion event and sleeps before the next try.
func waitRateLimited(ctx context.Context, j *job, failure *rateLimitError) error {
	if time.Since(j.lastByteAt()) > stallGiveUp {
		return failure
	}
	halved := j.gateRef().OnRateLimit()
	strikes := j.noteRateLimit(halved)
	return sleepContext(ctx, backoffDelay(failure.retryAfter, strikes, rand.Float64()))
}

// sleepContext waits for d, or returns early with the context error.
func sleepContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
