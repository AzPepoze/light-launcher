package library

import (
	"regexp"
	"sort"
	"strings"
)

// Matches a downloaded file name to a library game.

const minTitleLength = 4

// AutoMatchScore is the score a match needs to be imported without asking.
const AutoMatchScore = 0.9

// SuggestScore is the lowest score worth showing as a suggestion.
const SuggestScore = 0.5

// releaseNoise are release tags that appear in file names but not in game titles.
var releaseNoise = map[string]bool{
	"repack": true, "gog": true, "multi": true, "multiplayer": true, "uncensored": true, "censored": true,
	"compressed": true, "win64": true, "x64": true, "x86": true, "64bit": true, "32bit": true, "cracked": true,
	"portable": true, "installer": true, "cheat": true, "walkthrough": true, "guide": true, "mod": true,
	"update": true, "patch": true, "dlc": true, "bonus": true, "zip": true, "rar": true, "7z": true, "tar": true,
	"gz": true, "tgz": true, "xz": true, "bz2": true, "iso": true, "exe": true, "apk": true, "dmg": true,
	"msi": true, "bin": true, "jpg": true, "jpeg": true, "png": true, "webp": true, "gif": true, "mp4": true,
	"webm": true, "mp3": true, "ogg": true, "wav": true, "pdf": true, "txt": true, "md": true, "html": true,
}

var (
	bracketedPattern   = regexp.MustCompile(`\[[^\]]*\]|\([^)]*\)`)
	nonAlnumLowerRegex = regexp.MustCompile(`[^a-z0-9]`)
	namedVersion       = regexp.MustCompile(`(?i)(?:v|ver|version)[._\s-]?\d{1,4}(?:[._\s-]\d{1,4}){0,3}[a-z]?`)
	numericVersion     = regexp.MustCompile(`(?i)\d{1,3}(?:[._]\d{1,4}){1,3}[a-z]?`)
	riskyFileName      = regexp.MustCompile(`(?i)\b(?:demo|patch|update|update-only|dlc|soundtrack|addon|expansion|cheat|walkthrough|guide|mod|repack|compressed|torrent)\b`)
)

// normalizeName is the old strict form: lowercase, no brackets, letters and digits only.
func normalizeName(input string) string {
	lower := strings.ToLower(bracketedPattern.ReplaceAllString(input, ""))
	return nonAlnumLowerRegex.ReplaceAllString(lower, "")
}

// stripVersion removes version-like fragments such as v0129, 0.16 or 1.2.3 before tokenising.
func stripVersion(input string) string {
	text := replaceUnlessAdjacent(namedVersion, input, isVersionBoundary)
	return replaceUnlessAdjacent(numericVersion, text, isVersionBoundary)
}

// replaceUnlessAdjacent blanks each match whose neighbours are not word characters.
// Go's regexp has no lookarounds, so the boundary check is done by hand.
func replaceUnlessAdjacent(re *regexp.Regexp, input string, isWord func(byte) bool) string {
	var b strings.Builder
	last := 0
	for _, loc := range re.FindAllStringIndex(input, -1) {
		start, end := loc[0], loc[1]
		if (start > 0 && isWord(input[start-1])) || (end < len(input) && isWord(input[end])) {
			continue
		}
		b.WriteString(input[last:start])
		b.WriteByte(' ')
		last = end
	}
	b.WriteString(input[last:])
	return b.String()
}

func isVersionBoundary(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// wordTokens are the matchable words of a file name or title, versions and release noise removed.
func wordTokens(input string) []string {
	var out []string
	for _, token := range matchTokens(stripVersion(input)) {
		if !releaseNoise[token] {
			out = append(out, token)
		}
	}
	return out
}

// editDistance is the Levenshtein distance between two strings.
func editDistance(a, b string) int {
	if a == b {
		return 0
	}
	if abs(len(a)-len(b)) > 1 {
		return 2
	}
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		next := make([]int, len(b)+1)
		next[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			next[j] = min(prev[j]+1, next[j-1]+1, prev[j-1]+cost)
		}
		prev = next
	}
	return prev[len(b)]
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

// tokenClose reports whether two tokens are the same word, allowing a typo or a prefix.
func tokenClose(a, b string) bool {
	if a == b {
		return true
	}
	// numbers are sequels, not typos
	if isDigits(a) || isDigits(b) {
		return false
	}
	if len(a) < 4 || len(b) < 4 {
		return false
	}
	if editDistance(a, b) <= 1 {
		return true
	}
	return strings.HasPrefix(a, b) || strings.HasPrefix(b, a)
}

func isDigits(text string) bool {
	if text == "" {
		return false
	}
	for i := 0; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return false
		}
	}
	return true
}

