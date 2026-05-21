package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"light-launcher/internal/adapter"
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

	showLogs bool
)

func main() {
	flag.StringVar(&gamePath, "game", "", "Path to the game executable")
	flag.StringVar(&launcherPath, "launcher", "", "Path to the launcher executable")
	flag.StringVar(&prefixPath, "prefix", "", "Path to the WINEPREFIX")
	flag.StringVar(&protonPath, "proton-path", "", "Full path to the Proton tool")
	flag.StringVar(&protonPattern, "proton-pattern", "", "Proton pattern for UMU")
	flag.BoolVar(&showLogs, "logs", true, "Show terminal logs")

	for _, a := range adapter.GetAdapters() {
		a.RegisterFlags(flag.CommandLine)
	}

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
