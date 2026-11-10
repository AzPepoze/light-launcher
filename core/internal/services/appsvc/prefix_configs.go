package appsvc

import (
	"os"
	"path/filepath"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/types"
)

// loadPrefixConfig returns the saved prefix config, or defaults when none is saved or it is broken.
func loadPrefixConfig(prefixName string) types.LaunchOptions {
	settings := loadAppSettings()
	configPath := prefixConfigFilePath(prefixName, settings.CustomPrefixDir)
	if _, err := os.Stat(configPath); err == nil {
		cfg, err := loadJSON[types.LaunchOptions](configPath)
		if err == nil {
			return cfg
		}
		logger.Error("Config", "Failed to parse %s: %v", configPath, err)
	}
	return defaultPrefixOptions(prefixName)
}

func savePrefixConfig(prefixName string, options types.LaunchOptions) error {
	settings := loadAppSettings()
	return saveJSON(prefixConfigFilePath(prefixName, settings.CustomPrefixDir), options)
}

// defaultPrefixOptions uses the default prefixes folder, even when a custom folder is set.
func defaultPrefixOptions(prefixName string) types.LaunchOptions {
	return types.LaunchOptions{
		Name:       prefixName,
		PrefixPath: filepath.Join(config.GetPrefixBaseDirectory(), prefixName),
		Extras:     defaultExtras(),
	}
}
