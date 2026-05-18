package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/getlantern/systray"
)

var (
	// Game and launcher paths
	gamePath      string
	launcherPath  string
	prefixPath    string
	protonPath    string
	protonPattern string

	// Feature flags
	mango     bool
	gamemode  bool
	gamescope bool
	lsfg      bool
	lsfgPerf  bool
	memoryMin bool
	showLogs  bool

	// Gamescope configuration
	gsW string
	gsH string
	gsR string

	// LSFG configuration
	lsfgMult    string
	lsfgDllPath string

	// Memory configuration
	memoryMinValue string

	// Logging
	logFileHandle *os.File
)

func main() {
	flag.StringVar(&gamePath, "game", "", "Path to the game executable")
	flag.StringVar(&launcherPath, "launcher", "", "Path to the launcher executable")
	flag.StringVar(&prefixPath, "prefix", "", "Path to the WINEPREFIX")
	flag.StringVar(&protonPath, "proton-path", "", "Full path to the Proton tool")
	flag.StringVar(&protonPattern, "proton-pattern", "", "Proton pattern for UMU")
	flag.BoolVar(&mango, "mango", false, "Enable MangoHud")
	flag.BoolVar(&gamemode, "gamemode", false, "Enable GameMode")
	flag.BoolVar(&gamescope, "gamescope", false, "Enable Gamescope")
	flag.BoolVar(&lsfg, "lsfg", false, "Enable LSFG-VK")
	flag.StringVar(&lsfgMult, "lsfg-mult", "2", "LSFG Multiplier")
	flag.BoolVar(&lsfgPerf, "lsfg-perf", false, "Enable LSFG Performance Mode")
	flag.StringVar(&lsfgDllPath, "lsfg-dll-path", "", "Path to Lossless.dll")
	flag.BoolVar(&memoryMin, "memory-min", false, "Enable Memory Protection (min RAM)")
	flag.StringVar(&memoryMinValue, "memory-min-value", "", "Memory Protection Value (e.g. 4G)")
	flag.StringVar(&gsW, "gs-w", "1920", "Width")
	flag.StringVar(&gsH, "gs-h", "1080", "Height")
	flag.StringVar(&gsR, "gs-r", "60", "Refresh Rate")
	flag.BoolVar(&showLogs, "logs", true, "Show terminal logs")
	flag.Parse()

	if gamePath == "" {
		os.Exit(1)
	}

	if prefixPath == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			prefixPath = filepath.Join(home, "LightLauncher", "prefixes", "Default")
		}
	}

	logPath := getLogPath()
	var err error
	logFileHandle, err = os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFileHandle)
		// Trim log file to last 500 lines to keep queue behavior
		_ = trimLogFile(logPath, 500)
	}

	systray.Run(func() { onReady(logPath) }, onExit)
}