// isPrefixMatch reports whether the file tokens match the first title tokens in order.
func isPrefixMatch(fileWords, titleWordList []string) bool {
	if len(fileWords) == 0 || len(fileWords) > len(titleWordList) {
		return false
	}
	for i, token := range fileWords {
		if !tokenClose(token, titleWordList[i]) {
			return false
		}
	}
	return true
}

func compact(input string) string {
	return nonAlnumLowerRegex.ReplaceAllString(compactTitle(stripVersion(input)), "")
}

// countClose counts the words in from that have a close word in against.
func countClose(from, against []string) int {
	count := 0
	for _, token := range from {
		for _, other := range against {
			if tokenClose(token, other) {
				count++
				break
			}
		}
	}
	return count
}

// scoreFileName rates how well a file name matches a game title, 0 to 1.
func scoreFileName(fileName, gameTitle string) float64 {
	fileWords := wordTokens(fileName)
	titleWordList := wordTokens(gameTitle)
	if len(fileWords) == 0 || len(titleWordList) == 0 {
		return 0
	}

	matchedFile := countClose(fileWords, titleWordList)
	matchedTitle := countClose(titleWordList, fileWords)
	score := 0.65*float64(matchedFile)/float64(len(fileWords)) + 0.25*float64(matchedTitle)/float64(len(titleWordList))

	if isPrefixMatch(fileWords, titleWordList) {
		score += 0.2
	}
	fileCompact := compact(fileName)
	titleCompact := compact(gameTitle)
	if fileCompact != "" && titleCompact != "" &&
		(strings.Contains(fileCompact, titleCompact) || strings.Contains(titleCompact, fileCompact)) {
		score += 0.1
	}

	// one weak token is not enough unless the game title is that one token
	if matchedFile < min(2, len(titleWordList)) {
		score *= 0.5
	}
	if len(titleCompact) < minTitleLength {
		score *= 0.5
	}
	// patches, demos and mods are easy to mistake for the game; always ask first
	if riskyFileName.MatchString(fileName) {
		score = min(score, AutoMatchScore-0.01)
	}
	return min(score, 1)
}

// TitleScore is one title ranked against a file name. Index points into the titles slice.
type TitleScore struct {
	Index int
	Score float64
}

// rankTitles returns the best titles for a file name, highest score first.
func rankTitles(fileName string, titles []string, limit int) []TitleScore {
	if len(wordTokens(fileName)) == 0 {
		return nil
	}
	var ranked []TitleScore
	for i, title := range titles {
		score := scoreFileName(fileName, title)
		if score >= SuggestScore {
			ranked = append(ranked, TitleScore{Index: i, Score: score})
		}
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].Score != ranked[j].Score {
			return ranked[i].Score > ranked[j].Score
		}
		return len(titles[ranked[i].Index]) > len(titles[ranked[j].Index])
	})
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}
	return ranked
}

// MatchResult is the best title for a file, with the lead it has over the runner-up.
type MatchResult struct {
	Index  int
	Score  float64
	Margin float64
}

// bestConfidentMatch returns the single obvious match, or false when a person should choose.
func bestConfidentMatch(fileName string, titles []string) (MatchResult, bool) {
	ranked := rankTitles(fileName, titles, 2)
	if len(ranked) == 0 || ranked[0].Score < AutoMatchScore {
		return MatchResult{}, false
	}
	if len(ranked) > 1 && ranked[0].Score-ranked[1].Score < 0.12 {
		return MatchResult{}, false
	}
	return MatchResult{Index: ranked[0].Index, Score: ranked[0].Score}, true
}
