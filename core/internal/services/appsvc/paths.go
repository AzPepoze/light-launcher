package appsvc

import (
	"os"
	"path/filepath"
	"strings"

	"light-launcher/core/internal/config"
)

func settingsFilePath() string {
	return filepath.Join(config.GetBaseDirectory(), "settings.json")
}

// prefixBaseDirectory returns the custom prefix folder when one is set, else the default prefixes folder.
func prefixBaseDirectory(customDir string) string {
	if trimmed := strings.TrimSpace(customDir); trimmed != "" {
		return expandPath(trimmed)
	}
	return config.GetPrefixBaseDirectory()
}

func prefixConfigFilePath(prefixName, customDir string) string {
	return filepath.Join(prefixBaseDirectory(customDir), prefixName, "light-launcher.json")
}

// expandPath resolves a leading "~" to the home folder and normalizes the path.
func expandPath(p string) string {
	trimmed := strings.TrimSpace(p)
	if trimmed == "" {
		return ""
	}
	home, _ := os.UserHomeDir()
	switch {
	case trimmed == "~":
		return home
	case strings.HasPrefix(trimmed, "~/"):
		return filepath.Join(home, trimmed[2:])
	}
	return normalizePath(trimmed)
}

// safeExists reports whether the expanded path exists. It never fails.
func safeExists(p string) bool {
	if p == "" {
		return false
	}
	_, err := os.Stat(expandPath(p))
	return err == nil
}
