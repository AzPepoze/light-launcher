package adapter

import (
	"light-launcher/core/internal/platform"
	"light-launcher/core/internal/types"
	"light-launcher/core/internal/utils/gamemode"
	"light-launcher/core/internal/utils/gamescope"
	"light-launcher/core/internal/utils/lsfg"
	"light-launcher/core/internal/utils/mangohud"
	"light-launcher/core/internal/utils/memory"
)

var cachedAdapters []types.LaunchAdapter

func init() {
	cachedAdapters = enabledAdapters(platform.Current())
}

// enabledAdapters lists the tools the platform supports, in wrapping order.
// Memory has no platform feature. Its wrapper only runs when systemd-run exists.
func enabledAdapters(features platform.Features) []types.LaunchAdapter {
	var adapters []types.LaunchAdapter
	if features.Lsfg {
		adapters = append(adapters, lsfg.NewAdapter())
	}
	if features.MangoHud {
		adapters = append(adapters, mangohud.NewAdapter())
	}
	if features.Gamescope {
		adapters = append(adapters, gamescope.NewAdapter())
	}
	if features.GameMode {
		adapters = append(adapters, gamemode.NewAdapter())
	}
	return append(adapters, memory.NewAdapter())
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
