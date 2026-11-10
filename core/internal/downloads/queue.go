package downloads

import (
	"context"
	"fmt"
	"slices"
	"time"

	"light-launcher/core/internal/logger"
)

// Pause stops a queued or running download and keeps its progress.
func (m *Manager) Pause(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, err := m.store.Get(ctx, id)
	if err != nil {
		return err
	}
	if d.State != StateQueued && d.State != StateDownloading {
		return stateError(id, d.State)
	}
	m.stopLocked(id, stopPause)
	if err := m.updateLocked(ctx, id, statePatch(StatePaused, "")); err != nil {
		return err
	}
	m.pumpLocked()
	return nil
}

// Resume starts a paused download again from the bytes it already has.
func (m *Manager) Resume(ctx context.Context, id int64) error {
	return m.requeue(ctx, id, StatePaused)
}

// Retry queues a failed or canceled download again.
func (m *Manager) Retry(ctx context.Context, id int64) error {
	return m.requeue(ctx, id, StateError, StateCanceled)
}

// Cancel stops a download. Its part file is kept, so Retry can continue it.
func (m *Manager) Cancel(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, err := m.store.Get(ctx, id)
	if err != nil {
		return err
	}
	if d.State == StateDone {
		return stateError(id, d.State)
	}
	m.stopLocked(id, stopCancel)
	if err := m.updateLocked(ctx, id, statePatch(StateCanceled, "")); err != nil {
		return err
	}
	m.pumpLocked()
	return nil
}

// Remove deletes a download. A running transfer is stopped first. Files are deleted only when deleteFile is set.
func (m *Manager) Remove(ctx context.Context, id int64, deleteFile bool) error {
	j, d, err := m.detach(ctx, id)
	if err != nil {
		return err
	}
	if j != nil {
		<-j.done
	}
	if deleteFile {
		removeFiles(d.Dest)
	}
	return nil
}

// detach stops a running transfer and deletes its row. It returns the job to wait for.
func (m *Manager) detach(ctx context.Context, id int64) (*job, Download, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, err := m.store.Get(ctx, id)
	if err != nil {
		return nil, Download{}, err
	}
	j := m.active[id]
	m.stopLocked(id, stopRemove)
	if err := m.store.Delete(ctx, id); err != nil {
		return nil, Download{}, err
	}
	return j, d, nil
}

func (m *Manager) requeue(ctx context.Context, id int64, allowed ...State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, err := m.store.Get(ctx, id)
	if err != nil {
		return err
	}
	if !slices.Contains(allowed, d.State) {
		return stateError(id, d.State)
	}
	if err := m.updateLocked(ctx, id, statePatch(StateQueued, "")); err != nil {
		return err
	}
	m.pumpLocked()
	return nil
}

func stateError(id int64, state State) error {
	return fmt.Errorf("download %d is %s", id, state)
}

// stopLocked asks a running transfer to stop. The first stop reason wins.
func (m *Manager) stopLocked(id int64, kind stopKind) {
	if j := m.active[id]; j != nil {
		j.requestStop(kind)
	}
}

// pumpLocked starts queued downloads while slots are free.
func (m *Manager) pumpLocked() {
	if m.closed {
		return
	}
	for len(m.active) < m.settings.MaxActive {
		next, ok := m.nextQueuedLocked()
		if !ok {
			return
		}
		m.startLocked(next)
	}
}

// nextQueuedLocked returns the oldest queued download that is not already running.
func (m *Manager) nextQueuedLocked() (Download, bool) {
	rows, err := m.store.List(m.ctx)
	if err != nil {
		logger.Warn(logCategory, "could not read the queue: %v", err)
		return Download{}, false
	}
	var next Download
	found := false
	for _, d := range rows {
		if d.State != StateQueued || m.active[d.ID] != nil {
			continue
		}
		if !found || d.ID < next.ID {
			next = d
			found = true
		}
	}
	return next, found
}

func (m *Manager) startLocked(d Download) {
	saved, err := m.store.Segments(m.ctx, d.ID)
	if err != nil {
		logger.Warn(logCategory, "could not load segments of download %d: %v", d.ID, err)
	}
	jobCtx, cancel := context.WithCancel(m.ctx)
	j := newJob(d, saved)
	j.cancel = cancel
	m.active[d.ID] = j

	if err := m.updateLocked(m.ctx, d.ID, startPatch(time.Now())); err != nil {
		logger.Warn(logCategory, "could not mark download %d as running: %v", d.ID, err)
	}
	t := &transferer{client: m.client, limiter: m.limiter, hosts: m.hosts, retries: m.settings.Retries}
	m.jobs.Add(1)
	go m.runJob(jobCtx, j, t)
}
