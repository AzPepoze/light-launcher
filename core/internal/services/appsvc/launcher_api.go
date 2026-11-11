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

// setPreferredLauncher remembers executablePath as the launcher of a game folder and returns the refreshed card.
func (s *service) setPreferredLauncher(folder, executablePath string) (types.GameInfo, error) {
	cleanFolder := config.CleanPath(folder)
	cleanPath := config.CleanPath(executablePath)
	if err := s.requireWatchedGameFolder(cleanFolder); err != nil {
		return types.GameInfo{}, err
	}
	if err := checkLauncherPath(cleanFolder, cleanPath); err != nil {
		return types.GameInfo{}, err
	}
	if err := s.preferences.Set(cleanFolder, cleanPath); err != nil {
		return types.GameInfo{}, err
	}
	resetScanCache()
	return s.scannedGameAt(cleanFolder)
}

// clearPreferredLauncher forgets the chosen launcher of a game folder and returns the refreshed card.
func (s *service) clearPreferredLauncher(folder string) (types.GameInfo, error) {
	cleanFolder := config.CleanPath(folder)
	if err := s.requireWatchedGameFolder(cleanFolder); err != nil {
		return types.GameInfo{}, err
	}
	if err := s.preferences.Clear(cleanFolder); err != nil {
		return types.GameInfo{}, err
	}
	resetScanCache()
	return s.scannedGameAt(cleanFolder)
}

// requireWatchedGameFolder rejects folders that are not inside a watched folder.
func (s *service) requireWatchedGameFolder(folder string) error {
	if folder == "" {
		return errors.New("Folder path cannot be empty")
	}
	if !isInAnyFolder(s.settings.Load().ScanFolders, folder) {
		return fmt.Errorf("Folder is not a watched game folder: %s", folder)
	}
	return nil
}

// checkLauncherPath accepts an .exe file inside the game folder that is not a helper program.
func checkLauncherPath(folder, executablePath string) error {
	if !config.IsSubPath(folder, executablePath) {
		return fmt.Errorf("Launcher is not inside the game folder: %s", executablePath)
	}
	if !strings.EqualFold(config.ExtName(executablePath), ".exe") {
		return fmt.Errorf("Launcher must be an .exe file: %s", executablePath)
	}
	info, err := os.Stat(executablePath)
	if err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("Launcher file not found: %s", executablePath)
	}
	if isJunkExecutable(filepath.Base(executablePath)) {
		return fmt.Errorf("Helper programs cannot be the launcher: %s", executablePath)
	}
	return nil
}

// scannedGameAt returns the card of one game folder after a fresh scan.
func (s *service) scannedGameAt(folder string) (types.GameInfo, error) {
	groups, err := s.getAutoScannedGames(false)
	if err != nil {
		return types.GameInfo{}, err
	}
	for _, group := range groups {
		for _, game := range group.Games {
			if config.PathKey(game.Folder) == config.PathKey(folder) {
				return game, nil
			}
		}
	}
	return types.GameInfo{}, fmt.Errorf("No launchers found in game folder: %s", folder)
}
