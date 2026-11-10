package appsvc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

// getAllGames lists the saved game configs. A game is flagged auto-scanned when it lives under a scan folder.
func getAllGames() []types.GameInfo {
	configs := listGameConfigs()
	scanFolders := loadAppSettings().ScanFolders

	games := []types.GameInfo{}
	for _, gameConfig := range configs {
		cleaned := cleanPath(firstNonEmpty(gameConfig.LauncherPath, gameConfig.GamePath))
		if cleaned == "" {
			continue
		}
		games = append(games, types.GameInfo{
			Name:          gameConfigName(gameConfig),
			Path:          cleaned,
			Icon:          "",
			Config:        gameConfig,
			IsRecent:      false,
			IsAutoScanned: isInAnyFolder(scanFolders, cleaned),
		})
	}
	return games
}

func gameConfigName(cfg types.LaunchOptions) string {
	if cfg.Name != "" {
		return cfg.Name
	}
	return fileNameWithoutExt(firstNonEmpty(cfg.GamePath, cfg.LauncherPath))
}

func isInAnyFolder(folders []string, target string) bool {
	for _, folder := range folders {
		if isSubPath(folder, target) {
			return true
		}
	}
	return false
}

// removeGame deletes the config folder of the game at executablePath.
func removeGame(executablePath string) error {
	resetScanCache()
	if executablePath == "" {
		return errors.New("Executable path cannot be empty")
	}
	if cfg := loadGameConfig(executablePath); cfg != nil {
		return removeConfigFolder(cfg.Name, cfg.ID)
	}

	target := strings.ToLower(cleanPath(executablePath))
	for _, cfg := range listGameConfigs() {
		if matchesCleanPath(cfg.GamePath, target) || matchesCleanPath(cfg.LauncherPath, target) {
			return removeConfigFolder(cfg.Name, cfg.ID)
		}
	}
	return fmt.Errorf("Could not find game config to remove for path: %s", executablePath)
}

func matchesCleanPath(p, target string) bool {
	return p != "" && strings.ToLower(cleanPath(p)) == target
}

func removeConfigFolder(name, id string) error {
	folder := config.GetExecutableConfigPath(name, id)
	if _, err := os.Stat(folder); err != nil {
		return nil
	}
	return os.RemoveAll(folder)
}

// searchExecutables lists .exe files under folderPath, skipping names that match excludeNames.
func searchExecutables(folderPath string, maxDepth int, excludeNames []string) []string {
	return scanDirectoryFiles(folderPath, scanOptions{
		MaxDepth:        maxDepth,
		ExcludePatterns: excludeNames,
		Extensions:      []string{".exe"},
	})
}

// newDefaultGameConfig builds the config for an executable found by a folder scan.
func newDefaultGameConfig(name, executablePath string) types.LaunchOptions {
	return types.LaunchOptions{
		ID:           generateID(),
		Name:         name,
		LauncherPath: executablePath,
		GamePath:     executablePath,
		PrefixPath:   filepath.Join(config.GetPrefixBaseDirectory(), "Default"),
		Extras:       defaultExtras(),
	}
}
