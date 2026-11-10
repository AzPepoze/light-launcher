package config

import (
	"os"
	"path/filepath"
)

// configHome returns $XDG_CONFIG_HOME, or ~/.config when it is not set.
func configHome() string {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		return xdgConfigHome
	}
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return ".config"
	}
	return filepath.Join(homeDirectory, ".config")
}
