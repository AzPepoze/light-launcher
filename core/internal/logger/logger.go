package logger

import (
	"fmt"
	"io"
	"os"
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

const timeLayout = "15:04:05.000"

var levelAnsi = map[Level]string{
	LevelDebug: "\x1b[38;2;176;190;197m",
	LevelInfo:  "\x1b[38;2;129;212;250m",
	LevelWarn:  "\x1b[38;2;255;224;130m",
	LevelError: "\x1b[38;2;239;154;154m",
}

var (
	mu               sync.Mutex
	logFile          *os.File
	currentPath      string
	minLevel                   = LevelDebug
	outWriter        io.Writer = os.Stderr
	colorOn                    = colorEnabled(os.Stderr, os.LookupEnv)
	notificationIcon string
	notifier         func(msg string)
)

// SetNotificationIcon sets the icon that NotifySend shows. It does not turn on notifications by itself.
func SetNotificationIcon(iconPath string) {
	mu.Lock()
	defer mu.Unlock()
	notificationIcon = iconPath
}

// SetNotifier sets the function that gets every error message. Nil, the default, means no notifications.
func SetNotifier(fn func(msg string)) {
	mu.Lock()
	defer mu.Unlock()
	notifier = fn
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

// Log writes one line to stderr and to the log file. Error lines also go to the notifier, if one is set.
func Log(level Level, category string, format string, args ...interface{}) {
	if !levelEnabled(level) {
		return
	}
	message := fmt.Sprintf(format, args...)
	notify := writeLine(level, category, message)
	if notify != nil {
		notify(fmt.Sprintf("[%s] %s", category, message))
	}
}

func levelEnabled(level Level) bool {
	mu.Lock()
	defer mu.Unlock()
	return level >= minLevel
}

// writeLine writes the line under the lock. It returns the notifier for error lines, so it runs after unlocking.
func writeLine(level Level, category string, message string) func(msg string) {
	mu.Lock()
	defer mu.Unlock()

	timestamp := time.Now().Format(timeLayout)
	if outWriter != nil {
		fmt.Fprint(outWriter, terminalLine(timestamp, level, category, message))
	}
	if logFile != nil {
		_, _ = fmt.Fprintf(logFile, "[%s] [%-5s] [%s] %s\n", timestamp, level.String(), category, message)
		_ = logFile.Sync()
	}
	if level == LevelError {
		return notifier
	}
	return nil
}

// terminalLine formats one stderr line. Colors are added only when colorOn is set.
func terminalLine(timestamp string, level Level, category string, message string) string {
	if !colorOn {
		return fmt.Sprintf("%s %-5s [%s] %s\n", timestamp, level.String(), category, message)
	}
	timeBadge := AnsiDim + timestamp + AnsiReset
	levelBadge := levelAnsi[level] + fmt.Sprintf("%-5s", level.String()) + AnsiReset
	categoryBadge := categoryColor(category) + AnsiBold + "[" + category + "]" + AnsiReset
	return fmt.Sprintf("%s %s %s %s\n", timeBadge, levelBadge, categoryBadge, message)
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
