package appsvc

import (
	"os"
	"sync"

	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/types"
)

// settingsCache holds the parsed settings.json. It is keyed by path, so a new base directory reloads.
var settingsCache struct {
	mu    sync.Mutex
	path  string
	value *types.AppSettings
}

// loadAppSettings returns a copy of the settings. Load problems fall back to defaults.
func loadAppSettings() types.AppSettings {
	settingsCache.mu.Lock()
	defer settingsCache.mu.Unlock()
	return cloneSettings(loadSettingsLocked())
}

// saveAppSettings writes settings.json and replaces the cached copy.
func saveAppSettings(settings types.AppSettings) error {
	settingsCache.mu.Lock()
	defer settingsCache.mu.Unlock()
	return writeSettingsLocked(&settings)
}

// updateAppSettings applies a change to a copy of the settings under the lock, then saves it.
func updateAppSettings(change func(settings *types.AppSettings) error) error {
	settingsCache.mu.Lock()
	defer settingsCache.mu.Unlock()
	settings := cloneSettings(loadSettingsLocked())
	if err := change(&settings); err != nil {
		return err
	}
	return writeSettingsLocked(&settings)
}

// loadSettingsLocked must be called with settingsCache.mu held.
func loadSettingsLocked() *types.AppSettings {
	path := settingsFilePath()
	if settingsCache.value != nil && settingsCache.path == path {
		return settingsCache.value
	}
	value := readSettingsFile(path)
	settingsCache.path = path
	settingsCache.value = value
	return value
}

// writeSettingsLocked must be called with settingsCache.mu held.
func writeSettingsLocked(settings *types.AppSettings) error {
	path := settingsFilePath()
	if err := saveJSON(path, settings); err != nil {
		return err
	}
	copied := cloneSettings(settings)
	settingsCache.path = path
	settingsCache.value = &copied
	return nil
}

func readSettingsFile(path string) *types.AppSettings {
	if _, err := os.Stat(path); err != nil {
		return defaultSettings()
	}
	settings, err := loadAndNormalizeSettings(path)
	if err != nil {
		logger.Error("Config", "Error loading app settings: %v", err)
		return defaultSettings()
	}
	return settings
}

// loadAndNormalizeSettings rewrites the file only when normalization changed something.
func loadAndNormalizeSettings(path string) (*types.AppSettings, error) {
	settings, err := loadJSON[types.AppSettings](path)
	if err != nil {
		return nil, err
	}
	if !normalizeSettings(&settings) {
		return &settings, nil
	}
	if err := saveJSON(path, settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

func cloneSettings(settings *types.AppSettings) types.AppSettings {
	copied := *settings
	copied.ScanFolders = cloneStrings(settings.ScanFolders)
	copied.Blacklist = cloneStrings(settings.Blacklist)
	copied.ScanFolderConfigs = cloneScanFolderConfigs(settings.ScanFolderConfigs)
	return copied
}

func cloneScanFolderConfigs(configs []types.ScanFolderConfig) []types.ScanFolderConfig {
	if configs == nil {
		return nil
	}
	copied := make([]types.ScanFolderConfig, len(configs))
	for i, cfg := range configs {
		cfg.ExcludeNames = cloneStrings(cfg.ExcludeNames)
		copied[i] = cfg
	}
	return copied
}
