package scheduler

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var testStart = time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)

type fakeWaiter struct {
	at time.Time
	ch chan time.Time
}

type fakeClock struct {
	mu      sync.Mutex
	now     time.Time
	waiters []fakeWaiter
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: testStart}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) After(until time.Time) <-chan time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	ch := make(chan time.Time, 1)
	if !until.After(c.now) {
		ch <- c.now
		return ch
	}
	c.waiters = append(c.waiters, fakeWaiter{at: until, ch: ch})
	return ch
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
	kept := c.waiters[:0]
	for _, waiter := range c.waiters {
		if waiter.at.After(c.now) {
			kept = append(kept, waiter)
			continue
		}
		waiter.ch <- c.now
	}
	c.waiters = kept
}

type memoryMeta struct {
	mu     sync.Mutex
	values map[string]string
}

func newMemoryMeta() *memoryMeta {
	return &memoryMeta{values: map[string]string{}}
}

func (m *memoryMeta) GetMeta(_ context.Context, key string) (string, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, ok := m.values[key]
	return value, ok, nil
}

func (m *memoryMeta) SetMeta(_ context.Context, key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.values[key] = value
	return nil
}

func startScheduler(t *testing.T, sched *Scheduler) (context.CancelFunc, <-chan struct{}) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		sched.Run(ctx)
	}()
	<-sched.Ready()
	return cancel, done
}

func stopScheduler(t *testing.T, cancel context.CancelFunc, done <-chan struct{}) {
	t.Helper()
	cancel()
	waitChannel(t, done, "scheduler did not stop")
}

func waitChannel(t *testing.T, ch <-chan struct{}, message string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal(message)
	}
}

func waitName(t *testing.T, ch <-chan string, message string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal(message)
	}
}

func waitFor(t *testing.T, check func() bool, message string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !check() {
		if time.Now().After(deadline) {
			t.Fatal(message)
		}
		time.Sleep(time.Millisecond)
	}
}

func jobStatus(t *testing.T, sched *Scheduler, name string) JobStatus {
	t.Helper()
	for _, item := range sched.Status().Jobs {
		if item.Name == name {
			return item
		}
	}
	t.Fatalf("job %s is missing", name)
	return JobStatus{}
}

func parseTime(t *testing.T, text *string) time.Time {
	t.Helper()
	if text == nil {
		t.Fatal("time is nil")
	}
	at, err := time.Parse(time.RFC3339Nano, *text)
	if err != nil {
		t.Fatal(err)
	}
	return at
}

func waitNext(t *testing.T, sched *Scheduler, name string, want time.Time) {
	t.Helper()
	waitFor(t, func() bool {
		status := jobStatus(t, sched, name)
		if status.NextRunAt == nil {
			return false
		}
		at, err := time.Parse(time.RFC3339Nano, *status.NextRunAt)
		return err == nil && at.Equal(want)
	}, "next run was not "+want.Format(time.RFC3339Nano))
}

