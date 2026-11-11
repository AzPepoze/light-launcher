package appsvc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

// getAllGames lists the saved game configs. A game is flagged auto-scanned when it lives under a scan folder.
func (s *service) getAllGames() []types.GameInfo {
	configs := s.games.List()
	scanFolders := s.settings.Load().ScanFolders

	games := []types.GameInfo{}
	for _, gameConfig := range configs {
		cleaned := config.CleanPath(config.FirstNonEmpty(gameConfig.LauncherPath, gameConfig.GamePath))
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
	return config.NameWithoutExt(config.FirstNonEmpty(cfg.GamePath, cfg.LauncherPath))
}

func isInAnyFolder(folders []string, target string) bool {
	for _, folder := range folders {
		if config.IsSubPath(folder, target) {
			return true
		}
	}
	return false
}

// removeGame deletes the config folder of the game at executablePath.
func (s *service) removeGame(executablePath string) error {
	resetScanCache()
	if executablePath == "" {
		return errors.New("Executable path cannot be empty")
	}
	cfg := s.games.Find(executablePath)
	if cfg == nil {
		return fmt.Errorf("Could not find game config to remove for path: %s", executablePath)
	}
	return s.removeConfigFolder(cfg.Name, cfg.ID)
}

func (s *service) removeConfigFolder(name, id string) error {
	folder := s.games.Path(name, id)
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
		ID:           config.NewID(),
		Name:         name,
		LauncherPath: executablePath,
		GamePath:     executablePath,
		PrefixPath:   filepath.Join(config.GetPrefixBaseDirectory(), "Default"),
		Extras:       config.DefaultExtras(),
	}
}
