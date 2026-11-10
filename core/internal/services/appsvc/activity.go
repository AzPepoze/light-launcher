package appsvc

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/types"
)

// activityStore matches the activity.json layout that activity.Finalize writes.
type activityStore struct {
	Games map[string]*types.GameActivity `json:"games"`
}

// activityMu serializes read-modify-write cycles on activity.json inside this process.
var activityMu sync.Mutex

func activityStorePath() string {
	return filepath.Join(config.GetBaseDirectory(), "activity.json")
}

func activityKey(gamePath string) string {
	return strings.ToLower(normalizePath(gamePath))
}

func loadActivityStore() activityStore {
	path := activityStorePath()
	if _, err := os.Stat(path); err != nil {
		return emptyActivityStore()
	}
	store, err := loadJSON[activityStore](path)
	if err != nil {
		logger.Warn("Activity", "Failed to load activity store: %v", err)
		return emptyActivityStore()
	}
	if store.Games == nil {
		return emptyActivityStore()
	}
	return store
}

func emptyActivityStore() activityStore {
	return activityStore{Games: map[string]*types.GameActivity{}}
}

func saveActivityStore(store activityStore) error {
	return saveJSON(activityStorePath(), store)
}

// recordLaunch marks a game as started now, keeping its earlier playtime.
func recordLaunch(options types.LaunchOptions) error {
	if !playtimeTrackingEnabled() {
		return nil
	}
	activityMu.Lock()
	defer activityMu.Unlock()

	now := time.Now().UnixMilli()
	gamePath := normalizePath(firstNonEmpty(options.GamePath, options.LauncherPath))
	store := loadActivityStore()
	key := activityKey(gamePath)
	previous := store.Games[key]
	if previous == nil {
		previous = &types.GameActivity{}
	}

	store.Games[key] = &types.GameActivity{
		GamePath:             gamePath,
		GameName:             firstNonEmpty(options.Name, fileNameWithoutExt(gamePath)),
		ProfileID:            firstNonEmpty(options.ID, previous.ProfileID),
		CustomIconPath:       firstNonEmpty(options.CustomIconPath, previous.CustomIconPath),
		LastPlayedAt:         now,
		TotalPlaytimeSeconds: previous.TotalPlaytimeSeconds,
		ActiveSince:          activeSinceOrNow(previous, now),
	}
	return saveActivityStore(store)
}

func activeSinceOrNow(previous *types.GameActivity, now int64) *int64 {
	if isActive(previous) {
		return previous.ActiveSince
	}
	return &now
}

func playtimeTrackingEnabled() bool {
	settings := loadAppSettings()
	return settings.TrackPlaytime == nil || *settings.TrackPlaytime
}

// syncActivities clears stale "playing" markers and marks running sessions. It returns the sorted list.
func syncActivities(sessions []types.RunningSession) ([]types.GameActivity, error) {
	activityMu.Lock()
	defer activityMu.Unlock()

	store := loadActivityStore()
	now := time.Now().UnixMilli()
	running := runningKeys(sessions)
	changed := clearStaleMarkers(store, running)
	if playtimeTrackingEnabled() {
		if markRunningSessions(store, sessions, now) {
			changed = true
		}
	}

	if changed {
		if err := saveActivityStore(store); err != nil {
			return nil, err
		}
	}
	return sortedActivities(store), nil
}

// runningKeys returns the sessions to keep active. With tracking off it is empty, so every marker is cleared.
func runningKeys(sessions []types.RunningSession) map[string]bool {
	keys := map[string]bool{}
	if !playtimeTrackingEnabled() {
		return keys
	}
	for _, session := range sessions {
		keys[activityKey(session.GamePath)] = true
	}
	return keys
}

// clearStaleMarkers removes the active marker from games that are not running.
func clearStaleMarkers(store activityStore, running map[string]bool) bool {
	changed := false
	for key, activity := range store.Games {
		if isActive(activity) && !running[key] {
			activity.ActiveSince = nil
			changed = true
		}
	}
	return changed
}

// markRunningSessions adds sessions that are missing or not marked active, as after a launcher restart.
func markRunningSessions(store activityStore, sessions []types.RunningSession, now int64) bool {
	changed := false
	for _, session := range sessions {
		startedAt := session.StartedAt
		if startedAt == 0 {
			startedAt = now
		}
		key := activityKey(session.GamePath)
		existing := store.Games[key]
		if existing == nil {
			store.Games[key] = &types.GameActivity{
				GamePath:     session.GamePath,
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

// getActivities returns the stored activity, most recently played first.
func getActivities() []types.GameActivity {
	activityMu.Lock()
	defer activityMu.Unlock()
	return sortedActivities(loadActivityStore())
}

func isActive(activity *types.GameActivity) bool {
	return activity != nil && activity.ActiveSince != nil && *activity.ActiveSince != 0
}

// sortedActivities orders by last played time, then by key so the result is stable.
func sortedActivities(store activityStore) []types.GameActivity {
	keys := make([]string, 0, len(store.Games))
	for key := range store.Games {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	sort.SliceStable(keys, func(i, j int) bool {
		return store.Games[keys[i]].LastPlayedAt > store.Games[keys[j]].LastPlayedAt
	})

	activities := make([]types.GameActivity, 0, len(keys))
	for _, key := range keys {
		activities = append(activities, *store.Games[key])
	}
	return activities
}
