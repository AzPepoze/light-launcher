package config

import (
	"os"
	"path/filepath"
)

// configHome returns %APPDATA%, or the default Roaming folder when it is not set.
func configHome() string {
	appData := os.Getenv("APPDATA")
	if appData != "" {
		return appData
	}
	homeDirectory, _ := os.UserHomeDir()
	return filepath.Join(homeDirectory, "AppData", "Roaming")
}
