package library

import (
	"regexp"
	"strings"
)

var (
	underscores    = regexp.MustCompile(`_+`)
	windowsWord    = regexp.MustCompile(`(?i)[-_ ]*windows?\b`)
	spaces         = regexp.MustCompile(`\s+`)
	trialSuffix    = regexp.MustCompile(`(?i)\s*\((?:patreon|trial|demo)\)\s*$`)
	platformSuffix = regexp.MustCompile(`(?i)[-_ ]*(?:win|x64|64bit|64-bit)$`)
)

// prettyTitle turns a folder name into a title: "Mad_Island_0.34" becomes "Mad Island 0.34".
func prettyTitle(folder string) string {
	title := underscores.ReplaceAllString(folder, " ")
	title = windowsWord.ReplaceAllString(title, " ")
	title = strings.TrimSpace(spaces.ReplaceAllString(title, " "))
	title = strings.TrimSpace(trialSuffix.ReplaceAllString(title, ""))
	title = strings.TrimSpace(platformSuffix.ReplaceAllString(title, ""))
	if title == "" {
		return folder
	}
	return title
}
