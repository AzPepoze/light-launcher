package plugins

import (
	"sync"
	"time"
)

// fakeClock is a manually advanced clock for idle tests.
type fakeClock struct {
	mu     sync.Mutex
	now    time.Time
	timers []*fakeTimer
}

type fakeTimer struct {
	at time.Time
	ch chan time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) After(d time.Duration) <-chan time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	timer := &fakeTimer{at: c.now.Add(d), ch: make(chan time.Time, 1)}
	c.timers = append(c.timers, timer)
	return timer.ch
}

// HasTimer reports whether a timer is waiting for exactly d from now.
func (c *fakeClock) HasTimer(d time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	due := c.now.Add(d)
	for _, timer := range c.timers {
		if timer.at.Equal(due) {
			return true
		}
	}
	return false
}

// Advance moves the clock forward and fires every timer that is due.
func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
	var pending []*fakeTimer
	for _, timer := range c.timers {
		if timer.at.After(c.now) {
			pending = append(pending, timer)
			continue
		}
		select {
		case timer.ch <- timer.at:
		default:
		}
	}
	c.timers = pending
}
