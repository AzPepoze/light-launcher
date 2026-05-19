package main

import (
	"fmt"
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

	if logFileHandle != nil {
		_, _ = fmt.Fprint(logFileHandle, sb.String())
		_ = logFileHandle.Sync()
	} else {
		fmt.Print(sb.String())
	}
}
