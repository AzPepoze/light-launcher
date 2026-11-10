package runtimesvc

import (
	"os"
	"path/filepath"
	"strings"

	"light-launcher/core/internal/config"
)

// expandPath trims the value, expands a leading "~" and cleans the path. Empty input gives "".
func expandPath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	return filepath.Clean(config.ExpandPath(trimmed))
}

// nodeBasename mirrors path.basename: an empty input stays empty instead of ".".
func nodeBasename(value string) string {
	if strings.Trim(value, "/") == "" {
		return ""
	}
	return filepath.Base(value)
}

// nameWithoutExt mirrors path.parse(p).name: a leading dot is part of the name, not an extension.
func nameWithoutExt(base string) string {
	index := strings.LastIndex(base, ".")
	if index <= 0 {
		return base
	}
	return base[:index]
}

// matchKey normalizes a path for case-insensitive comparison. An empty path stays empty.
func matchKey(value string) string {
	if value == "" {
		return ""
	}
	return strings.ToLower(filepath.Clean(value))
}

func pathExists(target string) bool {
	_, err := os.Stat(target)
	return err == nil
}

// prefixBaseDir returns the custom prefix folder when set, else the default prefixes folder.
func prefixBaseDir(customDir string) string {
	if strings.TrimSpace(customDir) != "" {
		return expandPath(customDir)
	}
	return config.GetPrefixBaseDirectory()
}

func prefixConfigPath(prefixName string, customDir string) string {
	return filepath.Join(prefixBaseDir(customDir), prefixName, "light-launcher.json")
}
