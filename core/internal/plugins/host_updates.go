package plugins

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"light-launcher/core/internal/plugins/updater"
)

// ErrUpdateNotAvailable means the requested source is not the release the last check found.
var ErrUpdateNotAvailable = errors.New("no update available, check for updates first")

// UpdateChecker finds the newest release of one plugin repository.
type UpdateChecker interface {
	Check(ctx context.Context, req updater.Request) (updater.Result, error)
}

// PluginUpdateStatus is the cached update check of one plugin.
type PluginUpdateStatus struct {
	ID               string `json:"id"`
	InstalledVersion string `json:"installedVersion"`
	Available        bool   `json:"available"`
	Version          string `json:"version,omitempty"`
	AssetURL         string `json:"assetUrl,omitempty"`
	SHA256URL        string `json:"sha256Url,omitempty"`
	SHA256           string `json:"sha256,omitempty"`
	Notes            string `json:"notes,omitempty"`
	PublishedAt      string `json:"publishedAt,omitempty"`
	Error            string `json:"error,omitempty"`
	// Code is a stable reason when the check failed: not_found, rate_limited, untrusted_host or check_failed.
	Code string `json:"code,omitempty"`
}

// PluginUpdatesStatus is the latest update check of every plugin.
type PluginUpdatesStatus struct {
	CheckedAt string               `json:"checkedAt,omitempty"`
	Plugins   []PluginUpdateStatus `json:"plugins"`
}

// updateTracker keeps the last check results in memory. Checks themselves run one at a time.
type updateTracker struct {
	mu        sync.Mutex
	checkMu   sync.Mutex
	checker   UpdateChecker
	checkedAt time.Time
	entries   map[string]PluginUpdateStatus
}

func newUpdateTracker(checker UpdateChecker) *updateTracker {
	return &updateTracker{checker: checker, entries: map[string]PluginUpdateStatus{}}
}

// CheckUpdates checks one plugin, or every plugin when id is empty.
func (h *Host) CheckUpdates(ctx context.Context, id string) (PluginUpdatesStatus, error) {
	h.updates.checkMu.Lock()
	defer h.updates.checkMu.Unlock()
	targets, err := h.updateTargets(id)
	if err != nil {
		return PluginUpdatesStatus{}, err
	}
	entries := make([]PluginUpdateStatus, 0, len(targets))
	for _, target := range targets {
		entries = append(entries, h.checkOne(ctx, target))
	}
	status := h.updates.store(entries, id != "")
	h.emitUpdates(status)
	return status, nil
}

// UpdatesStatus returns the cached results without asking GitHub.
func (h *Host) UpdatesStatus() PluginUpdatesStatus {
	return h.updates.snapshot()
}

type updateTarget struct {
	id         string
	version    string
	repository string
	err        error
}

func (h *Host) updateTargets(id string) ([]updateTarget, error) {
	if id != "" {
		installed, err := h.loadInstalled(id)
		if err != nil {
			return nil, err
		}
		return []updateTarget{{id: id, version: installed.version, repository: installed.manifest.Repository}}, nil
	}
	ids, err := h.layout.installedIDs()
	if err != nil {
		return nil, err
	}
	sort.Strings(ids)
	targets := make([]updateTarget, 0, len(ids))
	for _, pluginID := range ids {
		installed, err := h.loadInstalled(pluginID)
		if err != nil {
			targets = append(targets, updateTarget{id: pluginID, err: err})
			continue
		}
		targets = append(targets, updateTarget{
			id:         pluginID,
			version:    installed.version,
			repository: installed.manifest.Repository,
		})
	}
	return targets, nil
}

func (h *Host) checkOne(ctx context.Context, target updateTarget) PluginUpdateStatus {
	entry := PluginUpdateStatus{ID: target.id, InstalledVersion: target.version}
	if target.err != nil {
		entry.Error = target.err.Error()
		entry.Code = updateErrorCode(target.err)
		return entry
	}
	if target.repository == "" {
		return entry
	}
	result, err := h.updates.checker.Check(ctx, updater.Request{
		ID:         target.id,
		Repository: target.repository,
		Version:    target.version,
	})
	if err != nil {
		entry.Error = err.Error()
		entry.Code = updateErrorCode(err)
		return entry
	}
	entry.Available = result.Available
	entry.Version = result.Version
	entry.AssetURL = result.AssetURL
	entry.SHA256URL = result.SHA256URL
	entry.SHA256 = result.SHA256
	entry.Notes = result.Notes
	entry.PublishedAt = result.PublishedAt
	return entry
}

