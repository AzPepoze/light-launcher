package app

import (
	"fmt"
	"path/filepath"
	"strings"

	"light-launcher/internal/config"
	"light-launcher/internal/types"
)

func (app *App) GetAutoScannedGames() ([]types.ScannedFolderGroup, error) {
	settings := config.LoadAppSettings()
	if settings == nil {
		return nil, nil
	}

	// 1. Get manually registered games to filter them out (only those that are NOT auto-scanned)
	manualGames, err := app.GetAllGames()
	if err != nil {
		manualGames = nil
	}
	manualPaths := make(map[string]bool)
	for _, mg := range manualGames {
		if !mg.IsAutoScanned {
			norm := normPath(mg.Path)
			manualPaths[norm] = true
		}
	}

	// Load all existing configs once for fast lookup
	configs, err := config.ListGameConfigs()
	existingConfigs := make(map[string]types.LaunchOptions)
	if err == nil {
		for _, cfg := range configs {
			normGame := normPath(cfg.GamePath)
			normLaunch := normPath(cfg.LauncherPath)
			if normGame != "" {
				existingConfigs[normGame] = cfg
			}
			if normLaunch != "" {
				existingConfigs[normLaunch] = cfg
			}
		}
	}

	// 2. Prepare blacklist lookup
	blacklist := make(map[string]bool)
	for _, p := range settings.Blacklist {
		blacklist[normPath(p)] = true
	}

	groups := make([]types.ScannedFolderGroup, 0)

	for _, folderCfg := range settings.ScanFolderConfigs {
		cleanedFolder := filepath.Clean(folderCfg.Path)
		folderName := filepath.Base(cleanedFolder)

		// Scan using folderCfg.Depth and folderCfg.ExcludeNames
		executables, err := app.SearchExecutables(cleanedFolder, folderCfg.Depth, folderCfg.ExcludeNames)
		games := make([]types.GameInfo, 0)

		if err == nil {
			for _, exePath := range executables {
				normExe := normPath(exePath)
				if blacklist[normExe] || manualPaths[normExe] {
					continue
				}

				name := filepath.Base(exePath)
				name = strings.TrimSuffix(name, filepath.Ext(name))

				var cfg types.LaunchOptions
				if existing, exists := existingConfigs[normExe]; exists {
					cfg = existing
				} else {
					defaultPrefix := filepath.Join(config.GetPrefixBaseDirectory(), "Default")
					cfg = types.LaunchOptions{
						ID:           config.GenerateID(),
						Name:         name,
						LauncherPath: exePath,
						GamePath:     exePath,
						PrefixPath:   defaultPrefix,
						Extras: types.ExtrasConfig{
							Lsfg: types.LsfgConfig{
								Multiplier: "2",
							},
							Memory: types.MemoryConfig{
								Value: "4G",
							},
							Gamescope: types.GamescopeConfig{
								Width:       "1920",
								Height:      "1080",
								RefreshRate: "60",
							},
						},
					}
					_ = config.SaveGameConfig(cfg)
				}

				// Add GameInfo
				games = append(games, types.GameInfo{
					Name:          name,
					Path:          exePath,
					Config:        cfg,
					IsAutoScanned: true,
				})
			}
		}

		groups = append(groups, types.ScannedFolderGroup{
			FolderPath: cleanedFolder,
			FolderName: folderName,
			Games:      games,
		})
	}

	return groups, nil
}

func (app *App) AddScanFolder(folderPath string) error {
	settings := config.LoadAppSettings()
	if settings == nil {
		settings = &types.AppSettings{}
	}

	cleaned := filepath.Clean(folderPath)
	// Check if already exists in ScanFolders
	existsLegacy := false
	for _, f := range settings.ScanFolders {
		if filepath.Clean(f) == cleaned {
			existsLegacy = true
			break
		}
	}
	if !existsLegacy {
		settings.ScanFolders = append(settings.ScanFolders, cleaned)
	}

	// Check if already exists in ScanFolderConfigs
	existsConfig := false
	for _, cfg := range settings.ScanFolderConfigs {
		if filepath.Clean(cfg.Path) == cleaned {
			existsConfig = true
			break
		}
	}
	if !existsConfig {
		settings.ScanFolderConfigs = append(settings.ScanFolderConfigs, types.ScanFolderConfig{
			Path:         cleaned,
			Depth:        2,
			ExcludeNames: []string{"UnityCrashHandler64", "uninstall", "redist", "vc_redist", "dxsetup"},
		})
	}

	return config.SaveAppSettings(*settings)
}

