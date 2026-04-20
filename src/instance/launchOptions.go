package main

import "light-launcher/pkg/core"

// buildLaunchOptions creates the launch options from command line flags
func buildLaunchOptions() core.LaunchOptions {
	return core.LaunchOptions{
		MainExecutablePath: gamePath,
		LauncherPath:       launcherPath,
		PrefixPath:         prefixPath,
		ProtonPattern:      protonPattern,
		ProtonPath:         protonPath,
		EnableMangoHud:     mango,
		EnableGamemode:     gamemode,
		EnableGamescope:    gamescope,
		GamescopeW:         gsW,
		GamescopeH:         gsH,
		GamescopeR:         gsR,
		EnableLsfgVk:       lsfg,
		LsfgMultiplier:     lsfgMult,
		LsfgPerfMode:       lsfgPerf,
		LsfgDllPath:        lsfgDllPath,
		EnableMemoryMin:    memoryMin,
		MemoryMinValue:     memoryMinValue,
	}
}