// updateErrorCode turns a check error into a code the UI can act on without reading the text.
func updateErrorCode(err error) string {
	var limit *updater.RateLimitError
	switch {
	case errors.As(err, &limit), errors.Is(err, updater.ErrRateLimited):
		return "rate_limited"
	case errors.Is(err, updater.ErrNoRelease):
		return "not_found"
	case errors.Is(err, updater.ErrUntrustedHost):
		return "untrusted_host"
	default:
		return "check_failed"
	}
}

// store replaces the cache on a full check and merges one plugin otherwise.
func (t *updateTracker) store(entries []PluginUpdateStatus, replace bool) PluginUpdatesStatus {
	t.mu.Lock()
	defer t.mu.Unlock()
	if replace {
		t.entries = map[string]PluginUpdateStatus{}
	}
	for _, entry := range entries {
		t.entries[entry.ID] = entry
	}
	t.checkedAt = time.Now().UTC()
	return t.snapshotLocked()
}

func (t *updateTracker) snapshot() PluginUpdatesStatus {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.snapshotLocked()
}

func (t *updateTracker) snapshotLocked() PluginUpdatesStatus {
	status := PluginUpdatesStatus{Plugins: make([]PluginUpdateStatus, 0, len(t.entries))}
	if !t.checkedAt.IsZero() {
		status.CheckedAt = t.checkedAt.Format(time.RFC3339)
	}
	ids := make([]string, 0, len(t.entries))
	for id := range t.entries {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		status.Plugins = append(status.Plugins, t.entries[id])
	}
	return status
}

// updateFor returns the cached release when source is exactly the asset the check found.
func (t *updateTracker) updateFor(id, source string) (PluginUpdateStatus, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	entry, ok := t.entries[id]
	if !ok || !entry.Available || entry.AssetURL != source {
		return PluginUpdateStatus{}, false
	}
	return entry, true
}

func (t *updateTracker) forget(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, id)
}

func (h *Host) emitUpdates(status PluginUpdatesStatus) {
	if h.deps.events != nil {
		h.deps.events.Emit("plugin-updates", status)
	}
}

// forgetUpdate drops a stale result after the installed version changed.
func (h *Host) forgetUpdate(id string) {
	h.updates.forget(id)
	h.emitUpdates(h.updates.snapshot())
}

// updateEntry checks that source is the release found for the installed version.
func (h *Host) updateEntry(id, source string) (PluginUpdateStatus, error) {
	installed, err := h.loadInstalled(id)
	if err != nil {
		return PluginUpdateStatus{}, err
	}
	entry, ok := h.updates.updateFor(id, source)
	if !ok || entry.InstalledVersion != installed.version {
		return PluginUpdateStatus{}, ErrUpdateNotAvailable
	}
	return entry, nil
}

// updateSource returns the source to install from and the checksum to verify when known.
func (h *Host) updateSource(id, source, sha256 string) (string, string, error) {
	if source == "" {
		source = h.install.state.get(id).Source
		if source == "" {
			return "", "", ErrNoSource
		}
		return source, sha256, nil
	}
	entry, err := h.updateEntry(id, source)
	if err != nil {
		return "", "", err
	}
	if sha256 == "" {
		sha256 = entry.SHA256
	}
	return source, sha256, nil
}

func (h *Host) startUpdateTarget(ctx context.Context, id string, keepRunning bool) error {
	if !h.install.state.get(id).Enabled {
		return nil
	}
	supervisor, err := h.supervisorFor(id)
	if err != nil {
		return err
	}
	if err := supervisor.Start(ctx); err != nil {
		return err
	}
	if !keepRunning {
		h.stopPlugin(id)
	}
	return nil
}

func (h *Host) rollbackFailedUpdate(ctx context.Context, id string, wasRunning bool, cause error) error {
	h.stopPlugin(id)
	if err := h.install.Rollback(id); err != nil {
		return fmt.Errorf("updated plugin %s failed to start: %w; rollback failed: %v", id, cause, err)
	}
	if wasRunning {
		if _, err := h.supervisorFor(id); err == nil {
			if err := h.startUpdateTarget(ctx, id, true); err != nil {
				return fmt.Errorf("updated plugin %s failed to start: %w; the previous version did not start either: %v", id, cause, err)
			}
		}
	}
	h.emitState(id)
	return fmt.Errorf("updated plugin %s failed to start, rolled back: %w", id, cause)
}
