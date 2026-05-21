package main

import (
	"light-launcher/internal/adapter"
	"light-launcher/internal/types"
)

// buildLaunchOptions creates the launch options from command line flags
func buildLaunchOptions() types.LaunchOptions {
	opts := types.LaunchOptions{
		GamePath:     gamePath,
		LauncherPath: launcherPath,
		PrefixPath:   prefixPath,
		ProtonPath:   protonPath,
	}

	for _, a := range adapter.GetAdapters() {
		a.ExtractOptions(&opts)
	}

	return opts
}
