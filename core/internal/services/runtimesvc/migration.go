package runtimesvc

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/logger"
)

const legacyPrefixesMarker = "/LightLauncher/prefixes/"

var legacyPrefixesPattern = regexp.MustCompile(`^.*/LightLauncher/prefixes/`)

// repairPrefixSymlinks points symbolic links from an old prefix folder at the current prefix folder.
// Old installs stored links under a LightLauncher folder that no longer exists.
func repairPrefixSymlinks() {
	baseDir := config.GetPrefixBaseDirectory()
	if !config.PathExists(baseDir) {
		return
	}
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		logger.Error("Migration", "Failed to repair prefix symlinks: %v", err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			repairSymlinksIn(filepath.Join(baseDir, entry.Name()), baseDir)
		}
	}
}

func repairSymlinksIn(directory string, baseDir string) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return
	}
	for _, entry := range entries {
		fullPath := filepath.Join(directory, entry.Name())
		switch {
		case entry.Type()&os.ModeSymlink != 0:
			repairSymlink(fullPath, baseDir)
		case entry.IsDir():
			repairSymlinksIn(fullPath, baseDir)
		}
	}
}

func repairSymlink(linkPath string, baseDir string) {
	target, err := os.Readlink(linkPath)
	if err != nil || !strings.Contains(target, legacyPrefixesMarker) {
		return
	}
	fixedTarget := legacyPrefixesPattern.ReplaceAllLiteralString(target, baseDir+"/")
	if err := os.Remove(linkPath); err != nil {
		return
	}
	_ = os.Symlink(fixedTarget, linkPath)
}
