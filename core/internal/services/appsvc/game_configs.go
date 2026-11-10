package appsvc

import (
	"os"
	"path/filepath"
	"strings"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/types"
)

// listGameConfigs reads every game config.json and skips folders that are missing or broken.
func listGameConfigs() []types.LaunchOptions {
	configDir := config.GetConfigDirectory()
	if _, err := os.Stat(configDir); err != nil {
		return []types.LaunchOptions{}
	}
	entries, err := os.ReadDir(configDir)
	if err != nil {
		logger.Error("Config", "Error listing game configs: %v", err)
		return []types.LaunchOptions{}
	}
	configs := []types.LaunchOptions{}
	for _, entry := range entries {
		if cfg, ok := readGameConfigEntry(configDir, entry); ok {
			configs = append(configs, cfg)
		}
	}
	return configs
}

func readGameConfigEntry(configDir string, entry os.DirEntry) (types.LaunchOptions, bool) {
	if !entry.IsDir() {
		return types.LaunchOptions{}, false
	}
	configPath := filepath.Join(configDir, entry.Name(), "config.json")
	if _, err := os.Stat(configPath); err != nil {
		return types.LaunchOptions{}, false
	}
	cfg, err := loadJSON[types.LaunchOptions](configPath)
	if err != nil {
		logger.Error("Config", "Failed to parse %s: %v", configPath, err)
		return types.LaunchOptions{}, false
	}
	return cfg, true
}

// saveGameConfig stores the config under its id, generating an id for new configs.
func saveGameConfig(options *types.LaunchOptions) error {
	if options.ID == "" {
		options.ID = generateID()
	}
	return saveJSON(config.GetGameConfigFilePath(options.Name, options.ID), *options)
}

// loadGameConfig finds the config whose game or launcher path matches executablePath (case-insensitive).
func loadGameConfig(executablePath string) *types.LaunchOptions {
	target := strings.ToLower(normalizePath(executablePath))
	for _, cfg := range listGameConfigs() {
		if matchesNormalizedPath(cfg.GamePath, target) || matchesNormalizedPath(cfg.LauncherPath, target) {
			return &cfg
		}
	}
	return nil
}

func matchesNormalizedPath(p, target string) bool {
	return p != "" && strings.ToLower(normalizePath(p)) == target
}
