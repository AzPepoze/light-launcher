package runtimesvc

import (
	"os"
	"path/filepath"
	"strings"

	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/types"
)

const defaultLaunchPrefix = "Default"

// resolvePrefixPath moves the prefix into the current prefix folder and creates it when missing.
func resolvePrefixPath(launch *types.LaunchOptions) error {
	basePrefixDir := prefixBaseFolder()
	if launch.PrefixPath == "" || strings.Contains(launch.PrefixPath, legacyPrefixesMarker) {
		prefixName := defaultLaunchPrefix
		if launch.PrefixPath != "" {
			prefixName = nodeBasename(launch.PrefixPath)
		}
		launch.PrefixPath = filepath.Join(basePrefixDir, prefixName)
	} else if !pathExists(launch.PrefixPath) {
		candidate := filepath.Join(basePrefixDir, nodeBasename(launch.PrefixPath))
		if pathExists(candidate) {
			launch.PrefixPath = candidate
		}
	}

	launch.PrefixPath = expandPath(launch.PrefixPath)
	if !pathExists(launch.PrefixPath) {
		if err := os.MkdirAll(launch.PrefixPath, 0755); err != nil {
			return err
		}
	}
	repairPrefixSymlinks()
	return nil
}

// resolveProtonPath picks the Proton build. A custom Proton wins, then the prefix default, then the first install.
func resolveProtonPath(launch *types.LaunchOptions) error {
	if !launch.UseCustomProton {
		launch.ProtonPath = prefixDefaultProton(launch.PrefixPath)
	}

	tools := scanProtonVersions()
	if launch.ProtonPath == "" {
		if len(tools) == 0 {
			return errNoProtonRuntime
		}
		launch.ProtonPath = tools[0].Path
		return nil
	}

	if match := findProtonMatch(launch.ProtonPath, tools); match != nil {
		launch.ProtonPath = match.Path
		return nil
	}
	if pathExists(expandPath(launch.ProtonPath)) {
		return nil
	}
	if len(tools) == 0 {
		return errNoProtonRuntime
	}
	logger.Warn("Runner", "Configured Proton \"%s\" not found. Falling back to \"%s\".", launch.ProtonPath, tools[0].DisplayName)
	launch.ProtonPath = tools[0].Path
	return nil
}

// prefixDefaultProton returns the Proton path saved in the prefix's light-launcher.json, or "".
func prefixDefaultProton(prefixPath string) string {
	prefixConfigFile := filepath.Join(prefixPath, "light-launcher.json")
	if !pathExists(prefixConfigFile) {
		return ""
	}
	prefixConfig, err := loadJSON[types.LaunchOptions](prefixConfigFile)
	if err != nil {
		logger.Warn("Runner", "Failed to read prefix runtime defaults: %v", err)
		return ""
	}
	return prefixConfig.ProtonPath
}

// syncLsfgProfile writes the LSFG profile when LSFG is on for this game, or disables it when off.
func syncLsfgProfile(launch types.LaunchOptions) error {
	lsfg := launch.Extras.Lsfg
	if !lsfg.Enabled {
		return disableLsfgProfile(launch.Name, launch.GamePath)
	}

	gpu := lsfg.Gpu
	if gpu == "" {
		if gpus := listGpuNames(); len(gpus) > 0 {
			gpu = gpus[0]
		}
	}
	return saveLsfgProfile(SaveLsfgProfileInput{
		ProfileName:     launch.Name,
		GamePath:        launch.GamePath,
		Multiplier:      parseLaunchMultiplier(lsfg.Multiplier),
		PerformanceMode: lsfg.PerfMode,
		DllPath:         lsfg.DllPath,
		GPU:             gpu,
		FlowScale:       lsfg.FlowScale,
		Pacing:          lsfg.Pacing,
		AllowFp16:       lsfg.AllowFp16,
	})
}

// parseLaunchMultiplier reads the multiplier text. Empty or unreadable text means 2.
func parseLaunchMultiplier(value string) int {
	if number, ok := parseIntLike(stringOrDefault(value, "2")); ok {
		return number
	}
	return 2
}
