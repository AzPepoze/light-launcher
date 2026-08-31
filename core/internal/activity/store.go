package activity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

type store struct {
	Games map[string]*types.GameActivity `json:"games"`
}

func TrackingEnabled() bool {
	data, err := os.ReadFile(filepath.Join(config.GetBaseDirectory(), "settings.json"))
	if err != nil {
		return true
	}
	var settings types.AppSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return true
	}
	return settings.TrackPlaytime == nil || *settings.TrackPlaytime
}

// Finalize persists playtime using the instance's actual exit time.
func Finalize(gamePath, gameName string, startedAt, endedAt time.Time) error {
	if !TrackingEnabled() {
		return nil
	}

	storePath := filepath.Join(config.GetBaseDirectory(), "activity.json")
	current := store{Games: map[string]*types.GameActivity{}}
	if data, err := os.ReadFile(storePath); err == nil {
		_ = json.Unmarshal(data, &current)
		if current.Games == nil {
			current.Games = map[string]*types.GameActivity{}
		}
	}

	cleanPath := filepath.Clean(gamePath)
	key := strings.ToLower(cleanPath)
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
	if entry.ActiveSince != nil && *entry.ActiveSince > 0 {
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

	if err := os.MkdirAll(filepath.Dir(storePath), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return err
	}
	tempPath := storePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tempPath, storePath)
}
