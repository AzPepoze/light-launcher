package app

import (
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

	// 1. Get manually registered games to filter them out
	manualGames, err := app.GetAllGames()
	if err != nil {
		manualGames = nil
	}
	manualPaths := make(map[string]bool)
	for _, mg := range manualGames {
		norm := normPath(mg.Path)
		manualPaths[norm] = true
	}

	// 2. Prepare blacklist lookup
	blacklist := make(map[string]bool)
	for _, p := range settings.Blacklist {
		blacklist[normPath(p)] = true
	}

	groups := make([]types.ScannedFolderGroup, 0)
	defaultExcludes := []string{"UnityCrashHandler64", "uninstall", "redist", "vc_redist", "dxsetup"}

	for _, folder := range settings.ScanFolders {
		cleanedFolder := filepath.Clean(folder)
		folderName := filepath.Base(cleanedFolder)

		// Scan up to depth 2
		executables, err := app.SearchExecutables(cleanedFolder, 2, defaultExcludes)
		games := make([]types.GameInfo, 0)

		if err == nil {
			for _, exePath := range executables {
				normExe := normPath(exePath)
				if blacklist[normExe] || manualPaths[normExe] {
					continue
				}

				name := filepath.Base(exePath)
				name = strings.TrimSuffix(name, filepath.Ext(name))

				// Create a default virtual launch option
				cfg := types.LaunchOptions{
					Name:         name,
					LauncherPath: exePath,
					GamePath:     exePath,
					PrefixPath:   "", // empty initially
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

				// Add virtual GameInfo
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
	// Check if already exists
	for _, f := range settings.ScanFolders {
		if filepath.Clean(f) == cleaned {
			return nil
		}
	}

	settings.ScanFolders = append(settings.ScanFolders, cleaned)
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
	return config.SaveAppSettings(*settings)
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
