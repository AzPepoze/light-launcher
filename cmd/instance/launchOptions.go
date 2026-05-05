package main

import "light-launcher/internal/types"

// buildLaunchOptions creates the launch options from command line flags
func buildLaunchOptions() types.LaunchOptions {
	return types.LaunchOptions{
		GamePath:      gamePath,
		RunnerPath:    launcherPath,
		PrefixPath:    prefixPath,
		ProtonPattern: protonPattern,
		ProtonPath:    protonPath,
		Extras: types.ExtrasConfig{
			EnableMangoHud: mango,
			EnableGamemode: gamemode,
			Lsfg: types.LsfgConfig{
				Enabled:    lsfg,
				Multiplier: lsfgMult,
				PerfMode:   lsfgPerf,
				DllPath:    lsfgDllPath,
			},
			Memory: types.MemoryConfig{
				Enabled: memoryMin,
				Value:   memoryMinValue,
			},
			Gamescope: types.GamescopeConfig{
				Enabled:     gamescope,
				Width:       gsW,
				Height:      gsH,
				RefreshRate: gsR,
			},
		},
	}
}
