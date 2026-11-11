// Package scheduler runs background jobs on a bounded pool with jitter, timeouts and backoff.
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"light-launcher/core/internal/logger"
)

const (
	defaultWorkers = 2
	firstRunMin    = 30 * time.Second
	firstRunMax    = 120 * time.Second
	maxBackoff     = 6 * time.Hour
)

// ErrSkipped means a job found nothing to do. It keeps its last result and runs again next interval.
var ErrSkipped = errors.New("job skipped")

// Job is one named background task. A zero Jitter adds no delay. A zero Timeout sets no deadline.
type Job struct {
	Name           string
	PluginID       string
	Interval       time.Duration
	Jitter         time.Duration
	Timeout        time.Duration
	ChangesLibrary bool
	Run            func(ctx context.Context) error
}

// Options configures a Scheduler. Optional fields take defaults; State nil keeps runs in memory only.
type Options struct {
	Jobs    []Job
	Workers int
	Clock   Clock
	State   MetaStore
	// Notify receives "library.changed" after a successful job that sets ChangesLibrary.
	Notify func(channel string, data any)
	Rand   *rand.Rand
}

// Status is the read-only view the API returns.
type Status struct {
	Jobs []JobStatus `json:"jobs"`
}

// JobStatus is one job in a Status.
type JobStatus struct {
	Name            string  `json:"name"`
	PluginID        string  `json:"pluginId,omitempty"`
	IntervalSeconds int64   `json:"intervalSeconds"`
	LastRunAt       *string `json:"lastRunAt"`
	LastDurationMs  int64   `json:"lastDurationMs"`
	LastError       string  `json:"lastError,omitempty"`
	NextRunAt       *string `json:"nextRunAt"`
	Running         bool    `json:"running"`
}

type job struct {
	Job
	running      bool
	lastRunAt    time.Time
	lastDuration time.Duration
	lastError    string
	failures     int
	nextRunAt    time.Time
	loggedError  string
}

// Scheduler runs jobs on a fixed worker pool until its context ends.
type Scheduler struct {
	clock  Clock
	state  MetaStore
	notify func(string, any)
	rng    *rand.Rand
	slots  chan struct{}
	free   chan struct{}
	ready  chan struct{}

	mu      sync.Mutex
	jobs    []*job
	byName  map[string]*job
	ctx     context.Context
	started bool
	closed  bool

	persistMu sync.Mutex
	wg        sync.WaitGroup
}

// New prepares the scheduler and loads the last run time of every job.
func New(ctx context.Context, opts Options) (*Scheduler, error) {
	clock := opts.Clock
	if clock == nil {
		clock = systemClock{}
	}
	workers := opts.Workers
	if workers <= 0 {
		workers = defaultWorkers
	}
	rng := opts.Rand
	if rng == nil {
		rng = rand.New(rand.NewSource(clock.Now().UnixNano()))
	}
	lastRuns, err := loadLastRuns(ctx, opts.State)
	if err != nil {
		return nil, err
	}
	s := &Scheduler{
		clock:  clock,
		state:  opts.State,
		notify: opts.Notify,
		rng:    rng,
		slots:  make(chan struct{}, workers),
		free:   make(chan struct{}, 1),
		ready:  make(chan struct{}),
		byName: map[string]*job{},
	}
	for _, spec := range opts.Jobs {
		if spec.Name == "" {
			return nil, errors.New("scheduler: job needs a name")
		}
		if spec.Interval <= 0 {
			return nil, fmt.Errorf("scheduler: job %s needs a positive interval", spec.Name)
		}
		if spec.Run == nil {
			return nil, fmt.Errorf("scheduler: job %s needs a run function", spec.Name)
		}
		if _, exists := s.byName[spec.Name]; exists {
			return nil, fmt.Errorf("scheduler: duplicate job %s", spec.Name)
		}
		item := &job{Job: spec, lastRunAt: lastRuns[spec.Name]}
		s.jobs = append(s.jobs, item)
		s.byName[spec.Name] = item
	}
	sort.Slice(s.jobs, func(i, j int) bool { return s.jobs[i].Name < s.jobs[j].Name })
	return s, nil
}

