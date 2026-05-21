package main

import (
	"fmt"
	"strings"

	"light-launcher/internal/logger"
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
func logGameStartup(cmdArgs []string) {
	var sb strings.Builder
	sb.WriteString(banner)
	sb.WriteString("\n======================================================================\n")
	sb.WriteString("                         EXECUTION CONTEXT                            \n")
	sb.WriteString("======================================================================\n\n")

	sb.WriteString("[ ENABLED FEATURES ]\n")
	hasFeatures := false
	if mango {
		sb.WriteString("  ✓ MangoHud\n")
		hasFeatures = true
	}
	if gamemode {
		sb.WriteString("  ✓ GameMode\n")
		hasFeatures = true
	}
	if gamescope {
		sb.WriteString(fmt.Sprintf("  ✓ Gamescope (%sx%s@%s)\n", gsW, gsH, gsR))
		hasFeatures = true
	}
	if lsfg {
		sb.WriteString(fmt.Sprintf("  ✓ LSFG-VK (x%s, PerfMode:%v)\n", lsfgMult, lsfgPerf))
		hasFeatures = true
	}
	if memoryMin {
		sb.WriteString(fmt.Sprintf("  ✓ Memory Protection (Min: %s)\n", memoryMinValue))
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
