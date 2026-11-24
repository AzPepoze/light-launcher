package config

import (
	"path/filepath"
)

func GetBaseDirectory() string {
	return filepath.Join(configHome(), AppName)
}

func GetConfigDirectory() string {
	return filepath.Join(GetBaseDirectory(), ConfigDirName)
}

func GetPrefixBaseDirectory() string {
	return filepath.Join(GetBaseDirectory(), PrefixesDirName)
}

func GetLogsDirectory() string {
	return filepath.Join(GetBaseDirectory(), LogsDirName)
}

func SettingsFilePath() string {
	return filepath.Join(GetBaseDirectory(), "settings.json")
}

func ActivityFilePath() string {
	return filepath.Join(GetBaseDirectory(), "activity.json")
}

// DriveSetupFilePath is the launcher-wide list of drives linked into every prefix.
func DriveSetupFilePath() string {
	return filepath.Join(GetBaseDirectory(), "drive-setup.json")
}