// Ready is closed once Run has scheduled every job for the first time.
func (s *Scheduler) Ready() <-chan struct{} {
	return s.ready
}

// Run dispatches jobs until ctx ends, then waits for every running job.
func (s *Scheduler) Run(ctx context.Context) {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return
	}
	s.started = true
	s.ctx = ctx
	now := s.clock.Now()
	for _, item := range s.jobs {
		s.scheduleFirstLocked(item, now)
	}
	s.mu.Unlock()
	close(s.ready)

	defer func() {
		s.mu.Lock()
		s.closed = true
		s.mu.Unlock()
		s.wg.Wait()
	}()

	for {
		next, at, ok := s.nextDue()
		if ok && !at.After(s.clock.Now()) {
			if s.startJob(ctx, next) {
				continue
			}
			select {
			case <-ctx.Done():
				return
			case <-s.free:
			}
			continue
		}

		var timer <-chan time.Time
		if ok {
			timer = s.clock.After(at)
		}
		select {
		case <-ctx.Done():
			return
		case <-s.free:
		case <-timer:
		}
	}
}

// Status returns a snapshot of every job. It never blocks on running jobs.
func (s *Scheduler) Status() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	status := Status{Jobs: make([]JobStatus, 0, len(s.jobs))}
	for _, item := range s.jobs {
		entry := JobStatus{
			Name:            item.Name,
			PluginID:        item.PluginID,
			IntervalSeconds: int64(item.Interval / time.Second),
			LastError:       item.lastError,
			Running:         item.running,
		}
		if !item.lastRunAt.IsZero() {
			entry.LastRunAt = timeText(item.lastRunAt)
			entry.LastDurationMs = item.lastDuration.Milliseconds()
		}
		if !item.running && !item.nextRunAt.IsZero() {
			entry.NextRunAt = timeText(item.nextRunAt)
		}
		status.Jobs = append(status.Jobs, entry)
	}
	return status
}

// RunNow starts a job once. A job that is already running, or a full worker pool, is skipped.
func (s *Scheduler) RunNow(name string) (bool, error) {
	s.mu.Lock()
	item, ok := s.byName[name]
	if !ok {
		s.mu.Unlock()
		return false, fmt.Errorf("unknown job %q", name)
	}
	if s.closed || s.ctx == nil || item.running {
		s.mu.Unlock()
		return false, nil
	}
	ctx := s.ctx
	s.mu.Unlock()
	if !s.startJob(ctx, item) {
		return false, nil
	}
	return true, nil
}

func (s *Scheduler) scheduleFirstLocked(item *job, now time.Time) {
	first := now.Add(firstDelay(s.rng))
	if !item.lastRunAt.IsZero() {
		due := item.lastRunAt.Add(item.Interval)
		if due.After(first) {
			first = due
		}
	}
	item.nextRunAt = first
}

func firstDelay(rng *rand.Rand) time.Duration {
	span := int64(firstRunMax - firstRunMin)
	return firstRunMin + time.Duration(rng.Int63n(span+1))
}

func (s *Scheduler) nextDue() (*job, time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var pick *job
	for _, item := range s.jobs {
		if item.running || item.nextRunAt.IsZero() {
			continue
		}
		if pick == nil || item.nextRunAt.Before(pick.nextRunAt) {
			pick = item
		}
	}
	if pick == nil {
		return nil, time.Time{}, false
	}
	return pick, pick.nextRunAt, true
}

func (s *Scheduler) startJob(ctx context.Context, item *job) bool {
	select {
	case s.slots <- struct{}{}:
	default:
		return false
	}
	s.mu.Lock()
	if s.closed || item.running {
		s.mu.Unlock()
		<-s.slots
		return false
	}
	item.running = true
	item.nextRunAt = time.Time{}
	s.wg.Add(1)
	s.mu.Unlock()

	go s.execute(ctx, item)
	return true
}

