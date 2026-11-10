package appsvc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"light-launcher/core/internal/types"
)

// addScanFolder watches a folder with the default depth and excludes.
func addScanFolder(folderPath string) error {
	resetScanCache()
	return updateAppSettings(func(settings *types.AppSettings) error {
		cleaned := cleanPath(folderPath)
		if !slices.Contains(settings.ScanFolders, cleaned) {
			settings.ScanFolders = append(settings.ScanFolders, cleaned)
		}
		if findScanFolderIndex(settings.ScanFolderConfigs, cleaned) < 0 {
			settings.ScanFolderConfigs = append(settings.ScanFolderConfigs, types.ScanFolderConfig{
				Path:         cleaned,
				Depth:        2,
				ExcludeNames: defaultScanExcludeNames(),
			})
		}
		return nil
	})
}

// renameScanFolder moves a watched folder to a new path and returns the cleaned new path.
func renameScanFolder(oldPath, newPath string, depth int, excludeNames []string) (string, error) {
	resetScanCache()
	cleanedOld := cleanPath(oldPath)
	cleanedNew := cleanPath(newPath)
	if cleanedNew == "" {
		return "", errors.New("Folder path cannot be empty")
	}
	if _, err := os.Stat(cleanedNew); err != nil {
		return "", fmt.Errorf("Folder not found: %s", cleanedNew)
	}

	err := updateAppSettings(func(settings *types.AppSettings) error {
		return renameInSettings(settings, cleanedOld, cleanedNew, depth, excludeNames)
	})
	if err != nil {
		return "", err
	}
	return cleanedNew, nil
}

func renameInSettings(settings *types.AppSettings, oldPath, newPath string, depth int, excludeNames []string) error {
	index := findScanFolderIndex(settings.ScanFolderConfigs, oldPath)
	if index < 0 {
		return fmt.Errorf("Watched folder not found: %s", oldPath)
	}
	if strings.ToLower(newPath) != strings.ToLower(oldPath) && hasCaseInsensitiveScanFolder(settings.ScanFolderConfigs, newPath) {
		return errors.New("That folder is already watched")
	}

	cfg := &settings.ScanFolderConfigs[index]
	cfg.Path = newPath
	cfg.Depth = depth
	cfg.ExcludeNames = excludeNames
	settings.ScanFolders = renameScanFolderEntries(settings.ScanFolders, oldPath, newPath)
	if !containsCleanPath(settings.ScanFolders, newPath) {
		settings.ScanFolders = append(settings.ScanFolders, newPath)
	}
	settings.Blacklist = remapBlacklist(settings.Blacklist, oldPath, newPath)
	return nil
}

func renameScanFolderEntries(folders []string, oldPath, newPath string) []string {
	renamed := make([]string, 0, len(folders))
	for _, folder := range folders {
		if cleanPath(folder) == oldPath {
			renamed = append(renamed, newPath)
		} else {
			renamed = append(renamed, folder)
		}
	}
	return renamed
}

// remapBlacklist moves blacklist entries that sit under the old folder to the new folder.
func remapBlacklist(blacklist []string, oldPath, newPath string) []string {
	remapped := make([]string, 0, len(blacklist))
	for _, entry := range blacklist {
		remapped = append(remapped, remapBlacklistEntry(entry, oldPath, newPath))
	}
	return remapped
}

func remapBlacklistEntry(entry, oldPath, newPath string) string {
	cleaned := cleanPath(entry)
	if strings.ToLower(cleaned) == strings.ToLower(oldPath) {
		return newPath
	}
	rel := relativePath(oldPath, cleaned)
	if rel != "" && !strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel) {
		return filepath.Join(newPath, rel)
	}
	return entry
}

// removeScanFolder stops watching a folder.
func removeScanFolder(folderPath string) error {
	resetScanCache()
	cleaned := cleanPath(folderPath)
	return updateAppSettings(func(settings *types.AppSettings) error {
		settings.ScanFolders = slices.DeleteFunc(settings.ScanFolders, func(folder string) bool {
			return cleanPath(folder) == cleaned
		})
		settings.ScanFolderConfigs = slices.DeleteFunc(settings.ScanFolderConfigs, func(cfg types.ScanFolderConfig) bool {
			return cleanPath(cfg.Path) == cleaned
		})
		return nil
	})
}

// updateScanFolderConfig changes depth and excludes. An unknown folder is added to the watch list.
func updateScanFolderConfig(folderPath string, depth int, excludeNames []string) error {
	resetScanCache()
	cleaned := cleanPath(folderPath)
	return updateAppSettings(func(settings *types.AppSettings) error {
		index := findScanFolderIndex(settings.ScanFolderConfigs, cleaned)
		if index >= 0 {
			settings.ScanFolderConfigs[index].Depth = depth
			settings.ScanFolderConfigs[index].ExcludeNames = excludeNames
			return nil
		}
		settings.ScanFolderConfigs = append(settings.ScanFolderConfigs, types.ScanFolderConfig{
			Path:         cleaned,
			Depth:        depth,
			ExcludeNames: excludeNames,
		})
		if !containsCleanPath(settings.ScanFolders, cleaned) {
			settings.ScanFolders = append(settings.ScanFolders, cleaned)
		}
		return nil
	})
}

// getScanFolderConfig returns the saved config for a folder, or the defaults when it is not watched.
func getScanFolderConfig(folderPath string) types.ScanFolderConfig {
	cleaned := cleanPath(folderPath)
	settings := loadAppSettings()
	if index := findScanFolderIndex(settings.ScanFolderConfigs, cleaned); index >= 0 {
		return settings.ScanFolderConfigs[index]
	}
	return types.ScanFolderConfig{
		Path:         cleaned,
		Depth:        2,
		ExcludeNames: defaultScanExcludeNames(),
	}
}

// blacklistGame hides an executable from auto scans.
func blacklistGame(executablePath string) error {
	resetScanCache()
	cleaned := cleanPath(executablePath)
	return updateAppSettings(func(settings *types.AppSettings) error {
		if !containsCleanPath(settings.Blacklist, cleaned) {
			settings.Blacklist = append(settings.Blacklist, cleaned)
		}
		return nil
	})
}

// unblacklistGame shows an executable in auto scans again.
func unblacklistGame(executablePath string) error {
	resetScanCache()
	cleaned := cleanPath(executablePath)
	return updateAppSettings(func(settings *types.AppSettings) error {
		settings.Blacklist = slices.DeleteFunc(settings.Blacklist, func(entry string) bool {
			return cleanPath(entry) == cleaned
		})
		return nil
	})
}

func findScanFolderIndex(configs []types.ScanFolderConfig, cleaned string) int {
	for i, cfg := range configs {
		if cleanPath(cfg.Path) == cleaned {
			return i
		}
	}
	return -1
}

func hasCaseInsensitiveScanFolder(configs []types.ScanFolderConfig, cleaned string) bool {
	for _, cfg := range configs {
		if strings.ToLower(cleanPath(cfg.Path)) == strings.ToLower(cleaned) {
			return true
		}
	}
	return false
}

func containsCleanPath(list []string, cleaned string) bool {
	for _, entry := range list {
		if cleanPath(entry) == cleaned {
			return true
		}
	}
	return false
}
