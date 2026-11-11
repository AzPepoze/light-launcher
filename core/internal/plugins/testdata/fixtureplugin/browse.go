package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
)

const (
	browseMaxLimit  = 100
	filtersMaxLimit = 200
	sortTitle       = "title"
	tagsGroup       = "tags"
)

// catalogGame is one game of the fixed catalog that source.browse and source.listFilters read.
type catalogGame struct {
	ID    string
	Title string
	Tags  []string
}

var browseCatalog = []catalogGame{
	{"b1", "Alpha Quest", []string{"rpg", "fantasy"}},
	{"b2", "Blue Harbor", []string{"visual-novel"}},
	{"b3", "Cinder Keep", []string{"rpg", "horror"}},
	{"b4", "Delta Puzzle", []string{"puzzle"}},
	{"b5", "Emerald Road", []string{"rpg", "visual-novel"}},
	{"b6", "Foggy Lantern", []string{"horror", "puzzle"}},
	{"b7", "Glass Orchard", []string{"fantasy"}},
	{"b8", "Harbor Lights", []string{"visual-novel", "puzzle"}},
}

// browseError is an error answer with its JSON-RPC code and optional data.
type browseError struct {
	code    int
	message string
	data    any
}

func (e browseError) Error() string { return e.message }

var errInvalidCursor = browseError{
	code:    codeInvalidParams,
	message: "invalid cursor",
	data:    map[string]string{"code": "invalid_cursor"},
}

func invalidBrowse(format string, args ...any) error {
	return browseError{code: codeInvalidParams, message: fmt.Sprintf(format, args...)}
}

type browseParams struct {
	Source  string         `json:"source"`
	Query   string         `json:"query"`
	Filters map[string]any `json:"filters"`
	Sort    string         `json:"sort"`
	Order   string         `json:"order"`
	Cursor  string         `json:"cursor"`
	Limit   int            `json:"limit"`
}

type browseResult struct {
	Games      []fixtureGame `json:"games"`
	NextCursor string        `json:"nextCursor,omitempty"`
	Total      int           `json:"total"`
}

// browseShape is the part of a browse request that a cursor is tied to.
type browseShape struct {
	Query   string         `json:"query"`
	Filters map[string]any `json:"filters"`
	Sort    string         `json:"sort"`
	Order   string         `json:"order"`
}

// cursorState is what a cursor holds: the offset of the next game and the request it belongs to.
type cursorState struct {
	Offset  int    `json:"offset"`
	Request string `json:"request"`
}

type filtersParams struct {
	Source string `json:"source"`
	Group  string `json:"group"`
	Query  string `json:"query"`
	Limit  int    `json:"limit"`
}

type filterOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type filtersResult struct {
	Options []filterOption `json:"options"`
}

// answerBrowse writes the source.browse answer.
func answerBrowse(msg message) {
	result, err := browse(msg.Params)
	respondAnswer(msg.ID, result, err)
}

// answerListFilters writes the source.listFilters answer.
func answerListFilters(msg message) {
	result, err := listFilters(msg.Params)
	respondAnswer(msg.ID, result, err)
}

// respondAnswer sends the result, or the error with its code and data.
func respondAnswer(id json.RawMessage, result any, err error) {
	if err == nil {
		respond(id, result)
		return
	}
	var failure browseError
	if errors.As(err, &failure) {
		respondErrorData(id, failure.code, failure.message, failure.data)
		return
	}
	respondError(id, codeInvalidParams, err.Error())
}

// browse returns one page of the catalog games that match the query and filters. Unknown filter groups are ignored.
func browse(raw json.RawMessage) (browseResult, error) {
	var params browseParams
	if err := json.Unmarshal(raw, &params); err != nil {
		return browseResult{}, invalidBrowse("params: %v", err)
	}
	if err := checkBrowseParams(params); err != nil {
		return browseResult{}, err
	}
	key, err := browseKey(params)
	if err != nil {
		return browseResult{}, err
	}
	offset, err := decodeCursor(params.Cursor, key)
	if err != nil {
		return browseResult{}, err
	}
	matches := browseMatches(params)
	if offset > len(matches) {
		return browseResult{}, errInvalidCursor
	}
	end := min(offset+params.Limit, len(matches))
	result := browseResult{Games: make([]fixtureGame, 0, end-offset), Total: len(matches)}
	for _, game := range matches[offset:end] {
		result.Games = append(result.Games, catalogFixtureGame(game))
	}
	if end < len(matches) {
		result.NextCursor = encodeCursor(end, key)
	}
	return result, nil
}

