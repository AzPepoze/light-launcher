package config

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/store"
	"light-launcher/core/internal/types"
)

// DataVersion is written into settings, game and prefix config files.
const DataVersion = 1

// SettingsLoader gives the activity store access to the current settings.
type SettingsLoader interface {
	Load() types.AppSettings
}

// SettingsStore owns settings.json. Reads always hit the file, so a change made outside
// this process is picked up on the next call.
type SettingsStore struct {
	mu   sync.Mutex
	file *store.File
}

// NewSettingsStore opens settings.json at path.
func NewSettingsStore(path string) *SettingsStore {
	return &SettingsStore{file: store.NewFile(path, DataVersion)}
}

// Load returns a copy of the settings. A missing, broken or old file gives the defaults.
func (s *SettingsStore) Load() types.AppSettings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

// Save writes settings.json after checking the values the UI can change.
func (s *SettingsStore) Save(settings types.AppSettings) error {
	if err := validateAppSettings(settings); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.file.Save(settings)
}

// validateAppSettings checks the settings that the UI can change.
func validateAppSettings(settings types.AppSettings) error {
	minutes := settings.PluginIdleTimeoutMinutes
	if minutes == nil {
		return nil
	}
	if *minutes < 0 {
		return &api.FieldsError{Fields: []api.FieldError{{
			Field:   "pluginIdleTimeoutMinutes",
			Message: "must be at least 0",
		}}}
	}
	if *minutes > maxPluginIdleMinutes {
		return &api.FieldsError{Fields: []api.FieldError{{
			Field:   "pluginIdleTimeoutMinutes",
			Message: fmt.Sprintf("must be at most %d", maxPluginIdleMinutes),
		}}}
	}
	return nil
}

// Update applies a change to a copy of the settings under the lock, then saves it.
func (s *SettingsStore) Update(change func(settings *types.AppSettings) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	settings := s.loadLocked()
	if err := change(&settings); err != nil {
		return err
	}
	return s.file.Save(settings)
}

func (s *SettingsStore) loadLocked() types.AppSettings {
	var settings types.AppSettings
	found, err := s.file.Load(&settings)
	if err != nil {
		if errors.Is(err, store.ErrVersion) {
			logger.Warn("Config", "Ignoring settings with an unknown version: %v", err)
		} else {
			logger.Error("Config", "Error loading app settings: %v", err)
		}
		return DefaultSettings()
	}
	if !found {
		return DefaultSettings()
	}
	if NormalizeSettings(&settings) {
		if err := s.file.Save(settings); err != nil {
			logger.Error("Config", "Failed to save normalized settings: %v", err)
		}
	}
	return settings
}

// NormalizeSettings fills missing lists, adds legacy scan folders and refreshes old default excludes.
// It reports whether the file needs saving.
func NormalizeSettings(settings *types.AppSettings) bool {
	if settings.ScanFolderConfigs == nil {
		settings.ScanFolderConfigs = []types.ScanFolderConfig{}
	}
	if settings.ScanFolders == nil {
		settings.ScanFolders = []string{}
	}
	if settings.Blacklist == nil {
		settings.Blacklist = []string{}
	}

	modified := addLegacyScanFolders(settings)
	if backfillExcludeNames(settings) {
		modified = true
	}
	if normalizePluginIdleTimeout(settings) {
		modified = true
	}
	return modified
}

// normalizePluginIdleTimeout fills a missing idle timeout and replaces values outside the allowed range.
func normalizePluginIdleTimeout(settings *types.AppSettings) bool {
	minutes := settings.PluginIdleTimeoutMinutes
	if minutes == nil || *minutes < 0 || *minutes > maxPluginIdleMinutes {
		settings.PluginIdleTimeoutMinutes = defaultPluginIdleTimeout()
		return true
	}
	return false
}

// addLegacyScanFolders turns plain ScanFolders entries into ScanFolderConfigs with default settings.
func addLegacyScanFolders(settings *types.AppSettings) bool {
	added := false
	for _, folder := range settings.ScanFolders {
		cleaned := NormalizePath(folder)
		if hasNormalizedScanFolder(settings.ScanFolderConfigs, cleaned) {
			continue
		}
		settings.ScanFolderConfigs = append(settings.ScanFolderConfigs, types.ScanFolderConfig{
			Path:         cleaned,
			Depth:        2,
			ExcludeNames: DefaultScanExcludeNames(),
		})
		added = true
	}
	return added
}

// backfillExcludeNames fills empty or legacy exclude lists. Custom lists are kept.
func backfillExcludeNames(settings *types.AppSettings) bool {
	changed := false
	for i := range settings.ScanFolderConfigs {
		cfg := &settings.ScanFolderConfigs[i]
		if len(cfg.ExcludeNames) == 0 || isLegacyExcludeList(cfg.ExcludeNames) {
			cfg.ExcludeNames = DefaultScanExcludeNames()
			changed = true
		}
	}
	return changed
}

func hasNormalizedScanFolder(configs []types.ScanFolderConfig, normalized string) bool {
	for _, cfg := range configs {
		if NormalizePath(cfg.Path) == normalized {
			return true
		}
	}
	return false
}

func isLegacyExcludeList(excludeNames []string) bool {
	normalized := normalizedExcludeNames(excludeNames)
	for _, legacy := range legacyExcludeSets {
		if len(legacy) == len(normalized) && containsAllLowered(legacy, normalized) {
			return true
		}
	}
	return false
}

// normalizedExcludeNames trims and lowercases names and drops blanks.
func normalizedExcludeNames(excludeNames []string) []string {
	normalized := []string{}
	for _, name := range excludeNames {
		if trimmed := strings.ToLower(strings.TrimSpace(name)); trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	return normalized
}

func containsAllLowered(names []string, list []string) bool {
	for _, name := range names {
		if !slices.Contains(list, strings.ToLower(name)) {
			return false
		}
	}
	return true
}
