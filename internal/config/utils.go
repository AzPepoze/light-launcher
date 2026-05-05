package config

import (
	"os/user"
	"path/filepath"
	"strings"
)

func ExpandPath(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		currentUser, err := user.Current()
		if err != nil {
			return path
		}
		if path == "~" {
			return currentUser.HomeDir
		}
		return filepath.Join(currentUser.HomeDir, path[2:])
	}
	return path
}
