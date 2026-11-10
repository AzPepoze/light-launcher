package appsvc

import (
	"regexp"
	"strings"
)

// jsRegexSpecials are escaped before "*" becomes ".*", matching the Electron pattern compiler.
const jsRegexSpecials = `.+^${}()|[]\`

// compileWildcardPatterns turns "*.exe" or "unins*" into case-insensitive whole-name regexps.
// Names without "*" or "." match by prefix. Invalid patterns are skipped.
func compileWildcardPatterns(patterns []string) []*regexp.Regexp {
	regexes := []*regexp.Regexp{}
	for _, raw := range patterns {
		pattern := strings.TrimSpace(raw)
		if pattern == "" {
			continue
		}
		regex, err := regexp.Compile("(?i)^" + wildcardToRegex(prefixPattern(pattern)) + "$")
		if err != nil {
			continue
		}
		regexes = append(regexes, regex)
	}
	return regexes
}

func prefixPattern(pattern string) string {
	if !strings.Contains(pattern, "*") && !strings.Contains(pattern, ".") {
		return pattern + "*"
	}
	return pattern
}

func wildcardToRegex(pattern string) string {
	var builder strings.Builder
	for _, char := range pattern {
		switch {
		case char == '*':
			builder.WriteString(".*")
		case strings.ContainsRune(jsRegexSpecials, char):
			builder.WriteRune('\\')
			builder.WriteRune(char)
		default:
			builder.WriteRune(char)
		}
	}
	return builder.String()
}
