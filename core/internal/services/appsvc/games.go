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
// Each config is one launcher here; the grouped view of scan folders is getAutoScannedGames.
func (s *service) getAllGames() []types.GameInfo {
	configs := s.games.List()
	scanFolders := s.settings.Load().ScanFolders

	games := []types.GameInfo{}
	for _, gameConfig := range configs {
		cleaned := config.CleanPath(config.FirstNonEmpty(gameConfig.LauncherPath, gameConfig.GamePath))
		if cleaned == "" {
			continue
		}
		games = append(games, singleLauncherGame(gameConfig, cleaned, isInAnyFolder(scanFolders, cleaned)))
	}
	return games
}

// singleLauncherGame builds a card for a game that has one launcher, such as a game added by hand.
func singleLauncherGame(gameConfig types.LaunchOptions, cleaned string, autoScanned bool) types.GameInfo {
	launcher := types.Launcher{
		Name:    config.NameWithoutExt(cleaned),
		Path:    cleaned,
		Kind:    types.LauncherMain,
		Primary: true,
	}
	return types.GameInfo{
		Name:          gameConfigName(gameConfig),
		Path:          cleaned,
		Icon:          "",
		Config:        gameConfig,
		IsRecent:      false,
		IsAutoScanned: autoScanned,
		Folder:        cleaned,
		Launchers:     []types.Launcher{launcher},
		LauncherCount: 1,
	}
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

// searchExecutables lists the .exe paths under folderPath, skipping names that match excludeNames.
func searchExecutables(folderPath string, maxDepth int, excludeNames []string) []string {
	files := scanExecutables(folderPath, maxDepth, excludeNames)
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	return paths
}

// scanExecutables lists the .exe files under folderPath with their sizes.
func scanExecutables(folderPath string, maxDepth int, excludeNames []string) []scannedFile {
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
