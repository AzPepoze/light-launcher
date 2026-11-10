package runtimesvc

import (
	"path/filepath"
	"strings"
	"sync"
	"time"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/types"
)

// activityMu serializes reads and writes of activity.json between launches and session scans.
var activityMu sync.Mutex

type activityStore struct {
	Games map[string]types.GameActivity `json:"games"`
}

func activityStorePath() string {
	return filepath.Join(config.GetBaseDirectory(), "activity.json")
}

// activityKey identifies a game in the store: its cleaned path, lower case.
func activityKey(gamePath string) string {
	return strings.ToLower(filepath.Clean(gamePath))
}

func loadActivityStore() activityStore {
	empty := activityStore{Games: map[string]types.GameActivity{}}
	storePath := activityStorePath()
	if !pathExists(storePath) {
		return empty
	}
	store, err := loadJSON[activityStore](storePath)
	if err != nil {
		logger.Warn("Activity", "Failed to load activity store: %v", err)
		return empty
	}
	if store.Games == nil {
		return empty
	}
	return store
}

func saveActivityStore(store activityStore) error {
	return saveJSON(activityStorePath(), store)
}

// recordLaunch marks the game as played now and starts its active session.
func recordLaunch(options types.LaunchOptions) error {
	if playtimeTrackingOff() {
		return nil
	}

	now := time.Now().UnixMilli()
	gamePath := filepath.Clean(firstNonEmpty(options.GamePath, options.LauncherPath))

	activityMu.Lock()
	defer activityMu.Unlock()

	store := loadActivityStore()
	key := activityKey(gamePath)
	previous, hasPrevious := store.Games[key]

	activeSince := now
	if hasPrevious && previous.ActiveSince != nil {
		activeSince = *previous.ActiveSince
	}
	entry := types.GameActivity{
		GamePath:             gamePath,
		GameName:             firstNonEmpty(options.Name, nameWithoutExt(filepath.Base(gamePath))),
		ProfileID:            firstNonEmpty(options.ID, previous.ProfileID),
		CustomIconPath:       firstNonEmpty(options.CustomIconPath, previous.CustomIconPath),
		LastPlayedAt:         now,
		TotalPlaytimeSeconds: previous.TotalPlaytimeSeconds,
		ActiveSince:          &activeSince,
	}
	store.Games[key] = entry
	return saveActivityStore(store)
}

// syncActivitySessions clears stale active markers and starts markers for sessions that are running.
// The Go instance writes the final playtime when a game exits.
func syncActivitySessions(sessions []types.RunningSession) error {
	activityMu.Lock()
	defer activityMu.Unlock()

	store := loadActivityStore()
	if playtimeTrackingOff() {
		changed := clearActiveMarkers(store, func(string) bool { return false })
		return saveIfChanged(store, changed)
	}

	running := map[string]bool{}
	for _, session := range sessions {
		running[activityKey(session.GamePath)] = true
	}
	changed := clearActiveMarkers(store, func(key string) bool { return running[key] })
	for _, session := range sessions {
		if addSessionMarker(store, session) {
			changed = true
		}
	}
	return saveIfChanged(store, changed)
}

// clearActiveMarkers removes the active marker from games that are not kept running.
func clearActiveMarkers(store activityStore, keepRunning func(key string) bool) bool {
	changed := false
	for key, activity := range store.Games {
		if !isActive(activity) || keepRunning(key) {
			continue
		}
		activity.ActiveSince = nil
		store.Games[key] = activity
		changed = true
	}
	return changed
}

// addSessionMarker creates or updates a game entry for a running session. It reports whether the store changed.
func addSessionMarker(store activityStore, session types.RunningSession) bool {
	key := activityKey(session.GamePath)
	startedAt := session.StartedAt
	if startedAt == 0 {
		startedAt = time.Now().UnixMilli()
	}

	existing, found := store.Games[key]
	if !found {
		store.Games[key] = types.GameActivity{
			GamePath:     session.GamePath,
			GameName:     session.GameName,
			LastPlayedAt: startedAt,
			ActiveSince:  &startedAt,
		}
		return true
	}
	if isActive(existing) {
		return false
	}
	existing.ActiveSince = &startedAt
	existing.LastPlayedAt = max(existing.LastPlayedAt, startedAt)
	store.Games[key] = existing
	return true
}

// playtimeTrackingOff is true only when the setting is explicitly false, as in the Electron app.
func playtimeTrackingOff() bool {
	trackPlaytime := loadAppSettings().TrackPlaytime
	return trackPlaytime != nil && !*trackPlaytime
}

func isActive(activity types.GameActivity) bool {
	return activity.ActiveSince != nil && *activity.ActiveSince != 0
}

func saveIfChanged(store activityStore, changed bool) error {
	if !changed {
		return nil
	}
	return saveActivityStore(store)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