func (s *Scheduler) execute(parent context.Context, item *job) {
	defer s.wg.Done()
	defer func() {
		<-s.slots
		select {
		case s.free <- struct{}{}:
		default:
		}
	}()

	start := s.clock.Now()
	runCtx, cancel := context.WithCancel(parent)
	timeoutDone, timedOut := s.watchTimeout(runCtx, cancel, start, item.Timeout)
	err := item.Run(runCtx)
	cancel()
	if timeoutDone != nil {
		<-timeoutDone
	}
	if timedOut() && err != nil {
		err = fmt.Errorf("timed out after %s: %w", item.Timeout, err)
	}
	s.record(item, start, s.clock.Now(), err)
}

// watchTimeout cancels runCtx when the injected clock reaches the deadline. The channel closes
// once the watcher exits, and the function reports whether the timeout fired.
func (s *Scheduler) watchTimeout(runCtx context.Context, cancel context.CancelFunc, start time.Time, timeout time.Duration) (<-chan struct{}, func() bool) {
	if timeout <= 0 {
		return nil, func() bool { return false }
	}
	var fired atomic.Bool
	done := make(chan struct{})
	go func() {
		defer close(done)
		select {
		case <-s.clock.After(start.Add(timeout)):
			fired.Store(true)
			cancel()
		case <-runCtx.Done():
		}
	}()
	return done, fired.Load
}

func (s *Scheduler) record(item *job, start, finish time.Time, runErr error) {
	var failed, recovered, notify bool
	s.mu.Lock()
	item.running = false
	if errors.Is(runErr, ErrSkipped) {
		item.nextRunAt = finish.Add(item.Interval + s.jitter(item.Jitter))
		s.mu.Unlock()
		return
	}
	item.lastRunAt = finish
	item.lastDuration = finish.Sub(start)
	item.lastError = ""
	if runErr != nil {
		item.lastError = runErr.Error()
		item.failures++
		item.nextRunAt = finish.Add(backoff(item.Interval, item.failures) + s.jitter(item.Jitter))
		if item.loggedError != item.lastError {
			item.loggedError = item.lastError
			failed = true
		}
	} else {
		item.failures = 0
		item.nextRunAt = finish.Add(item.Interval + s.jitter(item.Jitter))
		recovered = item.loggedError != ""
		item.loggedError = ""
		notify = item.ChangesLibrary && s.notify != nil
	}
	s.mu.Unlock()

	switch {
	case failed:
		logger.Warn("scheduler", "job %s failed: %v", item.Name, runErr)
	case recovered:
		logger.Info("scheduler", "job %s works again", item.Name)
	}
	s.save()
	if notify {
		s.notify("library.changed", map[string]any{})
	}
}

func (s *Scheduler) jitter(max time.Duration) time.Duration {
	if max <= 0 {
		return 0
	}
	return time.Duration(s.rng.Int63n(int64(max)))
}

func (s *Scheduler) save() {
	if s.state == nil {
		return
	}
	s.persistMu.Lock()
	defer s.persistMu.Unlock()
	s.mu.Lock()
	times := s.lastRunsLocked()
	s.mu.Unlock()
	raw, err := encodeState(times)
	if err != nil {
		logger.Warn("scheduler", "could not encode last run times: %v", err)
		return
	}
	if err := s.state.SetMeta(context.Background(), stateKey, raw); err != nil {
		logger.Warn("scheduler", "could not save last run times: %v", err)
	}
}

func (s *Scheduler) lastRunsLocked() map[string]time.Time {
	times := make(map[string]time.Time, len(s.jobs))
	for _, item := range s.jobs {
		if !item.lastRunAt.IsZero() {
			times[item.Name] = item.lastRunAt
		}
	}
	return times
}

func backoff(interval time.Duration, failures int) time.Duration {
	delay := interval
	for i := 1; i < failures && delay < maxBackoff; i++ {
		if delay > maxBackoff/2 {
			return maxBackoff
		}
		delay *= 2
	}
	if delay > maxBackoff {
		return maxBackoff
	}
	return delay
}

func timeText(at time.Time) *string {
	text := at.UTC().Format(time.RFC3339Nano)
	return &text
}
