package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"light-launcher/internal/config"
	"light-launcher/internal/logger"

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
	gsW            string
	gsH            string
	gsOutW         string
	gsOutH         string
	gsR            string
	gsFRLimit      string
	gsWindowMode   string
	gsScaler       string
	gsFilter       string
	gsSharpness    string
	gsHDR          bool
	gsAdaptiveSync bool
	gsMangoapp     bool
	gsCustomArgs   string

	// LSFG configuration
	lsfgMult    string
	lsfgDllPath string

	// Memory configuration
	memoryMinValue string
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
	flag.StringVar(&gsW, "gs-w", "", "Width")
	flag.StringVar(&gsH, "gs-h", "", "Height")
	flag.StringVar(&gsOutW, "gs-out-w", "", "Output Width")
	flag.StringVar(&gsOutH, "gs-out-h", "", "Output Height")
	flag.StringVar(&gsR, "gs-r", "", "Refresh Rate")
	flag.StringVar(&gsFRLimit, "gs-fr-limit", "", "Framerate Limit")
	flag.StringVar(&gsWindowMode, "gs-window-mode", "", "Window Mode")
	flag.StringVar(&gsScaler, "gs-scaler", "", "Scaler")
	flag.StringVar(&gsFilter, "gs-filter", "", "Filter")
	flag.StringVar(&gsSharpness, "gs-sharpness", "", "Sharpness")
	flag.BoolVar(&gsHDR, "gs-hdr", false, "Enable HDR")
	flag.BoolVar(&gsAdaptiveSync, "gs-adaptive-sync", false, "Enable Adaptive Sync")
	flag.BoolVar(&gsMangoapp, "gs-mangoapp", false, "Enable Mangoapp")
	flag.StringVar(&gsCustomArgs, "gs-custom-args", "", "Custom Gamescope Arguments")
	flag.BoolVar(&showLogs, "logs", true, "Show terminal logs")
	flag.Parse()

	if gamePath == "" {
		os.Exit(1)
	}

	if prefixPath == "" {
		prefixPath = filepath.Join(config.GetPrefixBaseDirectory(), "Default")
	}

	timestamp := time.Now().Format("20060102-150405")
	exeName := filepath.Base(gamePath)
	logFileName := fmt.Sprintf("%s-%s.log", exeName, timestamp)

	if err := logger.Init(logFileName, 10); err != nil {
		log.Printf("Warning: failed to initialize logger: %v\n", err)
	}

	logPath := logger.GetLogPath()
	if logPath != "" {
		if file, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err == nil {
			log.SetOutput(file)
		}
	}

	_ = logger.TrimCurrentLogFile(500)

	systray.Run(func() { onReady(logPath) }, onExit)
}
