package main

import (
	"fmt"
	"strings"

	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/types"
)

const banner = `
   __    _       __    __    __                      __           
  / /   (_)___ _/ /_  / /_  / /   ____ ___  ______  / /_  ___  _____
 / /   / / __ '/ __ \/ __/ / /   / __ '/ / / / __ \/ __ \/ _ \/ ___/
/ /___/ / /_/ / / / / /_  / /___/ /_/ / /_/ / / / / / / /  __/ /    
\____/_/\__, /_/ /_/\__/ /_____/\__,_/\__,_/_/ /_/_/ /_/\___/_/     
       /____/                                                       
`

// logGameStartup logs the command and enabled features
func logGameStartup(cmdArgs []string, options types.LaunchOptions) {
	var sb strings.Builder
	sb.WriteString(banner)
	sb.WriteString("\n======================================================================\n")
	sb.WriteString("                         EXECUTION CONTEXT                            \n")
	sb.WriteString("======================================================================\n\n")

	sb.WriteString("[ ENABLED FEATURES ]\n")
	hasFeatures := false
	if options.Extras.EnableMangoHud {
		sb.WriteString("  ✓ MangoHud\n")
		hasFeatures = true
	}
	if options.Extras.EnableGamemode {
		sb.WriteString("  ✓ GameMode\n")
		hasFeatures = true
	}
	if options.Extras.Gamescope.Enabled {
		sb.WriteString(fmt.Sprintf("  ✓ Gamescope (%sx%s@%s)\n",
			options.Extras.Gamescope.Width,
			options.Extras.Gamescope.Height,
			options.Extras.Gamescope.RefreshRate))
		hasFeatures = true
	}
	if options.Extras.Lsfg.Enabled {
		sb.WriteString(fmt.Sprintf("  ✓ LSFG-VK (x%s, PerfMode:%v)\n",
			options.Extras.Lsfg.Multiplier,
			options.Extras.Lsfg.PerfMode))
		hasFeatures = true
	}
	if options.Extras.Memory.Enabled {
		sb.WriteString(fmt.Sprintf("  ✓ Memory Protection (Min: %s)\n", options.Extras.Memory.Value))
		hasFeatures = true
	}
	if !hasFeatures {
		sb.WriteString("  None\n")
	}

	sb.WriteString("\n[ COMMAND ]\n")
	sb.WriteString("  " + strings.Join(cmdArgs, " ") + "\n")
	sb.WriteString("\n======================================================================\n\n")

	logger.Log(logger.LevelInfo, sb.String())
}
