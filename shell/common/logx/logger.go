// Package logx writes one colored line per log entry. Safe for concurrent use.
package logx

import (
	"fmt"
	"hash/fnv"
	"io"
	"math"
	"os"
	"sync"
	"time"
)

type level int

const (
	levelInfo level = iota
	levelWarn
	levelError
)

const (
	colorReset = "\x1b[0m"
	// Pastel category colors: hue from the hash, fixed saturation and lightness.
	categorySaturation = 0.55
	categoryLightness  = 0.80
)

var levelNames = [...]string{"INFO", "WARN", "ERROR"}

var levelColors = [...]string{"\x1b[32m", "\x1b[33m", "\x1b[31m"}

// Logger writes entries to one writer. Colors are used only for terminals when NO_COLOR is unset.
type Logger struct {
	mu    sync.Mutex
	out   io.Writer
	color bool
}

// New returns a logger for out. Colors are on only when out is a terminal and NO_COLOR is unset.
func New(out *os.File) *Logger {
	return &Logger{out: out, color: isTerminal(out) && os.Getenv("NO_COLOR") == ""}
}

// Info logs an informational entry.
func (l *Logger) Info(category, format string, args ...any) {
	l.write(levelInfo, category, format, args...)
}

// Warn logs a warning entry.
func (l *Logger) Warn(category, format string, args ...any) {
	l.write(levelWarn, category, format, args...)
}

// Error logs an error entry.
func (l *Logger) Error(category, format string, args ...any) {
	l.write(levelError, category, format, args...)
}

func (l *Logger) write(lvl level, category, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	stamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")

	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.color {
		fmt.Fprintf(l.out, "%s %-5s [%s] %s\n", stamp, levelNames[lvl], category, msg)
		return
	}
	fmt.Fprintf(l.out, "%s %s%-5s%s %s[%s]%s %s\n",
		stamp, levelColors[lvl], levelNames[lvl], colorReset,
		categoryColor(category), category, colorReset, msg)
}

// categoryColor maps a category to the same pastel color on every run (FNV-1a hash to hue).
func categoryColor(category string) string {
	h := fnv.New32a()
	h.Write([]byte(category))
	hue := float64(h.Sum32() % 360)
	r, g, b := hslToRGB(hue, categorySaturation, categoryLightness)
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
}

func hslToRGB(hue, saturation, lightness float64) (uint8, uint8, uint8) {
	chroma := (1 - math.Abs(2*lightness-1)) * saturation
	x := chroma * (1 - math.Abs(math.Mod(hue/60, 2)-1))
	offset := lightness - chroma/2
	r, g, b := sectorRGB(int(hue/60)%6, chroma, x)
	return toByte(r + offset), toByte(g + offset), toByte(b + offset)
}

// sectorRGB returns the unshifted RGB values for one 60 degree sector of the hue wheel.
func sectorRGB(sector int, chroma, x float64) (float64, float64, float64) {
	switch sector {
	case 0:
		return chroma, x, 0
	case 1:
		return x, chroma, 0
	case 2:
		return 0, chroma, x
	case 3:
		return 0, x, chroma
	case 4:
		return x, 0, chroma
	default:
		return chroma, 0, x
	}
}

func toByte(v float64) uint8 {
	return uint8(math.Round(v * 255))
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
