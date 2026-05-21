package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"light-launcher/internal/config"
	"light-launcher/internal/types"
	"light-launcher/internal/utils/lsfg"
)

func (app *App) GetAllGames() ([]types.GameInfo, error) {
	configs, err := config.ListGameConfigs()
	if err != nil {
		return nil, err
	}

	settings := config.LoadAppSettings()
	var scanFolders []string
	if settings != nil {
		scanFolders = settings.ScanFolders
	}

	games := make([]types.GameInfo, 0)
	for _, gameConfig := range configs {
		name := gameConfig.Name
		if name == "" {
			name = filepath.Base(gameConfig.GamePath)
			name = strings.TrimSuffix(name, filepath.Ext(name))
		}

		cleanedPath := filepath.Clean(gameConfig.LauncherPath)
		if cleanedPath == "" {
			cleanedPath = filepath.Clean(gameConfig.GamePath)
		}
		if absolutePath, err := filepath.Abs(cleanedPath); err == nil {
			cleanedPath = absolutePath
		}

		inScanFolder := false
		for _, sf := range scanFolders {
			sfCleaned := filepath.Clean(sf)
			if abs, err := filepath.Abs(sfCleaned); err == nil {
				sfCleaned = abs
			}
			if isSubpath(cleanedPath, sfCleaned) {
				inScanFolder = true
				break
			}
		}

		games = append(games, types.GameInfo{
			Name:          name,
			Path:          cleanedPath,
			Config:        gameConfig,
			IsAutoScanned: inScanFolder,
		})
	}
	return games, nil
}

func isSubpath(path, base string) bool {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..") && rel != "."
}

func (app *App) RemoveGame(executablePath string) error {
	cfg, err := app.GetConfig(executablePath)
	if err != nil {
		return fmt.Errorf("could not find game to remove: %w", err)
	}

	configDirectory := config.GetExecutableConfigPath(cfg.Name, cfg.ID)

	if _, err := os.Stat(configDirectory); err == nil {
		if err := os.RemoveAll(configDirectory); err != nil {
			return fmt.Errorf("failed to remove game config: %w", err)
		}
	}

	_ = lsfg.DisableProfileInConfig(cfg.Name, executablePath)
	return nil
}

func (app *App) GetConfig(executablePath string) (*types.LaunchOptions, error) {
	return config.LoadGameConfig(executablePath)
}

func (app *App) SaveGameConfig(options types.LaunchOptions) error {
	return config.SaveGameConfig(options)
}
