package logger

import (
	"fmt"
	"hash/fnv"
	"math"
	"os"
)

// Category colors are pastel: hue from the name, fixed saturation and high lightness.
const (
	pastelSaturation = 0.55
	pastelLightness  = 0.85
)

// colorEnabled reports whether lines written to out get ANSI colors.
// NO_COLOR turns colors off. FORCE_COLOR turns them on even when out is not a terminal.
func colorEnabled(out *os.File, lookup func(key string) (string, bool)) bool {
	if value, _ := lookup("NO_COLOR"); value != "" {
		return false
	}
	if value, _ := lookup("FORCE_COLOR"); value != "" && value != "0" {
		return true
	}
	return isTerminal(out)
}

// isTerminal reports whether file is a character device, such as a console. Pipes and files are not.
func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// categoryColor returns the same 24-bit color for the same category in every run.
func categoryColor(category string) string {
	hue := float64(fnvHash(category) % 360)
	r, g, b := hslToRGB(hue, pastelSaturation, pastelLightness)
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
}

func fnvHash(text string) uint32 {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(text))
	return hash.Sum32()
}

// hslToRGB converts hue in degrees, saturation and lightness in 0..1 to 8-bit channels.
func hslToRGB(hue, saturation, lightness float64) (int, int, int) {
	chroma := (1 - math.Abs(2*lightness-1)) * saturation
	x := chroma * (1 - math.Abs(math.Mod(hue/60, 2)-1))
	r, g, b := sectorRGB(hue, chroma, x)
	m := lightness - chroma/2
	return channel(r + m), channel(g + m), channel(b + m)
}

// sectorRGB gives the unshifted channels for the hue sector.
func sectorRGB(hue, chroma, x float64) (float64, float64, float64) {
	switch {
	case hue < 60:
		return chroma, x, 0
	case hue < 120:
		return x, chroma, 0
	case hue < 180:
		return 0, chroma, x
	case hue < 240:
		return 0, x, chroma
	case hue < 300:
		return x, 0, chroma
	default:
		return chroma, 0, x
	}
}

func channel(value float64) int {
	return int(math.Round(value * 255))
}