func checkBrowseParams(params browseParams) error {
	if params.Source != fixtureSource {
		return invalidBrowse("unknown source %q", params.Source)
	}
	if params.Limit < 1 || params.Limit > browseMaxLimit {
		return invalidBrowse("limit must be 1 to %d, not %d", browseMaxLimit, params.Limit)
	}
	if params.Order != "" && params.Order != "asc" && params.Order != "desc" {
		return invalidBrowse("order must be asc or desc, not %q", params.Order)
	}
	if params.Sort != "" && params.Sort != sortTitle {
		return invalidBrowse("unknown sort %q", params.Sort)
	}
	return nil
}

// browseMatches returns the catalog games that pass the query and tag filter, sorted when a sort is set.
func browseMatches(params browseParams) []catalogGame {
	var matches []catalogGame
	for _, game := range browseCatalog {
		if containsText(game.Title, params.Query) && matchesTags(game.Tags, params.Filters[tagsGroup]) {
			matches = append(matches, game)
		}
	}
	if params.Sort == sortTitle {
		sortByTitle(matches, params.Order)
	}
	return matches
}

func sortByTitle(games []catalogGame, order string) {
	slices.SortFunc(games, func(a, b catalogGame) int { return strings.Compare(a.Title, b.Title) })
	if order == "desc" {
		slices.Reverse(games)
	}
}

// matchesTags keeps a game that has at least one of the selected tags. No selection keeps every game.
func matchesTags(gameTags []string, filter any) bool {
	selected := selectedTags(filter)
	if len(selected) == 0 {
		return true
	}
	for _, tag := range selected {
		if slices.Contains(gameTags, tag) {
			return true
		}
	}
	return false
}

// selectedTags reads a multi filter, which is a string or a list of strings.
func selectedTags(filter any) []string {
	switch value := filter.(type) {
	case string:
		if value != "" {
			return []string{value}
		}
	case []any:
		var tags []string
		for _, item := range value {
			if tag, ok := item.(string); ok {
				tags = append(tags, tag)
			}
		}
		return tags
	}
	return nil
}

// containsText is a case-insensitive substring test. An empty query matches everything.
func containsText(text, query string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(query))
}

// browseKey is the request shape as JSON. Cursors are valid only for the same key.
func browseKey(params browseParams) (string, error) {
	raw, err := json.Marshal(browseShape{Query: params.Query, Filters: params.Filters, Sort: params.Sort, Order: params.Order})
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func encodeCursor(offset int, key string) string {
	raw, _ := json.Marshal(cursorState{Offset: offset, Request: key})
	return base64.RawURLEncoding.EncodeToString(raw)
}

// decodeCursor returns the offset in a cursor. An empty cursor is the first page.
func decodeCursor(cursor, key string) (int, error) {
	if cursor == "" {
		return 0, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return 0, errInvalidCursor
	}
	var state cursorState
	if err := json.Unmarshal(raw, &state); err != nil || state.Request != key || state.Offset < 1 {
		return 0, errInvalidCursor
	}
	return state.Offset, nil
}

// listFilters counts the tags of the whole catalog and returns the ones that match the query.
func listFilters(raw json.RawMessage) (filtersResult, error) {
	var params filtersParams
	if err := json.Unmarshal(raw, &params); err != nil {
		return filtersResult{}, invalidBrowse("params: %v", err)
	}
	if params.Source != fixtureSource {
		return filtersResult{}, invalidBrowse("unknown source %q", params.Source)
	}
	if params.Group != tagsGroup {
		return filtersResult{}, invalidBrowse("unknown group %q", params.Group)
	}
	if params.Limit < 1 || params.Limit > filtersMaxLimit {
		return filtersResult{}, invalidBrowse("limit must be 1 to %d, not %d", filtersMaxLimit, params.Limit)
	}
	counts := tagCounts()
	tags := make([]string, 0, len(counts))
	for tag := range counts {
		tags = append(tags, tag)
	}
	slices.Sort(tags)
	result := filtersResult{Options: []filterOption{}}
	for _, tag := range tags {
		if len(result.Options) == params.Limit {
			break
		}
		if containsText(tag, params.Query) {
			result.Options = append(result.Options, filterOption{Value: tag, Label: tag, Count: counts[tag]})
		}
	}
	return result, nil
}

// tagCounts counts how many catalog games have each tag.
func tagCounts() map[string]int {
	counts := map[string]int{}
	for _, game := range browseCatalog {
		for _, tag := range game.Tags {
			counts[tag]++
		}
	}
	return counts
}

func catalogFixtureGame(game catalogGame) fixtureGame {
	return fixtureGame{
		Source:   fixtureSource,
		SourceID: game.ID,
		URL:      "https://example.test/g/" + game.ID,
		Title:    game.Title,
	}
}
