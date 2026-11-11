package config

import (
	"errors"
	"path/filepath"

	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/store"
	"light-launcher/core/internal/types"
)

// PrefixStore owns the light-launcher.json file of each prefix.
type PrefixStore struct {
	settings SettingsLoader
}

// NewPrefixStore opens prefix configs below the folder chosen by the settings.
func NewPrefixStore(settings SettingsLoader) *PrefixStore {
	return &PrefixStore{settings: settings}
}

// BaseDir returns the folder that holds all prefixes, honoring the custom prefix folder setting.
func (s *PrefixStore) BaseDir() string {
	return PrefixBaseDir(s.settings.Load().CustomPrefixDir)
}

func (s *PrefixStore) ConfigPath(prefixName string) string {
	return filepath.Join(s.BaseDir(), prefixName, "light-launcher.json")
}

// Load returns the saved prefix config, or defaults when it is missing or unreadable.
func (s *PrefixStore) Load(prefixName string) types.LaunchOptions {
	var cfg types.LaunchOptions
	file := store.NewFile(s.ConfigPath(prefixName), DataVersion)
	found, err := file.Load(&cfg)
	if err != nil {
		if errors.Is(err, store.ErrVersion) {
			logger.Warn("Config", "Ignoring prefix config with an unknown version: %v", err)
		} else {
			logger.Error("Config", "Failed to parse %s: %v", s.ConfigPath(prefixName), err)
		}
		return DefaultPrefixOptions(prefixName)
	}
	if !found {
		return DefaultPrefixOptions(prefixName)
	}
	return cfg
}

func (s *PrefixStore) Save(prefixName string, options types.LaunchOptions) error {
	file := store.NewFile(s.ConfigPath(prefixName), DataVersion)
	return file.Save(options)
}
