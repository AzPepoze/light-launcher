package config

import (
	"errors"
	"os"
	"path/filepath"

	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/store"
	"light-launcher/core/internal/types"
)

// GameStore owns the per-game config.json files under one executables directory.
type GameStore struct {
	directory string
}

// NewGameStore opens the executables directory.
func NewGameStore(directory string) *GameStore {
	return &GameStore{directory: directory}
}

// Path is the config folder for one saved game. The id wins when set, as in the Electron app.
func (s *GameStore) Path(name, id string) string {
	key := name
	if id != "" {
		key = id
	}
	return filepath.Join(s.directory, key)
}

// ConfigPath is the config.json for one saved game.
func (s *GameStore) ConfigPath(name, id string) string {
	return filepath.Join(s.Path(name, id), "config.json")
}

// List reads every game config.json and skips folders that are missing or broken.
func (s *GameStore) List() []types.LaunchOptions {
	if _, err := os.Stat(s.directory); err != nil {
		return []types.LaunchOptions{}
	}
	entries, err := os.ReadDir(s.directory)
	if err != nil {
		logger.Error("Config", "Error listing game configs: %v", err)
		return []types.LaunchOptions{}
	}

	configs := []types.LaunchOptions{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		cfg, err := s.read(entry.Name())
		if err != nil {
			if errors.Is(err, store.ErrVersion) {
				logger.Warn("Config", "Ignoring game config with an unknown version: %v", err)
			} else if !os.IsNotExist(err) {
				logger.Error("Config", "Failed to parse %s: %v", filepath.Join(s.directory, entry.Name(), "config.json"), err)
			}
			continue
		}
		configs = append(configs, cfg)
	}
	return configs
}

// Find returns the config whose game or launcher path matches executablePath, or nil.
func (s *GameStore) Find(executablePath string) *types.LaunchOptions {
	target := PathKey(executablePath)
	if target == "" {
		return nil
	}
	for _, cfg := range s.List() {
		if PathKey(cfg.GamePath) == target || PathKey(cfg.LauncherPath) == target {
			found := cfg
			return &found
		}
	}
	return nil
}

// Save stores the config under its id, generating an id for new configs.
func (s *GameStore) Save(options *types.LaunchOptions) error {
	if options.ID == "" {
		options.ID = NewID()
	}
	file := store.NewFile(s.ConfigPath(options.Name, options.ID), DataVersion)
	return file.Save(options)
}

func (s *GameStore) read(directoryName string) (types.LaunchOptions, error) {
	var cfg types.LaunchOptions
	file := store.NewFile(filepath.Join(s.directory, directoryName, "config.json"), DataVersion)
	found, err := file.Load(&cfg)
	if err != nil {
		return types.LaunchOptions{}, err
	}
	if !found {
		return types.LaunchOptions{}, os.ErrNotExist
	}
	return cfg, nil
}
