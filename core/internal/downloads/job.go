package downloads

import (
	"context"
	"sync"
	"time"
)

// stopKind says why a running download was asked to stop. The first reason wins.
type stopKind int

const (
	stopNone stopKind = iota
	stopPause
	stopCancel
	stopRemove
	stopShutdown
)

// job is one download while it runs. Its progress fields are guarded by mu.
type job struct {
	row    Download     // the stored row when this run started
	probe  *ProbeResult // probe from enqueue; the first run uses it instead of probing again
	resume []Segment    // segments left by the previous run
	cancel context.CancelFunc
	done   chan struct{} // closed when the run returns

	mu       sync.Mutex
	stop     stopKind
	gate     *Gate
	size     int64
	etag     string
	segments []Segment
	received int64
	meter    speedMeter
	lastByte time.Time
	strikes  int
}

func newJob(row Download, probe *ProbeResult, resume []Segment) *job {
	return &job{
		row:      row,
		probe:    probe,
		resume:   resume,
		done:     make(chan struct{}),
		lastByte: time.Now(),
	}
}

// requestStop records why the run is ending and cancels it.
func (j *job) requestStop(kind stopKind) {
	j.mu.Lock()
	if j.stop == stopNone {
		j.stop = kind
	}
	j.mu.Unlock()
	j.cancel()
}

func (j *job) stopReason() stopKind {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.stop
}

// setLayout records the size, ETag and segments once the file is open.
func (j *job) setLayout(size int64, etag string, segments []Segment) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.size = size
	j.etag = etag
	j.segments = segments
	j.gate = NewGate(len(segments))
	j.received = 0
	for _, seg := range segments {
		j.received += seg.Done
	}
	j.lastByte = time.Now()
}

// segment returns a copy of one segment.
func (j *job) segment(idx int) Segment {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.segments[idx]
}

// pendingSegments lists the indexes of segments that still need bytes.
func (j *job) pendingSegments() []int {
	j.mu.Lock()
	defer j.mu.Unlock()
	var out []int
	for _, seg := range j.segments {
		if remaining(seg) > 0 {
			out = append(out, seg.Index)
		}
	}
	return out
}

// snapshotSegments returns a copy of every segment, for persisting.
func (j *job) snapshotSegments() []Segment {
	j.mu.Lock()
	defer j.mu.Unlock()
	return append([]Segment(nil), j.segments...)
}

func (j *job) gateRef() *Gate {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.gate
}

// addBytes records n bytes written for a segment and feeds the speed meter.
func (j *job) addBytes(idx int, n int64) {
	now := time.Now()
	j.mu.Lock()
	j.segments[idx].Done += n
	j.received += n
	j.meter.add(now, n)
	j.lastByte = now
	j.strikes = 0
	gate := j.gate
	j.mu.Unlock()
	gate.OnProgress()
}

// resetStream starts a whole-file stream again after a failed attempt.
func (j *job) resetStream() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.segments[0].Done = 0
	j.received = 0
}

// noteRateLimit counts a rate limit that halved the limit and returns the strike count.
func (j *job) noteRateLimit(halved bool) int {
	j.mu.Lock()
	defer j.mu.Unlock()
	if halved {
		j.strikes++
	}
	return j.strikes
}

func (j *job) lastByteAt() time.Time {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.lastByte
}

// progress is a point-in-time view of a running download.
type progress struct {
	received    int64
	total       int64
	speed       float64
	connections int
}

func (j *job) progress(now time.Time) progress {
	j.mu.Lock()
	defer j.mu.Unlock()
	p := progress{received: j.received, total: j.size, speed: j.meter.rate(now)}
	if j.gate != nil {
		p.connections = j.gate.Limit()
	}
	return p
}
