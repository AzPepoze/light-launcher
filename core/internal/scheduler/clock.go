package scheduler

import "time"

// Clock is the time source of the scheduler. After returns once until is reached.
type Clock interface {
	Now() time.Time
	After(until time.Time) <-chan time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

func (systemClock) After(until time.Time) <-chan time.Time {
	wait := time.Until(until)
	if wait <= 0 {
		ready := make(chan time.Time, 1)
		ready <- time.Now()
		return ready
	}
	return time.After(wait)
}
