package runtimesvc

import (
	"os"
	"path/filepath"
	"strings"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

// scanProtonVersions lists Steam Proton installs and custom Proton builds. Each real folder is listed once.
func scanProtonVersions() []types.ProtonTool {
	home, err := os.UserHomeDir()
	if err != nil {
		return []types.ProtonTool{}
	}

	steamCommonPaths := []string{
		filepath.Join(home, ".steam/root/steamapps/common"),
		filepath.Join(home, ".local/share/Steam/steamapps/common"),
		filepath.Join(home, ".var/app/com.valvesoftware.Steam/.local/share/Steam/steamapps/common"),
	}
	customToolPaths := []string{
		filepath.Join(home, ".steam/root/compatibilitytools.d"),
		filepath.Join(home, ".local/share/Steam/compatibilitytools.d"),
		filepath.Join(home, ".var/app/com.valvesoftware.Steam/.local/share/Steam/compatibilitytools.d"),
	}

	seen := map[string]bool{}
	tools := []types.ProtonTool{}
	tools = appendProtonTools(tools, seen, steamCommonPaths, true, "Steam")
	tools = appendProtonTools(tools, seen, customToolPaths, false, "Custom")
	return tools
}

func appendProtonTools(tools []types.ProtonTool, seen map[string]bool, basePaths []string, isSteam bool, displayPrefix string) []types.ProtonTool {
	for _, basePath := range basePaths {
		if !config.PathExists(basePath) {
			continue
		}
		entries, err := os.ReadDir(basePath)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			if isSteam && !strings.HasPrefix(strings.ToLower(entry.Name()), "proton") {
				continue
			}

			protonPath := filepath.Join(basePath, entry.Name())
			realPath := resolveRealPath(protonPath)
			if seen[realPath] {
				continue
			}
			if !config.PathExists(filepath.Join(protonPath, "proton")) {
				continue
			}

			seen[realPath] = true
			tools = append(tools, types.ProtonTool{
				Name:        entry.Name(),
				Path:        protonPath,
				IsSteam:     isSteam,
				DisplayName: displayPrefix + ": " + entry.Name(),
			})
		}
	}
	return tools
}

// resolveRealPath follows symbolic links so one install reached through two folders is listed once.
func resolveRealPath(target string) string {
	resolved, err := filepath.EvalSymlinks(target)
	if err != nil {
		return target
	}
	return resolved
}
