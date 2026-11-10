package runtimesvc

import (
	"path/filepath"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/types"
)

func savePrefixConfig(prefixName string, options types.LaunchOptions) error {
	return saveJSON(prefixConfigPath(prefixName, loadAppSettings().CustomPrefixDir), options)
}

// loadPrefixConfig returns the saved prefix config, or defaults when it is missing or unreadable.
func loadPrefixConfig(prefixName string) types.LaunchOptions {
	configPath := prefixConfigPath(prefixName, loadAppSettings().CustomPrefixDir)
	if pathExists(configPath) {
		cfg, err := loadJSON[types.LaunchOptions](configPath)
		if err == nil {
			return cfg
		}
		logger.Error("Config", "Failed to parse %s: %v", configPath, err)
	}
	return defaultPrefixConfig(prefixName)
}

// loadPrefixConfigWithProton adds the matching installed Proton tool to the prefix config.
func loadPrefixConfigWithProton(prefixName string) types.PrefixConfigWithProton {
	cfg := loadPrefixConfig(prefixName)
	result := types.PrefixConfigWithProton{
		Config:     cfg,
		ProtonPath: cfg.ProtonPath,
	}
	if cfg.ProtonPath == "" {
		return result
	}

	match := findProtonMatch(cfg.ProtonPath, scanProtonVersions())
	if match == nil {
		result.ProtonDisplayName = cfg.ProtonPath
		return result
	}
	result.ProtonDisplayName = match.DisplayName
	result.ProtonName = match.Name
	result.ProtonPath = match.Path
	result.ProtonIsSteam = match.IsSteam
	return result
}

// defaultPrefixConfig matches the Electron app. Its prefix path ignores the custom prefix folder setting.
func defaultPrefixConfig(prefixName string) types.LaunchOptions {
	return types.LaunchOptions{
		Name:       prefixName,
		PrefixPath: filepath.Join(config.GetPrefixBaseDirectory(), prefixName),
		Extras: types.ExtrasConfig{
			Lsfg: types.LsfgConfig{
				Multiplier: config.DefaultMultiplier,
				FlowScale:  "1.0",
				Pacing:     "smooth",
			},
			Gamescope: types.GamescopeConfig{
				Width:       config.DefaultWidth,
				Height:      config.DefaultHeight,
				RefreshRate: config.DefaultRefreshRate,
				WindowMode:  "borderless",
				Scaler:      "auto",
				Filter:      "linear",
				Sharpness:   "0",
			},
			Memory: types.MemoryConfig{
				Value: config.DefaultMemoryValue,
			},
		},
	}
}
