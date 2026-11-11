package appsvc

import (
	"os"
	"path/filepath"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/logger"
)

// legacyDirectory is the ~/LightLauncher folder used before the XDG layout.
func legacyDirectory() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "LightLauncher")
}

// migrateIfNeeded moves data from the legacy folder into the current base directory.
// It reports whether a migration ran. Failures are logged and reported as false.
func migrateIfNeeded() (bool, error) {
	legacy := legacyDirectory()
	if _, err := os.Stat(legacy); err != nil {
		return false, nil
	}
	if err := os.MkdirAll(config.GetBaseDirectory(), 0o755); err != nil {
		return false, err
	}

	if err := moveLegacyData(legacy); err != nil {
		logger.Error("Migration", "Failed to migrate legacy LightLauncher data: %v", err)
		return false, nil
	}
	_ = os.RemoveAll(legacy)
	sanitizeExistingConfigs()
	return true, nil
}

func moveLegacyData(legacy string) error {
	if err := migrateLegacySettings(filepath.Join(legacy, "settings.json")); err != nil {
		return err
	}
	if err := os.MkdirAll(config.GetConfigDirectory(), 0o755); err != nil {
		return err
	}
	for _, oldDir := range []string{
		filepath.Join(legacy, "config", "executables"),
		filepath.Join(legacy, "games"),
	} {
		if err := moveFolderEntries(oldDir, config.GetConfigDirectory()); err != nil {
			return err
		}
	}
	if err := moveFolderEntries(filepath.Join(legacy, "prefixes"), config.GetPrefixBaseDirectory()); err != nil {
		return err
	}
	return moveFolderEntries(filepath.Join(legacy, "logs"), config.GetLogsDirectory())
}

// migrateLegacySettings keeps the current settings.json when one already exists.
func migrateLegacySettings(legacySettings string) error {
	if _, err := os.Stat(legacySettings); err != nil {
		return nil
	}
	target := config.SettingsFilePath()
	if _, err := os.Stat(target); err != nil {
		return moveOrCopy(legacySettings, target)
	}
	_ = os.Remove(legacySettings)
	return nil
}

// moveFolderEntries moves each entry that is missing in the target, then removes the legacy folder.
func moveFolderEntries(legacyDir, targetDir string) error {
	if _, err := os.Stat(legacyDir); err != nil {
		return nil
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(legacyDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		src := filepath.Join(legacyDir, entry.Name())
		dest := filepath.Join(targetDir, entry.Name())
		if _, statErr := os.Stat(dest); statErr == nil {
			continue
		}
		if err := moveOrCopy(src, dest); err != nil {
			return err
		}
	}
	_ = os.RemoveAll(legacyDir)
	return nil
}
