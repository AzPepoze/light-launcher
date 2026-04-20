package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func getLogPath() string {
	homeDir := os.ExpandEnv("$HOME")
	logDir := filepath.Join(homeDir, "LightLauncher/logs")
	os.MkdirAll(logDir, 0755)
	cleanupLogs(logDir, 10)
	timestamp := time.Now().Format("20060102-150405")
	exeName := filepath.Base(gamePath)
	return filepath.Join(logDir, fmt.Sprintf("%s-%s.log", exeName, timestamp))
}

// trimLogFile keeps only the last 500 lines of a log file (queue behavior)
func trimLogFile(filePath string, maxLines int) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) <= maxLines {
		return nil
	}

	// Keep only the last maxLines
	trimmed := lines[len(lines)-maxLines:]
	trimmedData := strings.Join(trimmed, "\n")

	return os.WriteFile(filePath, []byte(trimmedData), 0666)
}

func cleanupLogs(dir string, keep int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var files []os.FileInfo
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".log") {
			info, err := entry.Info()
			if err == nil {
				files = append(files, info)
			}
		}
	}
	if len(files) <= keep {
		return
	}
	sort.Slice(files, func(i, j int) bool { return files[i].ModTime().Before(files[j].ModTime()) })
	toDelete := len(files) - keep
	for i := 0; i < toDelete; i++ {
		os.Remove(filepath.Join(dir, files[i].Name()))
	}
}

// logGameStartup logs the command and enabled features
func logGameStartup(cmdArgs []string) {
	log.Printf("--- EXECUTION START ---")
	log.Printf("COMMAND: %s", strings.Join(cmdArgs, " "))
	log.Printf("ENABLED FEATURES:")

	if mango {
		log.Printf("  [+] MangoHud")
	}
	if gamemode {
		log.Printf("  [+] GameMode")
	}
	if gamescope {
		log.Printf("  [+] Gamescope (%sx%s@%s)", gsW, gsH, gsR)
	}
	if lsfg {
		log.Printf("  [+] LSFG-VK (x%s, PerfMode:%v)", lsfgMult, lsfgPerf)
	}
	if memoryMin {
		log.Printf("  [+] Memory Protection (Min: %s)", memoryMinValue)
	}

	log.Printf("-----------------------")

	// Sync to disk so terminal sees output immediately
	if logFileHandle != nil {
		_ = logFileHandle.Sync()
	}
}
