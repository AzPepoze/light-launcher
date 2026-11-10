package library

import (
	"html"
	"regexp"
	"strings"
)

var (
	scriptBlock    = regexp.MustCompile(`(?is)<(script|style)[^>]*>.*?</(script|style)>`)
	lineBreakTag   = regexp.MustCompile(`(?i)<br\s*/?>`)
	paragraphEnd   = regexp.MustCompile(`(?i)</p>`)
	anyTag         = regexp.MustCompile(`<[^>]*>`)
	extraBlankLine = regexp.MustCompile(`\n{3,}`)
)

// plainText turns description HTML into plain text. Scripts and styles go, tags become line breaks or nothing,
// and entities are decoded. It returns nil when nothing is left.
func plainText(markup *string) *string {
	if markup == nil {
		return nil
	}
	text := scriptBlock.ReplaceAllString(*markup, "")
	text = lineBreakTag.ReplaceAllString(text, "\n")
	text = paragraphEnd.ReplaceAllString(text, "\n\n")
	text = anyTag.ReplaceAllString(text, "")
	text = strings.TrimSpace(extraBlankLine.ReplaceAllString(html.UnescapeString(text), "\n\n"))
	if text == "" {
		return nil
	}
	return &text
}
