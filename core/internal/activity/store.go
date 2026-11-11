// Package activity owns activity.json: the playtime and running-session store.
package activity

import (
	"errors"
	"sort"
	"sync"
	"time"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/store"
	"light-launcher/core/internal/types"
)

// Version is the activity.json layout version.
const Version = 1

// Settings gives the store access to the playtime tracking flag.
type Settings interface {
	Load() types.AppSettings
}

// Store reads and writes activity.json. One store serializes its own read-modify-write cycles.
type Store struct {
	mu       sync.Mutex
	file     *store.File
	settings Settings
	now      func() time.Time
}

type activityFile struct {
	Games map[string]*types.GameActivity `json:"games"`
}

// New opens the activity store at path.
func New(path string, settings Settings) *Store {
	return &Store{
		file:     store.NewFile(path, Version),
		settings: settings,
		now:      time.Now,
	}
}

// Enabled reports whether playtime tracking is on. A missing setting means on.
func (s *Store) Enabled() bool {
	trackPlaytime := s.settings.Load().TrackPlaytime
	return trackPlaytime == nil || *trackPlaytime
}

// RecordLaunch marks a game as started now, keeping its earlier playtime.
func (s *Store) RecordLaunch(options types.LaunchOptions) error {
	if !s.Enabled() {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now().UnixMilli()
	gamePath := config.CleanPath(config.FirstNonEmpty(options.GamePath, options.LauncherPath))
	current := s.load()
	key := config.PathKey(gamePath)
	previous := current.Games[key]
	if previous == nil {
		previous = &types.GameActivity{}
	}

	current.Games[key] = &types.GameActivity{
		GamePath:             gamePath,
		GameName:             config.FirstNonEmpty(options.Name, config.NameWithoutExt(gamePath)),
		ProfileID:            config.FirstNonEmpty(options.ID, previous.ProfileID),
		CustomIconPath:       config.FirstNonEmpty(options.CustomIconPath, previous.CustomIconPath),
		LastPlayedAt:         now,
		TotalPlaytimeSeconds: previous.TotalPlaytimeSeconds,
		ActiveSince:          activeSinceOrNow(previous, now),
	}
	return s.save(current)
}

// Sync clears stale "playing" markers and marks running sessions. It returns the sorted list.
func (s *Store) Sync(sessions []types.RunningSession) ([]types.GameActivity, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current := s.load()
	enabled := s.Enabled()
	now := s.now().UnixMilli()
	running := map[string]bool{}
	if enabled {
		for _, session := range sessions {
			running[config.PathKey(session.GamePath)] = true
		}
	}

	changed := clearStaleMarkers(current, running)
	if enabled && markRunningSessions(current, sessions, now) {
		changed = true
	}
	if changed {
		if err := s.save(current); err != nil {
			return nil, err
		}
	}
	return sortedActivities(current), nil
}

// List returns the stored activity, most recently played first.
func (s *Store) List() []types.GameActivity {
	s.mu.Lock()
	defer s.mu.Unlock()
	return sortedActivities(s.load())
}

// Finalize persists playtime using the instance's actual exit time.
func (s *Store) Finalize(gamePath, gameName string, startedAt, endedAt time.Time) error {
	if !s.Enabled() {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cleanPath := config.CleanPath(gamePath)
	key := config.PathKey(cleanPath)
	current := s.load()
	entry := current.Games[key]
	if entry == nil {
		entry = &types.GameActivity{
			GamePath:     cleanPath,
			GameName:     gameName,
			LastPlayedAt: startedAt.UnixMilli(),
		}
		current.Games[key] = entry
	}

	actualStart := startedAt
	if isActive(entry) {
		actualStart = time.UnixMilli(*entry.ActiveSince)
	}
	if endedAt.Before(actualStart) {
		endedAt = actualStart
	}

	entry.GamePath = cleanPath
	if gameName != "" {
		entry.GameName = gameName
	}
	entry.LastPlayedAt = endedAt.UnixMilli()
	entry.TotalPlaytimeSeconds += int64(endedAt.Sub(actualStart).Seconds())
	entry.ActiveSince = nil
	return s.save(current)
}

func (s *Store) load() activityFile {
	current := activityFile{Games: map[string]*types.GameActivity{}}
	found, err := s.file.Load(&current)
	if err != nil {
		if errors.Is(err, store.ErrVersion) {
			logger.Warn("Activity", "Resetting activity with an unknown version: %v", err)
		} else {
			logger.Warn("Activity", "Failed to load activity store: %v", err)
		}
		return activityFile{Games: map[string]*types.GameActivity{}}
	}
	if !found || current.Games == nil {
		current.Games = map[string]*types.GameActivity{}
	}
	return current
}

func (s *Store) save(current activityFile) error {
	return s.file.Save(current)
}

func activeSinceOrNow(previous *types.GameActivity, now int64) *int64 {
	if isActive(previous) {
		return previous.ActiveSince
	}
	return &now
}

// clearStaleMarkers removes the active marker from games that are not running.
func clearStaleMarkers(current activityFile, running map[string]bool) bool {
	changed := false
	for key, activity := range current.Games {
		if isActive(activity) && !running[key] {
			activity.ActiveSince = nil
			changed = true
		}
	}
	return changed
}

// markRunningSessions adds sessions that are missing or not marked active, as after a launcher restart.
func markRunningSessions(current activityFile, sessions []types.RunningSession, now int64) bool {
	changed := false
	for _, session := range sessions {
		startedAt := session.StartedAt
		if startedAt == 0 {
			startedAt = now
		}
		key := config.PathKey(session.GamePath)
		existing := current.Games[key]
		if existing == nil {
			current.Games[key] = &types.GameActivity{
				GamePath:     config.CleanPath(session.GamePath),
				GameName:     session.GameName,
				LastPlayedAt: startedAt,
				ActiveSince:  &startedAt,
			}
			changed = true
		} else if !isActive(existing) {
			existing.ActiveSince = &startedAt
			existing.LastPlayedAt = max(existing.LastPlayedAt, startedAt)
			changed = true
		}
	}
	return changed
}

func isActive(activity *types.GameActivity) bool {
	return activity != nil && activity.ActiveSince != nil && *activity.ActiveSince != 0
}

// sortedActivities orders by last played time, then by key so the result is stable.
func sortedActivities(current activityFile) []types.GameActivity {
	keys := make([]string, 0, len(current.Games))
	for key := range current.Games {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	sort.SliceStable(keys, func(i, j int) bool {
		return current.Games[keys[i]].LastPlayedAt > current.Games[keys[j]].LastPlayedAt
	})

	activities := make([]types.GameActivity, 0, len(keys))
	for _, key := range keys {
		activities = append(activities, *current.Games[key])
	}
	return activities
}
