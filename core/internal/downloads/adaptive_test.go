package downloads

import (
	"context"
	"net/http"
	"testing"
	"time"
)

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time          { return c.now }
func (c *fakeClock) Advance(d time.Duration) { c.now = c.now.Add(d) }

func newTestGate(ceiling int) (*Gate, *fakeClock) {
	clock := &fakeClock{now: time.Unix(1_700_000_000, 0)}
	return newGate(ceiling, clock.Now), clock
}

func TestGateHalvesOnRateLimitNeverBelowOne(t *testing.T) {
	g, clock := newTestGate(8)
	g.OnRateLimit()
	if g.Limit() != 4 {
		t.Fatalf("limit = %d, want 4", g.Limit())
	}
	for _, want := range []int{2, 1, 1} {
		clock.Advance(halveCooldown)
		g.OnRateLimit()
		if g.Limit() != want {
			t.Fatalf("limit = %d, want %d", g.Limit(), want)
		}
	}
}

func TestGateBurstOfRateLimitsHalvesOnce(t *testing.T) {
	g, _ := newTestGate(8)
	halved := 0
	for i := 0; i < 8; i++ {
		if g.OnRateLimit() {
			halved++
		}
	}
	if halved != 1 || g.Limit() != 4 {
		t.Fatalf("halved %d times, limit %d; want once and 4", halved, g.Limit())
	}
}

func TestGateRampsUpOnlyAfterQuietPeriod(t *testing.T) {
	g, clock := newTestGate(4)
	g.OnRateLimit()
	clock.Advance(rampInterval - time.Second)
	if g.OnProgress() {
		t.Fatal("ramped before the quiet period ended")
	}
	clock.Advance(time.Second)
	if !g.OnProgress() || g.Limit() != 3 {
		t.Fatalf("expected a ramp to 3, limit = %d", g.Limit())
	}
}

func TestGateNeverRampsAboveCeiling(t *testing.T) {
	g, clock := newTestGate(2)
	clock.Advance(rampInterval * 5)
	g.OnProgress()
	g.OnProgress()
	if g.Limit() != 2 {
		t.Fatalf("limit = %d, want the ceiling 2", g.Limit())
	}
}

func TestGateAcquireBlocksAtLimitAndWakesOnRelease(t *testing.T) {
	g := NewGate(1)
	ctx := context.Background()
	if err := g.Acquire(ctx); err != nil {
		t.Fatal(err)
	}
	acquired := make(chan struct{})
	go func() {
		g.Acquire(ctx)
		close(acquired)
	}()
	select {
	case <-acquired:
		t.Fatal("second acquire passed the limit")
	case <-time.After(20 * time.Millisecond):
	}
	g.Release()
	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("waiter was not woken by release")
	}
}

func TestGateAcquireHonoursContext(t *testing.T) {
	g := NewGate(1)
	g.Acquire(context.Background())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := g.Acquire(ctx); err != context.Canceled {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestBackoffDelayHonoursHintWithoutJitter(t *testing.T) {
	if got := backoffDelay(10*time.Second, 0, 0.5); got != 10*time.Second {
		t.Fatalf("got %v, want 10s", got)
	}
}

func TestBackoffDelayDoublesPerStrike(t *testing.T) {
	if got := backoffDelay(2*time.Second, 1, 0.5); got != 4*time.Second {
		t.Fatalf("strike 1 = %v, want 4s", got)
	}
	if got := backoffDelay(2*time.Second, 2, 0.5); got != 8*time.Second {
		t.Fatalf("strike 2 = %v, want 8s", got)
	}
}

func TestBackoffDelayCapsAtTwoMinutes(t *testing.T) {
	if got := backoffDelay(10*time.Second, 10, 0.5); got != 2*time.Minute {
		t.Fatalf("capped = %v, want 2m", got)
	}
	if got := backoffDelay(0, 30, 1); got != 2*time.Minute {
		t.Fatalf("no hint capped = %v, want 2m", got)
	}
}

func TestBackoffJitterStaysWithinTwentyPercent(t *testing.T) {
	if got := backoffDelay(10*time.Second, 0, 0); got != 8*time.Second {
		t.Fatalf("low jitter = %v, want 8s", got)
	}
	if got := backoffDelay(10*time.Second, 0, 1); got != 12*time.Second {
		t.Fatalf("high jitter = %v, want 12s", got)
	}
}

func TestBackoffNoHintUsesOneSecondMinimum(t *testing.T) {
	if got := backoffDelay(0, 0, 0.5); got != time.Second {
		t.Fatalf("no hint = %v, want 1s", got)
	}
}

func TestParseRetryAfter(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	if got := parseRetryAfter("", now); got != 0 {
		t.Fatalf("empty = %v, want 0", got)
	}
	if got := parseRetryAfter("abc", now); got != 0 {
		t.Fatalf("garbage = %v, want 0", got)
	}
	if got := parseRetryAfter("5", now); got != 5*time.Second {
		t.Fatalf("seconds = %v, want 5s", got)
	}
	if got := parseRetryAfter("999", now); got != 2*time.Minute {
		t.Fatalf("clamped = %v, want 2m", got)
	}
	date := now.Add(30 * time.Second).UTC().Format(http.TimeFormat)
	if got := parseRetryAfter(date, now); got != 30*time.Second {
		t.Fatalf("http date = %v, want 30s", got)
	}
}
