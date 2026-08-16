package logger

import (
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"light-launcher/core/internal/config"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ANSI formatting constants
const (
	AnsiReset = "\x1b[0m"
	AnsiDim   = "\x1b[2m"
	AnsiBold  = "\x1b[1m"
)

var pastelAnsiPalette = []string{
	"\x1b[38;2;255;179;186m", // pastel-pink
	"\x1b[38;2;255;223;186m", // pastel-peach
	"\x1b[38;2;253;255;182m", // pastel-buttercup
	"\x1b[38;2;186;255;201m", // pastel-mint
	"\x1b[38;2;186;225;255m", // pastel-sky
	"\x1b[38;2;232;223;245m", // pastel-lavender
	"\x1b[38;2;181;234;215m", // pastel-seafoam
	"\x1b[38;2;252;225;228m", // pastel-rose
	"\x1b[38;2;189;178;255m", // pastel-violet
	"\x1b[38;2;255;198;255m", // pastel-magenta
	"\x1b[38;2;155;246;255m", // pastel-aqua
	"\x1b[38;2;202;255;191m", // pastel-sage
	"\x1b[38;2;160;196;255m", // pastel-periwinkle
}

var levelAnsi = map[Level]string{
	LevelDebug: "\x1b[38;2;176;190;197m",
	LevelInfo:  "\x1b[38;2;129;212;250m",
	LevelWarn:  "\x1b[38;2;255;224;130m",
	LevelError: "\x1b[38;2;239;154;154m",
}

func hashCategory(category string) int {
	clean := strings.ToLower(strings.TrimSpace(category))
	var hash uint32 = 5381
	for i := 0; i < len(clean); i++ {
		hash = ((hash << 5) + hash) ^ uint32(clean[i])
	}
	return int(hash % uint32(len(pastelAnsiPalette)))
}

func getCategoryAnsi(category string) string {
	if strings.TrimSpace(category) == "" {
		return pastelAnsiPalette[0]
	}
	idx := hashCategory(category)
	idx = int(math.Abs(float64(idx))) % len(pastelAnsiPalette)
	return pastelAnsiPalette[idx]
}

var (
	mu               sync.Mutex
	logFile          *os.File
	currentPath      string
	minLevel                   = LevelDebug
	outWriter        io.Writer = os.Stdout
	notificationIcon string
)

func SetNotificationIcon(iconPath string) {
	mu.Lock()
	defer mu.Unlock()
	notificationIcon = iconPath
}

func Init(fileName string, keepCount int) error {
	mu.Lock()
	defer mu.Unlock()

	logDir := config.GetLogsDirectory()
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	if logFile != nil {
		logFile.Close()
		logFile = nil
	}

	cleanupLogs(logDir, keepCount)

	logPath := filepath.Join(logDir, fileName)
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	logFile = file
	currentPath = logPath
	return nil
}

func Close() {
	mu.Lock()
	defer mu.Unlock()
	if logFile != nil {
		logFile.Close()
		logFile = nil
		currentPath = ""
	}
}

func SetLevel(level Level) {
	mu.Lock()
	defer mu.Unlock()
	minLevel = level
}

func Log(level Level, category string, format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()

	if level < minLevel {
		return
	}

	message := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("15:04:05.000")

	// 1. ANSI Terminal Output
	catColor := getCategoryAnsi(category)
	lvlColor := levelAnsi[level]
	if lvlColor == "" {
		lvlColor = AnsiReset
	}

	timeBadge := fmt.Sprintf("%s%s%s", AnsiDim, timestamp, AnsiReset)
	lvlBadge := fmt.Sprintf("%s%-5s%s", lvlColor, level.String(), AnsiReset)
	catBadge := fmt.Sprintf("%s%s[%s]%s", catColor, AnsiBold, category, AnsiReset)

	terminalLine := fmt.Sprintf("%s %s %s %s\n", timeBadge, lvlBadge, catBadge, message)
	if outWriter != nil {
		fmt.Fprint(outWriter, terminalLine)
	}

	// 2. File Output (Plain text without ANSI codes)
	plainLine := fmt.Sprintf("[%s] [%-5s] [%s] %s\n", timestamp, level.String(), category, message)
	if logFile != nil {
		_, _ = logFile.WriteString(plainLine)
		_ = logFile.Sync()
	}

	// 3. Desktop Notification on Error
	if level == LevelError {
		go func(cat, msg string) {
			title := fmt.Sprintf("LightLauncher [%s] Error", cat)
			if notificationIcon != "" {
				_ = exec.Command("notify-send", "-a", "LightLauncher", "-i", notificationIcon, title, msg).Run()
			} else {
				_ = exec.Command("notify-send", "-a", "LightLauncher", title, msg).Run()
			}
		}(category, message)
	}
}

func Debug(category string, format string, args ...interface{}) {
	Log(LevelDebug, category, format, args...)
}

func Info(category string, format string, args ...interface{}) {
	Log(LevelInfo, category, format, args...)
}

func Warn(category string, format string, args ...interface{}) {
	Log(LevelWarn, category, format, args...)
}

func Error(category string, format string, args ...interface{}) {
	Log(LevelError, category, format, args...)
}

func GetLogPath() string {
	mu.Lock()
	defer mu.Unlock()
	return currentPath
}

func TrimCurrentLogFile(maxLines int) error {
	mu.Lock()
	path := currentPath
	file := logFile
	mu.Unlock()

	if path == "" || file == nil {
		return nil
	}

	mu.Lock()
	defer mu.Unlock()

	file.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		reopenFile(path)
		return err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) <= maxLines {
		return reopenFile(path)
	}

	trimmed := lines[len(lines)-maxLines:]
	trimmedData := strings.Join(trimmed, "\n")

	if err := os.WriteFile(path, []byte(trimmedData), 0666); err != nil {
		reopenFile(path)
		return err
	}

	return reopenFile(path)
}

func reopenFile(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logFile = file
	}
	return err
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
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})
	toDelete := len(files) - keep
	for i := 0; i < toDelete; i++ {
		_ = os.Remove(filepath.Join(dir, files[i].Name()))
	}
}
