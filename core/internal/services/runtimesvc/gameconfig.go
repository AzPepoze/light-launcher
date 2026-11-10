package runtimesvc

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/types"
)

// listGameConfigs reads config.json from every game folder. Broken files are logged and skipped.
func listGameConfigs() []types.LaunchOptions {
	configDirectory := config.GetConfigDirectory()
	entries, err := os.ReadDir(configDirectory)
	if err != nil {
		return []types.LaunchOptions{}
	}

	configs := []types.LaunchOptions{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		configPath := filepath.Join(configDirectory, entry.Name(), "config.json")
		if !pathExists(configPath) {
			continue
		}
		cfg, err := loadJSON[types.LaunchOptions](configPath)
		if err != nil {
			logger.Error("Config", "Failed to parse %s: %v", configPath, err)
			continue
		}
		configs = append(configs, cfg)
	}
	return configs
}

// loadGameConfig finds the saved config whose game or launcher path matches executablePath.
// It returns nil when no config matches.
func loadGameConfig(executablePath string) *types.LaunchOptions {
	cleanPath := strings.ToLower(filepath.Clean(executablePath))
	for _, cfg := range listGameConfigs() {
		if matchKey(cfg.GamePath) == cleanPath || matchKey(cfg.LauncherPath) == cleanPath {
			found := cfg
			return &found
		}
	}
	return nil
}

// saveGameConfig stores the config and gives it a new ID when it has none.
func saveGameConfig(options *types.LaunchOptions) error {
	if options.ID == "" {
		options.ID = newConfigID()
	}
	configPath := config.GetGameConfigFilePath(options.Name, options.ID)
	return saveJSON(configPath, options)
}

// newConfigID returns 8 random bytes as 16 hex characters.
func newConfigID() string {
	buffer := make([]byte, 8)
	_, _ = rand.Read(buffer)
	return hex.EncodeToString(buffer)
}

// lsfgIdentity returns the config ID used to find a game's LSFG profile folder.
func lsfgIdentity(gamePath string) string {
	cfg := loadGameConfig(gamePath)
	if cfg == nil {
		return ""
	}
	return cfg.ID
}
