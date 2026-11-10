package appsvc

import (
	"light-launcher/core/internal/config"
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

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}
	return append(make([]string, 0, len(values)), values...)
}

func defaultScanExcludeNames() []string {
	return cloneStrings(defaultExcludeNames)
}

// defaultExtras is the extras block every new launch config starts with.
func defaultExtras() types.ExtrasConfig {
	return types.ExtrasConfig{
		Lsfg: types.LsfgConfig{
			Multiplier: config.DefaultMultiplier,
			FlowScale:  "1.0",
			Pacing:     "smooth",
		},
		Gamescope: types.GamescopeConfig{
			Width:       config.DefaultWidth,
			Height:      config.DefaultHeight,
			RefreshRate: config.DefaultRefreshRate,
			WindowMode:  "borderless",
			Scaler:      "auto",
			Filter:      "linear",
			Sharpness:   "0",
		},
		Memory: types.MemoryConfig{
			Value: config.DefaultMemoryValue,
		},
	}
}

func defaultSettings() *types.AppSettings {
	return &types.AppSettings{
		TransparentMode:   true,
		NativeWayland:     false,
		ScanFolders:       []string{},
		ScanFolderConfigs: []types.ScanFolderConfig{},
		Blacklist:         []string{},
	}
}
