package runtimesvc

import (
	"regexp"
	"strconv"
	"strings"
)

// The Electron app used JavaScript parseInt and parseFloat, which read a leading number and ignore the rest.
var (
	leadingIntPattern   = regexp.MustCompile(`^\s*[+-]?\d+`)
	leadingFloatPattern = regexp.MustCompile(`^\s*[+-]?(\d+\.?\d*|\.\d+)([eE][+-]?\d+)?`)
)

// parseIntLike mirrors parseInt(value, 10). ok is false when no number starts the value.
func parseIntLike(value string) (int, bool) {
	match := leadingIntPattern.FindString(value)
	if match == "" {
		return 0, false
	}
	number, err := strconv.Atoi(strings.TrimSpace(match))
	return number, err == nil
}

// parseFloatLike mirrors parseFloat(value). ok is false when no number starts the value.
func parseFloatLike(value string) (float64, bool) {
	match := leadingFloatPattern.FindString(value)
	if match == "" {
		return 0, false
	}
	number, err := strconv.ParseFloat(strings.TrimSpace(match), 64)
	return number, err == nil
}

// intOrDefault returns fallback when value is 0, the same as JavaScript value || fallback.
func intOrDefault(value int, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}

// flowScaleValue parses an LSFG flow scale. Missing or zero values become 1.0.
func flowScaleValue(value string) float64 {
	number, ok := parseFloatLike(value)
	if !ok || number == 0 {
		return 1.0
	}
	return number
}
