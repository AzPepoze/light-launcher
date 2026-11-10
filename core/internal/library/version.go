package library

import (
	"regexp"
	"strconv"
	"strings"
)

// Fuzzy version handling. Sources are wildly inconsistent:
// v0.7, 0.7a, 1.0.2-hotfix, Ch.5, Ep.3, Version: 0.12.4, 2024-05-01.

// parsedVersion is a version string split into its parts.
type parsedVersion struct {
	raw    string
	kind   string // num, v, ch, ep or date
	nums   []int
	suffix string
}

var (
	datePattern      = regexp.MustCompile(`(\d{4})[-./](\d{1,2})[-./](\d{1,2})`)
	versionPrefix    = regexp.MustCompile(`(?i)^(?:ch(?:apter)?|ep(?:isode)?|v(?:er(?:sion)?)?|ver\.?|build|update|part)\s*[.:_-]?\s*`)
	digitRun         = regexp.MustCompile(`\d+`)
	trailingLetters  = regexp.MustCompile(`(?i)([a-z]+)\s*$`)
	wrapperChars     = regexp.MustCompile(`^[\[(<{]+|[)\]>}]+$`)
	versionCandidate = regexp.MustCompile(`[vV]?(\d{1,3}(?:\.\d{1,4}){1,3}[a-z]?)`)
)

// parseVersion returns nil when the input has no version in it.
func parseVersion(input string) *parsedVersion {
	raw := strings.TrimSpace(wrapperChars.ReplaceAllString(strings.TrimSpace(input), ""))
	if raw == "" {
		return nil
	}
	if date := datePattern.FindStringSubmatch(raw); date != nil {
		return &parsedVersion{raw: raw, kind: "date", nums: []int{atoi(date[1]), atoi(date[2]), atoi(date[3])}}
	}

	kind := "num"
	rest := raw
	if prefix := versionPrefix.FindString(raw); prefix != "" {
		kind = prefixKind(strings.ToLower(prefix))
		rest = raw[len(prefix):]
	}

	numbers := digitRun.FindAllString(rest, -1)
	if len(numbers) == 0 {
		return nil
	}
	nums := make([]int, len(numbers))
	for i, number := range numbers {
		nums[i] = atoi(number)
	}
	suffix := ""
	if match := trailingLetters.FindStringSubmatch(rest); match != nil {
		suffix = strings.ToLower(match[1])
	}
	return &parsedVersion{raw: raw, kind: kind, nums: nums, suffix: suffix}
}

func prefixKind(prefix string) string {
	switch {
	case strings.HasPrefix(prefix, "ch"):
		return "ch"
	case strings.HasPrefix(prefix, "ep"):
		return "ep"
	default:
		return "v"
	}
}

// compareVersions returns negative when a is older, 0 when equal, positive when a is newer.
func compareVersions(a, b string) int {
	pa := parseVersion(a)
	pb := parseVersion(b)
	if pa == nil && pb == nil {
		return 0
	}
	if pa == nil {
		return -1
	}
	if pb == nil {
		return 1
	}

	for i := 0; i < max(len(pa.nums), len(pb.nums)); i++ {
		x := numAt(pa.nums, i)
		y := numAt(pb.nums, i)
		if x != y {
			return x - y
		}
	}
	if pa.suffix != pb.suffix {
		if pa.suffix < pb.suffix {
			return -1
		}
		return 1
	}
	// chapters and episodes never compare meaningfully, so fall back to text
	if !isNumericKind(pa.kind) && !isNumericKind(pb.kind) && pa.kind != pb.kind {
		return strings.Compare(pa.raw, pb.raw)
	}
	return 0
}

func isNumericKind(kind string) bool {
	return kind == "num" || kind == "v"
}

func numAt(nums []int, i int) int {
	if i < len(nums) {
		return nums[i]
	}
	return 0
}

// isNewer reports whether candidate is a newer version than current. An empty current counts as older.
func isNewer(candidate, current string) bool {
	if candidate == "" {
		return false
	}
	if current == "" {
		return true
	}
	if strings.TrimSpace(candidate) == strings.TrimSpace(current) {
		return false
	}
	return compareVersions(candidate, current) > 0
}

// versionKey is a stable key for equality checks: 0.7, v0.7 and 0.7.0 share one key.
func versionKey(input string) string {
	p := parseVersion(input)
	if p == nil {
		return ""
	}
	nums := append([]int(nil), p.nums...)
	for len(nums) > 1 && nums[len(nums)-1] == 0 {
		nums = nums[:len(nums)-1]
	}
	parts := make([]string, len(nums))
	for i, n := range nums {
		parts[i] = strconv.Itoa(n)
	}
	prefix := "v"
	if p.kind == "ch" || p.kind == "ep" {
		prefix = p.kind
	}
	return prefix + strings.Join(parts, ".") + p.suffix
}

// sameVersion reports whether two version strings name the same release.
func sameVersion(a, b string) bool {
	ka := versionKey(a)
	kb := versionKey(b)
	if ka != "" && kb != "" {
		return ka == kb
	}
	return a == b
}

// versionFromText finds the last version-like token in a file or folder name. Years are ignored.
func versionFromText(input string) string {
	matches := versionCandidate.FindAllStringSubmatchIndex(input, -1)
	for i := len(matches) - 1; i >= 0; i-- {
		start, end := matches[i][0], matches[i][1]
		if isWordByte(input, start-1) || isWordByte(input, end) {
			continue
		}
		candidate := input[matches[i][2]:matches[i][3]]
		if yearPattern.MatchString(candidate) {
			return ""
		}
		return candidate
	}
	return ""
}

var yearPattern = regexp.MustCompile(`^(?:19|20)\d{2}$`)

// isWordByte reports whether the byte at i is an ASCII letter (lowercase) or digit. Out of range is false.
func isWordByte(s string, i int) bool {
	if i < 0 || i >= len(s) {
		return false
	}
	c := s[i]
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z')
}

func atoi(text string) int {
	value, _ := strconv.Atoi(text)
	return value
}
