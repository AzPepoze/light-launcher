package appsvc

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/logger"
)

// legacyPrefixMarker is the part of an old prefix path that points into the legacy folder.
const legacyPrefixMarker = "/LightLauncher/prefixes/"

// sanitizeExistingConfigs rewrites legacy prefix paths in game configs, then fixes prefix symlinks.
func sanitizeExistingConfigs() {
	configDir := config.GetConfigDirectory()
	if _, err := os.Stat(configDir); err != nil {
		return
	}
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return
	}
	basePrefixDir := config.GetPrefixBaseDirectory()
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		configPath := filepath.Join(configDir, entry.Name(), "config.json")
		if _, err := os.Stat(configPath); err != nil {
			continue
		}
		_ = repairLaunchConfig(configPath, basePrefixDir)
	}
	repairPrefixSymlinks()
}

// repairLaunchConfig points a legacy PrefixPath at the current prefixes folder.
// Numbers keep their exact text, and the file is rewritten only when something changed.
func repairLaunchConfig(configPath, basePrefixDir string) error {
	raw, err := os.ReadFile(configPath)
	if err != nil || len(bytes.TrimSpace(raw)) == 0 {
		return err
	}
	cfg, err := decodeKeepingNumbers(raw)
	if err != nil {
		return err
	}
	prefixPath, ok := cfg["PrefixPath"].(string)
	if !ok || prefixPath == "" || !strings.Contains(prefixPath, legacyPrefixMarker) {
		return nil
	}
	cfg["PrefixPath"] = filepath.Join(basePrefixDir, baseName(prefixPath))
	data, err := marshalJSON(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0o644)
}

func decodeKeepingNumbers(raw []byte) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var cfg map[string]any
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// repairPrefixSymlinks re-points prefix symlinks that still target the legacy folder.
func repairPrefixSymlinks() {
	basePrefixDir := config.GetPrefixBaseDirectory()
	if _, err := os.Stat(basePrefixDir); err != nil {
		return
	}
	entries, err := os.ReadDir(basePrefixDir)
	if err != nil {
		logger.Error("Migration", "Failed to repair prefix symlinks: %v", err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			walkRepairSymlinks(filepath.Join(basePrefixDir, entry.Name()), basePrefixDir)
		}
	}
}

func walkRepairSymlinks(dir, basePrefixDir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		switch {
		case entry.Type()&os.ModeSymlink != 0:
			_ = repairSymlink(fullPath, basePrefixDir)
		case entry.IsDir():
			walkRepairSymlinks(fullPath, basePrefixDir)
		}
	}
}

// repairSymlink replaces a link into the legacy folder with one into basePrefixDir.
func repairSymlink(linkPath, basePrefixDir string) error {
	target, err := os.Readlink(linkPath)
	if err != nil || !strings.Contains(target, legacyPrefixMarker) {
		return err
	}
	// The last occurrence wins, like the greedy pattern the Electron version used.
	index := strings.LastIndex(target, legacyPrefixMarker)
	fixed := basePrefixDir + "/" + target[index+len(legacyPrefixMarker):]
	if err := os.Remove(linkPath); err != nil {
		return err
	}
	return os.Symlink(fixed, linkPath)
}
