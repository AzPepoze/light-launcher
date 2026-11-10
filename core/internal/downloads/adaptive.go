package downloads

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	rampInterval        = 20 * time.Second // raise the limit by one after this long without a rate limit
	halveCooldown       = 2 * time.Second  // a burst of rate limits halves the limit once
	maxBackoffDoublings = 10
	backoffJitter       = 0.2
	defaultRetryWait    = time.Second
	minRetryWait        = time.Second
	maxRetryWait        = 2 * time.Minute
	maxRetryDelay       = 8 * time.Second
)

// Gate caps concurrent requests. A host gate has a fixed limit; a download gate halves on rate limits.
type Gate struct {
	mu        sync.Mutex
	ceiling   int
	limit     int
	active    int
	changed   chan struct{}
	lastRamp  time.Time
	lastHalve time.Time
	now       func() time.Time
}

// NewGate returns a gate that starts fully open at `ceiling` connections.
func NewGate(ceiling int) *Gate {
	return newGate(ceiling, time.Now)
}

func newGate(ceiling int, now func() time.Time) *Gate {
	ceiling = max(ceiling, 1)
	return &Gate{ceiling: ceiling, limit: ceiling, changed: make(chan struct{}), lastRamp: now(), now: now}
}

// Acquire waits for a free slot or until ctx ends.
func (g *Gate) Acquire(ctx context.Context) error {
	for {
		g.mu.Lock()
		if g.active < g.limit {
			g.active++
			g.mu.Unlock()
			return nil
		}
		changed := g.changed
		g.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-changed:
		}
	}
}

// Release frees the slot taken by Acquire.
func (g *Gate) Release() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.active = max(g.active-1, 0)
	g.broadcastLocked()
}

// SetLimit changes the ceiling and the current limit, used by host gates.
func (g *Gate) SetLimit(n int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.ceiling = max(n, 1)
	g.limit = g.ceiling
	g.broadcastLocked()
}

// Limit is the number of connections allowed right now.
func (g *Gate) Limit() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.limit
}

// OnRateLimit halves the limit on a new congestion event. It reports whether it halved.
func (g *Gate) OnRateLimit() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := g.now()
	g.lastRamp = now
	if now.Sub(g.lastHalve) < halveCooldown {
		return false
	}
	g.lastHalve = now
	g.limit = max(g.limit/2, 1)
	return true
}

// OnProgress raises the limit by one after a quiet period. It reports whether it raised.
func (g *Gate) OnProgress() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := g.now()
	if g.limit >= g.ceiling || now.Sub(g.lastRamp) < rampInterval {
		return false
	}
	g.limit++
	g.lastRamp = now
	g.broadcastLocked()
	return true
}

// broadcastLocked wakes every waiter; they re-check the limit themselves.
func (g *Gate) broadcastLocked() {
	close(g.changed)
	g.changed = make(chan struct{})
}

// parseRetryAfter reads a Retry-After header given in seconds or as an HTTP date.
// It returns 0 when the header is missing or unusable.
func parseRetryAfter(value string, now time.Time) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if secs, err := strconv.Atoi(value); err == nil {
		return clampRetryWait(time.Duration(secs) * time.Second)
	}
	if at, err := http.ParseTime(value); err == nil {
		return clampRetryWait(at.Sub(now))
	}
	return 0
}

func clampRetryWait(d time.Duration) time.Duration {
	return max(minRetryWait, min(maxRetryWait, d))
}

// backoffDelay waits longer after each strike and adds jitter. `random` is in [0, 1).
func backoffDelay(hint time.Duration, strikes int, random float64) time.Duration {
	base := hint
	if base <= 0 {
		base = defaultRetryWait
	}
	doublings := min(strikes, maxBackoffDoublings)
	grown := min(maxRetryWait, base*time.Duration(1<<doublings))
	factor := 1 + (random*2*backoffJitter - backoffJitter)
	return clampRetryWait(time.Duration(float64(grown) * factor))
}

// retryDelay is the wait before a plain (non rate-limit) retry: 250ms doubling up to 8s.
func retryDelay(attempt int) time.Duration {
	return min(maxRetryDelay, 250*time.Millisecond<<min(attempt, 10))
}
