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

	"light-launcher/internal/config"
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

var (
	mu          sync.Mutex
	logFile     *os.File
	currentPath string
	minLevel              = LevelDebug
	outWriter   io.Writer = os.Stdout
)

// Init initializes the logger to write to the specified filename inside the app's log directory.
// It also cleans up older log files, keeping only the specified count.
func Init(fileName string, keepCount int) error {
	mu.Lock()
	defer mu.Unlock()

	logDir := config.GetLogsDirectory()
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Close existing log file if open
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}

	// Perform cleanup of old logs in logDir before opening a new file
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

// Close closes the underlying log file handle.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if logFile != nil {
		logFile.Close()
		logFile = nil
		currentPath = ""
	}
}

// SetLevel sets the minimum log level required to output a message.
func SetLevel(level Level) {
	mu.Lock()
	defer mu.Unlock()
	minLevel = level
}

// Log writes a message to the active outputs if its level meets the minimum requirement.
func Log(level Level, message string) {
	mu.Lock()
	defer mu.Unlock()

	if level < minLevel {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level.String(), message)

	if outWriter != nil {
		fmt.Fprint(outWriter, line)
	}

	if logFile != nil {
		_, _ = logFile.WriteString(line)
		_ = logFile.Sync()
	}
}

// Debug logs a debug-level message.
func Debug(format string, args ...interface{}) {
	Log(LevelDebug, fmt.Sprintf(format, args...))
}

// Info logs an info-level message.
func Info(format string, args ...interface{}) {
	Log(LevelInfo, fmt.Sprintf(format, args...))
}

// Warn logs a warning-level message.
func Warn(format string, args ...interface{}) {
	Log(LevelWarn, fmt.Sprintf(format, args...))
}

// Error logs an error-level message.
func Error(format string, args ...interface{}) {
	Log(LevelError, fmt.Sprintf(format, args...))
}

// GetLogPath returns the path to the currently active log file, or empty string.
func GetLogPath() string {
	mu.Lock()
	defer mu.Unlock()
	return currentPath
}

// TrimCurrentLogFile trims the active log file to keep only the last maxLines.
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

	// Close, read, trim, write, and reopen
	file.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		// Reopen file in append mode to ensure we don't break subsequent writes
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
