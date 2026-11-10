package downloads

import (
	"context"
	"sync"
	"time"
)

// RateLimiter is a token bucket shared by every download. A rate of zero means unlimited.
type RateLimiter struct {
	mu     sync.Mutex
	rate   float64 // bytes per second
	tokens float64
	last   time.Time
}

// SetRate changes the limit in bytes per second; zero turns it off.
func (l *RateLimiter) SetRate(bytesPerSecond float64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rate = bytesPerSecond
	l.tokens = 0
	l.last = time.Now()
}

// Wait blocks until n bytes may pass, or until ctx ends.
func (l *RateLimiter) Wait(ctx context.Context, n int) error {
	delay := l.reserve(n, time.Now())
	if delay <= 0 {
		return nil
	}
	return sleepContext(ctx, delay)
}

// reserve takes n tokens now and returns how long the caller must wait to stay within the rate.
func (l *RateLimiter) reserve(n int, now time.Time) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.rate <= 0 {
		return 0
	}
	// the bucket holds at most one second of traffic
	l.tokens = min(l.tokens+now.Sub(l.last).Seconds()*l.rate, l.rate)
	l.last = now
	l.tokens -= float64(n)
	if l.tokens >= 0 {
		return 0
	}
	return time.Duration(-l.tokens / l.rate * float64(time.Second))
}