func TestFirstRunUsesRandomStartDelay(t *testing.T) {
	clock := newFakeClock()
	sched, err := New(context.Background(), Options{
		Clock: clock,
		Rand:  rand.New(rand.NewSource(1)),
		Jobs: []Job{{
			Name:     "one",
			Interval: time.Hour,
			Run:      func(context.Context) error { return nil },
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	cancel, done := startScheduler(t, sched)
	defer stopScheduler(t, cancel, done)

	next := parseTime(t, jobStatus(t, sched, "one").NextRunAt)
	delay := next.Sub(clock.Now())
	if delay < firstRunMin || delay > firstRunMax {
		t.Fatalf("first delay = %v, want %v to %v", delay, firstRunMin, firstRunMax)
	}
}

func TestJobRunsOnItsInterval(t *testing.T) {
	clock := newFakeClock()
	ran := make(chan struct{}, 2)
	sched, err := New(context.Background(), Options{
		Clock: clock,
		Jobs: []Job{{
			Name:     "one",
			Interval: time.Hour,
			Run: func(context.Context) error {
				ran <- struct{}{}
				return nil
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	cancel, done := startScheduler(t, sched)
	defer stopScheduler(t, cancel, done)

	clock.Advance(firstRunMax)
	waitChannel(t, ran, "first run did not happen")
	waitNext(t, sched, "one", clock.Now().Add(time.Hour))

	clock.Advance(time.Hour - time.Second)
	select {
	case <-ran:
		t.Fatal("job ran before its interval ended")
	case <-time.After(20 * time.Millisecond):
	}

	clock.Advance(time.Second)
	waitChannel(t, ran, "second run did not happen")
}

func TestJitterStaysWithinBounds(t *testing.T) {
	clock := newFakeClock()
	ran := make(chan struct{}, 1)
	sched, err := New(context.Background(), Options{
		Clock: clock,
		Rand:  rand.New(rand.NewSource(2)),
		Jobs: []Job{{
			Name:     "one",
			Interval: time.Hour,
			Jitter:   10 * time.Minute,
			Run: func(context.Context) error {
				ran <- struct{}{}
				return nil
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	cancel, done := startScheduler(t, sched)
	defer stopScheduler(t, cancel, done)

	clock.Advance(firstRunMax)
	waitChannel(t, ran, "first run did not happen")
	waitFor(t, func() bool { return jobStatus(t, sched, "one").NextRunAt != nil }, "next run is missing")
	next := parseTime(t, jobStatus(t, sched, "one").NextRunAt)
	delay := next.Sub(clock.Now())
	if delay < time.Hour || delay >= time.Hour+10*time.Minute {
		t.Fatalf("jittered delay = %v, want %v to %v", delay, time.Hour, time.Hour+10*time.Minute)
	}
}

func TestRunningJobIsNeverStartedTwice(t *testing.T) {
	clock := newFakeClock()
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	sched, err := New(context.Background(), Options{
		Clock:   clock,
		Workers: 2,
		Jobs: []Job{{
			Name:     "one",
			Interval: time.Hour,
			Run: func(context.Context) error {
				started <- struct{}{}
				<-release
				return nil
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	cancel, done := startScheduler(t, sched)
	defer stopScheduler(t, cancel, done)

	clock.Advance(firstRunMax)
	waitChannel(t, started, "first run did not start")

	clock.Advance(48 * time.Hour)
	select {
	case <-started:
		t.Fatal("a running job started again")
	case <-time.After(20 * time.Millisecond):
	}
	if ran, err := sched.RunNow("one"); err != nil || ran {
		t.Fatalf("RunNow while running = %v, %v; want false, nil", ran, err)
	}

	close(release)
	waitNext(t, sched, "one", clock.Now().Add(time.Hour))
}

func TestPoolRunsAtMostWorkers(t *testing.T) {
	clock := newFakeClock()
	started := make(chan string, 2)
	release := make(chan struct{})
	run := func(name string) func(context.Context) error {
		return func(context.Context) error {
			started <- name
			<-release
			return nil
		}
	}
	sched, err := New(context.Background(), Options{
		Clock:   clock,
		Workers: 1,
		Jobs: []Job{
			{Name: "a", Interval: time.Hour, Run: run("a")},
			{Name: "b", Interval: time.Hour, Run: run("b")},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	cancel, done := startScheduler(t, sched)
	defer stopScheduler(t, cancel, done)

	clock.Advance(firstRunMax)
	waitName(t, started, "first job did not start")
	select {
	case name := <-started:
		t.Fatalf("job %s started on a full pool", name)
	case <-time.After(20 * time.Millisecond):
	}

	close(release)
	waitName(t, started, "second job did not start after a slot opened")
}

func TestJobTimeoutCancelsRun(t *testing.T) {
	clock := newFakeClock()
	started := make(chan struct{})
	result := make(chan error, 1)
	sched, err := New(context.Background(), Options{
		Clock: clock,
		Jobs: []Job{{
			Name:     "slow",
			Interval: time.Hour,
			Timeout:  20 * time.Millisecond,
			Run: func(ctx context.Context) error {
				close(started)
				<-ctx.Done()
				result <- ctx.Err()
				return ctx.Err()
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	cancel, done := startScheduler(t, sched)
	defer stopScheduler(t, cancel, done)

	clock.Advance(firstRunMax)
	waitChannel(t, started, "job did not start")
	clock.Advance(20 * time.Millisecond)
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("job context error = %v, want canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("job was not stopped by its timeout")
	}
	waitFor(t, func() bool {
		return strings.Contains(jobStatus(t, sched, "slow").LastError, "timed out")
	}, "timeout was not recorded")
}

func TestShutdownWaitsForRunningJob(t *testing.T) {
	clock := newFakeClock()
	running := make(chan struct{})
	release := make(chan struct{})
	sched, err := New(context.Background(), Options{
		Clock: clock,
		Jobs: []Job{{
			Name:     "one",
			Interval: time.Hour,
			Run: func(context.Context) error {
				close(running)
				<-release
				return nil
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	cancel, done := startScheduler(t, sched)

	clock.Advance(firstRunMax)
	waitChannel(t, running, "job did not start")
	cancel()
	select {
	case <-done:
		t.Fatal("scheduler stopped before its job finished")
	case <-time.After(20 * time.Millisecond):
	}

	close(release)
	waitChannel(t, done, "scheduler did not stop after the job finished")
}

func TestFailureBacksOffAndCapsAtSixHours(t *testing.T) {
	clock := newFakeClock()
	var runs atomic.Int32
	sched, err := New(context.Background(), Options{
		Clock: clock,
		Jobs: []Job{{
			Name:     "one",
			Interval: time.Hour,
			Run: func(context.Context) error {
				runs.Add(1)
				return errors.New("boom")
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	cancel, done := startScheduler(t, sched)
	defer stopScheduler(t, cancel, done)

	wantDelays := []time.Duration{time.Hour, 2 * time.Hour, 4 * time.Hour, 6 * time.Hour}
	for index, delay := range wantDelays {
		if index == 0 {
			clock.Advance(firstRunMax)
		} else {
			clock.Advance(wantDelays[index-1])
		}
		wantRuns := int32(index + 1)
		waitFor(t, func() bool { return runs.Load() == wantRuns }, "run count did not reach the expected value")
		waitNext(t, sched, "one", clock.Now().Add(delay))
	}
}

func TestPersistenceRestoresLastRun(t *testing.T) {
	store := newMemoryMeta()
	clock := newFakeClock()
	ran := make(chan struct{}, 1)
	job := Job{
		Name:     "one",
		Interval: 2 * time.Hour,
		Run: func(context.Context) error {
			ran <- struct{}{}
			return nil
		},
	}
	first, err := New(context.Background(), Options{Clock: clock, State: store, Jobs: []Job{job}})
	if err != nil {
		t.Fatal(err)
	}
	cancel, done := startScheduler(t, first)
	clock.Advance(firstRunMax)
	waitChannel(t, ran, "first run did not happen")
	waitFor(t, func() bool { return jobStatus(t, first, "one").LastRunAt != nil }, "last run was not recorded")
	lastRun := parseTime(t, jobStatus(t, first, "one").LastRunAt)
	stopScheduler(t, cancel, done)

	waitFor(t, func() bool {
		_, found, _ := store.GetMeta(context.Background(), stateKey)
		return found
	}, "last run was not persisted")

	secondClock := newFakeClockAt(lastRun.Add(30 * time.Minute))
	second, err := New(context.Background(), Options{Clock: secondClock, State: store, Jobs: []Job{job}})
	if err != nil {
		t.Fatal(err)
	}
	if got := parseTime(t, jobStatus(t, second, "one").LastRunAt); !got.Equal(lastRun) {
		t.Fatalf("restored last run = %v, want %v", got, lastRun)
	}
	cancel, done = startScheduler(t, second)
	next := parseTime(t, jobStatus(t, second, "one").NextRunAt)
	if want := lastRun.Add(2 * time.Hour); !next.Equal(want) {
		t.Fatalf("next run after restart = %v, want %v", next, want)
	}
	stopScheduler(t, cancel, done)

	thirdClock := newFakeClockAt(lastRun.Add(3 * time.Hour))
	third, err := New(context.Background(), Options{Clock: thirdClock, State: store, Jobs: []Job{job}})
	if err != nil {
		t.Fatal(err)
	}
	cancel, done = startScheduler(t, third)
	defer stopScheduler(t, cancel, done)
	delay := parseTime(t, jobStatus(t, third, "one").NextRunAt).Sub(thirdClock.Now())
	if delay < firstRunMin || delay > firstRunMax {
		t.Fatalf("overdue restart delay = %v, want %v to %v", delay, firstRunMin, firstRunMax)
	}
}

func newFakeClockAt(at time.Time) *fakeClock {
	return &fakeClock{now: at}
}

func TestRunNowRunsOneJob(t *testing.T) {
	clock := newFakeClock()
	ran := make(chan struct{}, 2)
	sched, err := New(context.Background(), Options{
		Clock: clock,
		Jobs: []Job{{
			Name:     "one",
			Interval: time.Hour,
			Run: func(context.Context) error {
				ran <- struct{}{}
				return nil
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if started, err := sched.RunNow("one"); err != nil || started {
		t.Fatalf("RunNow before Run = %v, %v; want false, nil", started, err)
	}
	cancel, done := startScheduler(t, sched)
	defer stopScheduler(t, cancel, done)

	if started, err := sched.RunNow("one"); err != nil || !started {
		t.Fatalf("RunNow = %v, %v; want true, nil", started, err)
	}
	waitChannel(t, ran, "manual run did not happen")
	if _, err := sched.RunNow("missing"); err == nil {
		t.Fatal("an unknown job must fail")
	}
}

func TestSkippedJobKeepsLastResult(t *testing.T) {
	clock := newFakeClock()
	notified := make(chan string, 1)
	sched, err := New(context.Background(), Options{
		Clock: clock,
		Notify: func(channel string, _ any) {
			notified <- channel
		},
		Jobs: []Job{{
			Name:           "one",
			Interval:       time.Hour,
			ChangesLibrary: true,
			Run:            func(context.Context) error { return ErrSkipped },
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	cancel, done := startScheduler(t, sched)
	defer stopScheduler(t, cancel, done)
	clock.Advance(firstRunMax)
	waitNext(t, sched, "one", clock.Now().Add(time.Hour))
	status := jobStatus(t, sched, "one")
	if status.LastRunAt != nil || status.LastError != "" {
		t.Fatalf("skipped job recorded a result: %+v", status)
	}
	select {
	case channel := <-notified:
		t.Fatalf("skipped job notified %s", channel)
	case <-time.After(20 * time.Millisecond):
	}
}

func TestSuccessfulLibraryJobNotifies(t *testing.T) {
	clock := newFakeClock()
	notified := make(chan string, 1)
	sched, err := New(context.Background(), Options{
		Clock: clock,
		Notify: func(channel string, _ any) {
			notified <- channel
		},
		Jobs: []Job{{
			Name:           "one",
			Interval:       time.Hour,
			ChangesLibrary: true,
			Run:            func(context.Context) error { return nil },
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	cancel, done := startScheduler(t, sched)
	defer stopScheduler(t, cancel, done)
	clock.Advance(firstRunMax)
	select {
	case channel := <-notified:
		if channel != "library.changed" {
			t.Fatalf("channel = %q, want library.changed", channel)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("library change was not notified")
	}
}

func TestEncodeStateVersion(t *testing.T) {
	at := testStart
	raw, err := encodeState(map[string]time.Time{"one": at})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(raw, `"version":1`) {
		t.Fatalf("state = %s, want version 1", raw)
	}
	runs, err := decodeState(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !runs["one"].Equal(at) {
		t.Fatalf("round trip = %v, want %v", runs["one"], at)
	}
}

func TestDecodeStateReadsVersionZero(t *testing.T) {
	at := testStart
	raw := `{"lastRuns":{"one":"` + at.Format(time.RFC3339Nano) + `"}}`
	runs, err := decodeState(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !runs["one"].Equal(at) {
		t.Fatalf("version 0 last run = %v, want %v", runs["one"], at)
	}
}

func TestDecodeStateRejectsNewerVersion(t *testing.T) {
	if _, err := decodeState(`{"version":2,"lastRuns":{}}`); err == nil {
		t.Fatal("a future state version must fail")
	}
}

func TestSetJobsAddsAndRemovesJobs(t *testing.T) {
	clock := newFakeClock()
	run := func(context.Context) error { return nil }
	sched, err := New(context.Background(), Options{
		Clock: clock,
		Rand:  rand.New(rand.NewSource(2)),
		Jobs:  []Job{{Name: "one", Interval: time.Hour, Run: run}},
	})
	if err != nil {
		t.Fatal(err)
	}
	cancel, done := startScheduler(t, sched)
	defer stopScheduler(t, cancel, done)

	if err := sched.SetJobs([]Job{
		{Name: "one", Interval: 2 * time.Hour, Run: run},
		{Name: "two", Interval: time.Minute, Run: run},
	}); err != nil {
		t.Fatal(err)
	}
	jobStatus(t, sched, "one")
	jobStatus(t, sched, "two")
	if got := jobStatus(t, sched, "one").IntervalSeconds; got != 7200 {
		t.Errorf("one interval = %d, want 7200", got)
	}

	if err := sched.SetJobs([]Job{{Name: "two", Interval: time.Minute, Run: run}}); err != nil {
		t.Fatal(err)
	}
	for _, item := range sched.Status().Jobs {
		if item.Name == "one" {
			t.Fatal("removed job one is still listed")
		}
	}
	jobStatus(t, sched, "two")
}

func TestSetJobsRejectsDuplicates(t *testing.T) {
	sched, err := New(context.Background(), Options{})
	if err != nil {
		t.Fatal(err)
	}
	run := func(context.Context) error { return nil }
	if err := sched.SetJobs([]Job{{Name: "same", Interval: time.Hour, Run: run}, {Name: "same", Interval: time.Hour, Run: run}}); err == nil {
		t.Fatal("duplicate job names must be rejected")
	}
}