func (app *App) RemoveScanFolder(folderPath string) error {
	settings := config.LoadAppSettings()
	if settings == nil {
		return nil
	}

	cleaned := filepath.Clean(folderPath)
	newFolders := make([]string, 0)
	for _, f := range settings.ScanFolders {
		if filepath.Clean(f) != cleaned {
			newFolders = append(newFolders, f)
		}
	}
	settings.ScanFolders = newFolders

	newConfigs := make([]types.ScanFolderConfig, 0)
	for _, cfg := range settings.ScanFolderConfigs {
		if filepath.Clean(cfg.Path) != cleaned {
			newConfigs = append(newConfigs, cfg)
		}
	}
	settings.ScanFolderConfigs = newConfigs

	return config.SaveAppSettings(*settings)
}

func (app *App) UpdateScanFolderConfig(folderPath string, depth int, excludeNames []string) error {
	settings := config.LoadAppSettings()
	if settings == nil {
		return fmt.Errorf("settings not found")
	}

	cleaned := filepath.Clean(folderPath)
	found := false
	for i, cfg := range settings.ScanFolderConfigs {
		if filepath.Clean(cfg.Path) == cleaned {
			settings.ScanFolderConfigs[i].Depth = depth
			settings.ScanFolderConfigs[i].ExcludeNames = excludeNames
			found = true
			break
		}
	}

	if !found {
		settings.ScanFolderConfigs = append(settings.ScanFolderConfigs, types.ScanFolderConfig{
			Path:         cleaned,
			Depth:        depth,
			ExcludeNames: excludeNames,
		})
		
		exists := false
		for _, sf := range settings.ScanFolders {
			if filepath.Clean(sf) == cleaned {
				exists = true
				break
			}
		}
		if !exists {
			settings.ScanFolders = append(settings.ScanFolders, cleaned)
		}
	}

	return config.SaveAppSettings(*settings)
}

func (app *App) GetScanFolderConfig(folderPath string) (*types.ScanFolderConfig, error) {
	settings := config.LoadAppSettings()
	if settings == nil {
		return nil, fmt.Errorf("settings not found")
	}

	cleaned := filepath.Clean(folderPath)
	for _, cfg := range settings.ScanFolderConfigs {
		if filepath.Clean(cfg.Path) == cleaned {
			return &cfg, nil
		}
	}

	for _, sf := range settings.ScanFolders {
		if filepath.Clean(sf) == cleaned {
			return &types.ScanFolderConfig{
				Path:         cleaned,
				Depth:        2,
				ExcludeNames: []string{"UnityCrashHandler64", "uninstall", "redist", "vc_redist", "dxsetup"},
			}, nil
		}
	}

	return nil, fmt.Errorf("folder not found: %s", folderPath)
}

func (app *App) BlacklistGame(executablePath string) error {
	settings := config.LoadAppSettings()
	if settings == nil {
		settings = &types.AppSettings{}
	}

	cleaned := filepath.Clean(executablePath)
	// Check if already blacklisted
	for _, p := range settings.Blacklist {
		if filepath.Clean(p) == cleaned {
			return nil
		}
	}

	settings.Blacklist = append(settings.Blacklist, cleaned)
	return config.SaveAppSettings(*settings)
}

func (app *App) UnblacklistGame(executablePath string) error {
	settings := config.LoadAppSettings()
	if settings == nil {
		return nil
	}

	cleaned := filepath.Clean(executablePath)
	newBlacklist := make([]string, 0)
	for _, p := range settings.Blacklist {
		if filepath.Clean(p) != cleaned {
			newBlacklist = append(newBlacklist, p)
		}
	}

	settings.Blacklist = newBlacklist
	return config.SaveAppSettings(*settings)
}

func normPath(p string) string {
	cleaned := filepath.Clean(p)
	if abs, err := filepath.Abs(cleaned); err == nil {
		cleaned = abs
	}
	return strings.ToLower(cleaned)
}
