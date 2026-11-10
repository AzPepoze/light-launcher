package library

import (
	"regexp"
	"strings"
)

// Shared title normalisation and scoring, used by the search index and the local matcher.

var titleNoise = map[string]bool{
	"the": true, "a": true, "an": true, "and": true, "of": true, "for": true, "my": true, "to": true,
	"in": true, "on": true, "public": true, "trial": true, "demo": true, "win": true, "windows": true,
	"pc": true, "ver": true, "version": true, "eng": true, "english": true, "final": true, "build": true,
	"update": true, "game": true, "games": true, "completed": true, "abandoned": true, "ongoing": true,
	"renpy": true, "unity": true, "rpgm": true, "html": true,
}

var (
	bracketPattern     = regexp.MustCompile(`\[[^\]]*\]`)
	parenPattern       = regexp.MustCompile(`\([^)]*\)`)
	quotePattern       = regexp.MustCompile("['’`]")
	camelPattern       = regexp.MustCompile(`([a-z])([A-Z])`)
	separatorPattern   = regexp.MustCompile(`[_\-.]+`)
	nonAlnumPattern    = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	versionWordPattern = regexp.MustCompile(`(?i)\b(?:v|ver|version)\s*\d+(?:\s*\d+)*[a-z]?\b`)
	spacesPattern      = regexp.MustCompile(`\s+`)
	compactNonAlnum    = regexp.MustCompile(`[^a-z0-9]`)
)

// normalizeTitle lowercases a title and strips brackets, versions and punctuation.
func normalizeTitle(input string) string {
	text := bracketPattern.ReplaceAllString(input, " ")
	text = parenPattern.ReplaceAllString(text, " ")
	text = quotePattern.ReplaceAllString(text, "")
	text = camelPattern.ReplaceAllString(text, "$1 $2")
	text = separatorPattern.ReplaceAllString(text, " ")
	text = nonAlnumPattern.ReplaceAllString(text, " ")
	text = versionWordPattern.ReplaceAllString(text, " ")
	text = spacesPattern.ReplaceAllString(text, " ")
	return strings.ToLower(strings.TrimSpace(text))
}

// compactTitle is the normalised title with spaces removed.
func compactTitle(input string) string {
	return compactNonAlnum.ReplaceAllString(normalizeTitle(input), "")
}

// titleTokens returns the distinct words of a title, without noise words and single letters.
func titleTokens(input string) []string {
	return distinctTokens(normalizeTitle(input), func(token string) bool {
		return len(token) > 1 && !titleNoise[token]
	})
}

// matchTokens is like titleTokens but keeps one-character tokens, so "Game 2" never matches "Game 3".
func matchTokens(input string) []string {
	return distinctTokens(normalizeTitle(input), func(token string) bool {
		return token != "" && !titleNoise[token]
	})
}

func distinctTokens(text string, keep func(string) bool) []string {
	seen := map[string]bool{}
	var out []string
	for _, token := range strings.Split(text, " ") {
		if keep(token) && !seen[token] {
			seen[token] = true
			out = append(out, token)
		}
	}
	return out
}

// scoreMatch rates how well a candidate title matches a query. Higher is better, about 0 to 2.5.
func scoreMatch(query, candidate string) float64 {
	q := normalizeTitle(query)
	c := normalizeTitle(candidate)
	if q == "" || c == "" {
		return 0
	}

	qTokens := titleTokens(query)
	cTokens := map[string]bool{}
	for _, token := range titleTokens(candidate) {
		cTokens[token] = true
	}
	hits := 0
	for _, token := range qTokens {
		if cTokens[token] {
			hits++
		}
	}

	score := 0.0
	if len(qTokens) > 0 {
		score = float64(hits) / float64(len(qTokens))
	}
	qCompact := compactTitle(query)
	cCompact := compactTitle(candidate)
	if len(qCompact) > 5 && strings.Contains(cCompact, qCompact) {
		score += 0.9
	}
	if strings.HasPrefix(c, q) && len(q) > 4 {
		score += 0.4
	}
	// A compact hit already proves the title is there, so don't penalise it for
	// having fewer separate token matches (Kel Pie Rays vs KelpieRays).
	if hits < min(2, len(qTokens)) && score < 0.9 {
		score *= 0.45
	}
	return min(score, 2.5)
}
