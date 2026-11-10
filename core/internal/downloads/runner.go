package downloads

import (
	"context"
	"os"
	"time"

	"light-launcher/core/internal/logger"
)

const (
	progressInterval = 250 * time.Millisecond
	persistInterval  = 2 * time.Second
)

// runJob drives one transfer and sends progress until it ends.
func (m *Manager) runJob(ctx context.Context, j *job, t *transferer) {
	defer m.jobs.Done()
	errs := make(chan error, 1)
	go func() {
		errs <- t.run(ctx, j)
	}()

	progressTick := time.NewTicker(progressInterval)
	defer progressTick.Stop()
	persistTick := time.NewTicker(persistInterval)
	defer persistTick.Stop()
	for {
		select {
		case err := <-errs:
			m.finish(j, err)
			return
		case <-progressTick.C:
			m.emitProgress(j)
		case <-persistTick.C:
			m.persist(j)
		}
	}
}

// finish records how a run ended, frees its slot and starts the next download.
func (m *Manager) finish(j *job, err error) {
	j.cancel()
	m.persist(j)

	m.mu.Lock()
	id := j.row.ID
	if m.active[id] == j {
		delete(m.active, id)
	}
	installNow := false
	if j.stopReason() == stopNone {
		if err == nil {
			m.completeLocked(j)
			installNow = j.row.Install && m.installer != nil
		} else {
			m.failLocked(j, err)
		}
	}
	close(j.done)
	m.pumpLocked()
	m.mu.Unlock()

	if installNow {
		m.startInstall(j.row)
	}
}

// completeLocked marks a finished run as done.
func (m *Manager) completeLocked(j *job) {
	now := time.Now()
	p := j.progress(now)
	id := j.row.ID
	patch := statePatch(StateDone, "")
	patch.FinishedAt = &now
	patch.BytesDone = &p.received
	patch.Size = &p.total
	if err := m.updateLocked(context.Background(), id, patch); err != nil {
		logger.Warn(logCategory, "could not mark download %d as done: %v", id, err)
		return
	}
	logger.Info(logCategory, "download %d finished", id)
}

func (m *Manager) failLocked(j *job, err error) {
	id := j.row.ID
	message := plainError(err)
	logger.Warn(logCategory, "download %d failed: %s", id, message)
	if updateErr := m.updateLocked(context.Background(), id, statePatch(StateError, message)); updateErr != nil {
		logger.Warn(logCategory, "could not save the failure of download %d: %v", id, updateErr)
	}
}

// persist saves progress and segments so a restart or a pause can resume the file.
func (m *Manager) persist(j *job) {
	segments := j.snapshotSegments()
	if len(segments) == 0 {
		return
	}
	p := j.progress(time.Now())
	id := j.row.ID
	ctx := context.Background()
	connections := p.connections
	patch := Patch{BytesDone: &p.received, Size: &p.total, Connections: &connections}
	if etag := j.etagValue(); etag != "" {
		patch.ETag = &etag
	}
	if err := m.store.Update(ctx, id, patch); err != nil {
		logger.Warn(logCategory, "could not save progress of download %d: %v", id, err)
	}
	if err := m.store.SaveSegments(ctx, id, segments); err != nil {
		logger.Warn(logCategory, "could not save segments of download %d: %v", id, err)
	}
}

// emitProgress sends a progress tick for a running download. A download that is stopping sends nothing.
func (m *Manager) emitProgress(j *job) {
	if j.stopReason() != stopNone {
		return
	}
	p := j.progress(time.Now())
	m.publish(ProgressEvent{
		ID:       j.row.ID,
		State:    StateDownloading,
		Received: p.received,
		Total:    p.total,
		Speed:    p.speed,
	})
}

// updateLocked saves a patch and, when it changes the state, tells the UI.
func (m *Manager) updateLocked(ctx context.Context, id int64, p Patch) error {
	if err := m.store.Update(ctx, id, p); err != nil {
		return err
	}
	if p.State != nil {
		m.publishStateLocked(ctx, id)
	}
	return nil
}

func statePatch(state State, errText string) Patch {
	return Patch{State: &state, Error: &errText}
}

func startPatch(now time.Time) Patch {
	p := statePatch(StateDownloading, "")
	p.StartedAt = &now
	return p
}

// startInstall runs the installer on its own goroutine so the download slot is already free.
func (m *Manager) startInstall(d Download) {
	m.installs.Add(1)
	go func() {
		defer m.installs.Done()
		if err := m.installer.Install(m.ctx, d.ID, d.GameID, d.Dest); err != nil {
			logger.Warn(logCategory, "install of download %d failed: %v", d.ID, err)
			return
		}
		if !m.Settings().KeepArchives {
			removeFiles(d.Dest)
		}
	}()
}

// removeFiles deletes the finished file and its part file. A missing file is not an error.
func removeFiles(dest string) {
	for _, p := range []string{dest, dest + ".part"} {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			logger.Warn(logCategory, "could not delete %s: %v", p, err)
		}
	}
}
