package plugins

import (
	"fmt"
	"strconv"
	"strings"
)

// parseSemver reads the major, minor and patch numbers. Pre-release and build parts are ignored.
func parseSemver(version string) ([3]int, error) {
	core, _, _ := strings.Cut(version, "+")
	core, _, _ = strings.Cut(core, "-")
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return [3]int{}, fmt.Errorf("invalid version %q", version)
	}
	var numbers [3]int
	for i, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return [3]int{}, fmt.Errorf("invalid version %q", version)
		}
		numbers[i] = number
	}
	return numbers, nil
}

// compareSemver returns -1, 0 or 1 when a is lower, equal or higher than b.
func compareSemver(a, b string) (int, error) {
	left, err := parseSemver(a)
	if err != nil {
		return 0, err
	}
	right, err := parseSemver(b)
	if err != nil {
		return 0, err
	}
	for i := range left {
		if left[i] < right[i] {
			return -1, nil
		}
		if left[i] > right[i] {
			return 1, nil
		}
	}
	return 0, nil
}
