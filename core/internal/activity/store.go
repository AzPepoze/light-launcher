package activity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"light-launcher/core/internal/config"
)

type gameActivity struct {
	GamePath             string `json:"gamePath"`
	GameName             string `json:"gameName"`
	ProfileID            string `json:"profileId,omitempty"`
	CustomIconPath       string `json:"customIconPath,omitempty"`
	LastPlayedAt         int64  `json:"lastPlayedAt"`
	TotalPlaytimeSeconds int64  `json:"totalPlaytimeSeconds"`
	SessionCount         int    `json:"sessionCount"`
	ActiveSince          *int64 `json:"activeSince,omitempty"`
}

type store struct {
	Games map[string]*gameActivity `json:"games"`
}

type appSettings struct {
	TrackPlaytime *bool `json:"TrackPlaytime"`
}

func trackingStillEnabled() bool {
	data, err := os.ReadFile(filepath.Join(config.GetBaseDirectory(), "settings.json"))
	if err != nil {
		return true
	}
	var settings appSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return true
	}
	return settings.TrackPlaytime == nil || *settings.TrackPlaytime
}

// Finalize records the exact game lifetime from the long-lived instance process.
// This keeps playtime correct even when the Electron launcher quits immediately after launch.
func Finalize(gamePath, gameName string, startedAt, endedAt time.Time) error {
	if !trackingStillEnabled() {
		return nil
	}

	storePath := filepath.Join(config.GetBaseDirectory(), "activity.json")
	current := store{Games: map[string]*gameActivity{}}
	if data, err := os.ReadFile(storePath); err == nil {
		_ = json.Unmarshal(data, &current)
		if current.Games == nil {
			current.Games = map[string]*gameActivity{}
		}
	}

	cleanPath := filepath.Clean(gamePath)
	key := strings.ToLower(cleanPath)
	entry := current.Games[key]
	if entry == nil {
		entry = &gameActivity{
			GamePath:     cleanPath,
			GameName:     gameName,
			LastPlayedAt: startedAt.UnixMilli(),
			SessionCount: 1,
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
