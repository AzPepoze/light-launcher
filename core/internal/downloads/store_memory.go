package downloads

import (
	"context"
	"sort"
	"sync"
	"time"
)

// MemoryStore keeps downloads in RAM. It is meant for tests and for running without a database.
type MemoryStore struct {
	mu       sync.Mutex
	nextID   int64
	rows     map[int64]Download
	segments map[int64][]Segment
}

// NewMemoryStore returns an empty in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{rows: map[int64]Download{}, segments: map[int64][]Segment{}}
}

func (s *MemoryStore) Insert(_ context.Context, d Download) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	d.ID = s.nextID
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now()
	}
	s.rows[d.ID] = d
	return d.ID, nil
}

func (s *MemoryStore) Get(_ context.Context, id int64) (Download, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.rows[id]
	if !ok {
		return Download{}, ErrNotFound
	}
	return d, nil
}

func (s *MemoryStore) List(_ context.Context) ([]Download, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Download, 0, len(s.rows))
	for _, d := range s.rows {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out, nil
}

func (s *MemoryStore) Update(_ context.Context, id int64, p Patch) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.rows[id]
	if !ok {
		return ErrNotFound
	}
	p.apply(&d)
	s.rows[id] = d
	return nil
}

func (s *MemoryStore) Delete(_ context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.rows[id]; !ok {
		return ErrNotFound
	}
	delete(s.rows, id)
	delete(s.segments, id)
	return nil
}

func (s *MemoryStore) SaveSegments(_ context.Context, id int64, segments []Segment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.rows[id]; !ok {
		return ErrNotFound
	}
	s.segments[id] = append([]Segment(nil), segments...)
	return nil
}

func (s *MemoryStore) Segments(_ context.Context, id int64) ([]Segment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.rows[id]; !ok {
		return nil, ErrNotFound
	}
	return append([]Segment(nil), s.segments[id]...), nil
}

// apply copies every set field of the patch onto d.
func (p Patch) apply(d *Download) {
	if p.State != nil {
		d.State = *p.State
	}
	if p.Error != nil {
		d.Error = *p.Error
	}
	if p.Size != nil {
		d.Size = *p.Size
	}
	if p.BytesDone != nil {
		d.BytesDone = *p.BytesDone
	}
	if p.Connections != nil {
		d.Connections = *p.Connections
	}
	if p.ETag != nil {
		d.ETag = *p.ETag
	}
	if p.StartedAt != nil {
		d.StartedAt = *p.StartedAt
	}
	if p.FinishedAt != nil {
		d.FinishedAt = *p.FinishedAt
	}
}
