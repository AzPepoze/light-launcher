// Package semver parses and orders the three-number versions the launcher and its plugins use.
package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is a semantic version with an optional pre-release part. Build metadata is ignored.
type Version struct {
	numbers [3]int
	pre     []string
}

// Parse reads a version such as 1.2.3, v1.2.3 or 1.2.3-beta.1+build.
func Parse(raw string) (Version, error) {
	text := strings.TrimPrefix(raw, "v")
	core, _, _ := strings.Cut(text, "+")
	core, pre, hasPre := strings.Cut(core, "-")
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version %q", raw)
	}
	var version Version
	for i, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 || (len(part) > 1 && part[0] == '0') {
			return Version{}, fmt.Errorf("invalid version %q", raw)
		}
		version.numbers[i] = number
	}
	if hasPre {
		if pre == "" {
			return Version{}, fmt.Errorf("invalid version %q", raw)
		}
		version.pre = strings.Split(pre, ".")
		for _, identifier := range version.pre {
			if identifier == "" || !validIdentifier(identifier) {
				return Version{}, fmt.Errorf("invalid version %q", raw)
			}
		}
	}
	return version, nil
}

func validIdentifier(identifier string) bool {
	for _, char := range identifier {
		if (char < '0' || char > '9') && (char < 'A' || char > 'Z') && (char < 'a' || char > 'z') && char != '-' {
			return false
		}
	}
	return true
}

// Compare returns -1, 0 or 1 when a is lower, equal or higher than b.
func Compare(a, b Version) int {
	for i := range a.numbers {
		if a.numbers[i] != b.numbers[i] {
			return order(a.numbers[i] < b.numbers[i])
		}
	}
	return comparePre(a.pre, b.pre)
}

// CompareStrings parses both versions and compares them.
func CompareStrings(a, b string) (int, error) {
	left, err := Parse(a)
	if err != nil {
		return 0, err
	}
	right, err := Parse(b)
	if err != nil {
		return 0, err
	}
	return Compare(left, right), nil
}

func comparePre(left, right []string) int {
	if len(left) == 0 && len(right) == 0 {
		return 0
	}
	if len(left) == 0 {
		return 1
	}
	if len(right) == 0 {
		return -1
	}
	for i := 0; i < len(left) && i < len(right); i++ {
		if order := compareIdentifier(left[i], right[i]); order != 0 {
			return order
		}
	}
	switch {
	case len(left) > len(right):
		return 1
	case len(left) < len(right):
		return -1
	default:
		return 0
	}
}

func compareIdentifier(left, right string) int {
	leftNumber, leftErr := strconv.Atoi(left)
	rightNumber, rightErr := strconv.Atoi(right)
	switch {
	case leftErr == nil && rightErr == nil:
		if leftNumber == rightNumber {
			return 0
		}
		return order(leftNumber < rightNumber)
	case leftErr == nil:
		return -1
	case rightErr == nil:
		return 1
	default:
		return strings.Compare(left, right)
	}
}

func order(less bool) int {
	if less {
		return -1
	}
	return 1
}
