package adapter

import (
	"light-launcher/core/internal/types"
	"light-launcher/core/internal/utils/gamemode"
	"light-launcher/core/internal/utils/gamescope"
	"light-launcher/core/internal/utils/lsfg"
	"light-launcher/core/internal/utils/mangohud"
	"light-launcher/core/internal/utils/memory"
)

var cachedAdapters []types.LaunchAdapter

func init() {
	cachedAdapters = []types.LaunchAdapter{
		lsfg.NewAdapter(),
		mangohud.NewAdapter(),
		gamescope.NewAdapter(),
		gamemode.NewAdapter(),
		memory.NewAdapter(),
	}
}

// GetAdapters returns the list of all registered tool adapters.
// The wrapping order is important:
// 1. Gamescope wraps first (inner wrapper).
// 2. GameMode wraps next.
// 3. Memory wraps last (outer wrapper).
// This matches the original execution builder design.
func GetAdapters() []types.LaunchAdapter {
	return cachedAdapters
}
