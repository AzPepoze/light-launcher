package config

import (
	"path/filepath"

	"light-launcher/core/internal/types"
)

// defaultExcludeNames are the file name patterns skipped by executable scans.
var defaultExcludeNames = []string{
	"UnityCrashHandler*.exe",
	"unins*.exe",
	"*redist*.exe",
	"DXSETUP.exe",
	"vc_redist*.exe",
	"CrashReport*.exe",
	"UE4PrereqSetup*.exe",
	"Uninstall*.exe",
	"setup*.exe",
	"EpicGamesLauncher.exe",
	"Steam.exe",
	"Touchup.exe",
	"Cleanup.exe",
	"EasyAntiCheat*.exe",
	"BEService*.exe",
	"dotNetFx*.exe",
	"vcredist*.exe",
	"installer*.exe",
}

// legacyExcludeSets are old default lists that are replaced by defaultExcludeNames on load.
var legacyExcludeSets = [][]string{
	{"UnityCrashHandler64", "uninstall", "redist"},
}

const (
	defaultPluginIdleMinutes = 5
	maxPluginIdleMinutes     = 1440
)

// defaultPluginIdleTimeout is the idle stop time when the settings file has no value.
func defaultPluginIdleTimeout() *int {
	minutes := defaultPluginIdleMinutes
	return &minutes
}

// DefaultScanExcludeNames returns a fresh copy of the executable scan excludes.
func DefaultScanExcludeNames() []string {
	return CloneStrings(defaultExcludeNames)
}

// DefaultExtras is the extras block every new launch config starts with.
func DefaultExtras() types.ExtrasConfig {
	return types.ExtrasConfig{
		Lsfg: types.LsfgConfig{
			Multiplier: DefaultMultiplier,
			FlowScale:  "1.0",
			Pacing:     "smooth",
		},
		Gamescope: types.GamescopeConfig{
			Width:       DefaultWidth,
			Height:      DefaultHeight,
			RefreshRate: DefaultRefreshRate,
			WindowMode:  "borderless",
			Scaler:      "auto",
			Filter:      "linear",
			Sharpness:   "0",
		},
		Memory: types.MemoryConfig{
			Value: DefaultMemoryValue,
		},
	}
}

// DefaultSettings is the settings.json content for a new install.
func DefaultSettings() types.AppSettings {
	return types.AppSettings{
		TransparentMode:          true,
		NativeWayland:            false,
		ScanFolders:              []string{},
		ScanFolderConfigs:        []types.ScanFolderConfig{},
		Blacklist:                []string{},
		PluginIdleTimeoutMinutes: defaultPluginIdleTimeout(),
	}
}

// DefaultPrefixOptions uses the default prefixes folder, even when a custom folder is set.
func DefaultPrefixOptions(prefixName string) types.LaunchOptions {
	return types.LaunchOptions{
		Name:       prefixName,
		PrefixPath: filepath.Join(GetPrefixBaseDirectory(), prefixName),
		Extras:     DefaultExtras(),
	}
}

// CloneStrings copies a string slice, keeping nil as nil.
func CloneStrings(values []string) []string {
	if values == nil {
		return nil
	}
	return append(make([]string, 0, len(values)), values...)
}
