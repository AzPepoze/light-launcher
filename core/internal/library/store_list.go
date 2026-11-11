package library

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// Filter is the payload of LibraryList. Zero values mean "no filter".
type Filter struct {
	Search    string   `json:"search"`
	Source    string   `json:"source"`
	Tags      []string `json:"tags"`
	Status    []string `json:"status"`
	Favorite  bool     `json:"favorite"`
	Installed *bool    `json:"installed"`
	Updates   bool     `json:"updates"`
	Hidden    bool     `json:"hidden"`
	Sort      string   `json:"sort"`
	Order     string   `json:"order"`
	Limit     int      `json:"limit"`
	Offset    int      `json:"offset"`
	// WantTotal asks for the total on any page. The first page always has it.
	WantTotal bool `json:"wantTotal"`
}

// Page is one page of games. Total counts every game that matches the filter. It is nil on later pages
// unless the filter asked for it, because counting costs as much as the rest of the query.
type Page struct {
	Items []Game `json:"items"`
	Total *int   `json:"total,omitempty"`
}

// TagCount is a tag and how many visible games use it.
type TagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// Stats is the summary shown in the library header.
type Stats struct {
	Total     int64 `json:"total"`
	Installed int64 `json:"installed"`
	Updates   int64 `json:"updates"`
	Bytes     int64 `json:"bytes"`
	Playtime  int64 `json:"playtime"`
}

const (
	defaultLimit = 200
	maxLimit     = 1000
	// trigramMinLength is the shortest search the trigram index can match.
	trigramMinLength = 3
)

// sortColumns maps the sort names the UI sends to SQL expressions.
var sortColumns = map[string]string{
	"added":      "added_at",
	"updated":    "coalesce(source_updated_at, updated_at)",
	"checked":    "coalesce(last_checked_at, added_at)",
	"title":      "title collate nocase",
	"rating":     "coalesce(rating, 0)",
	"size":       "coalesce(install_size, 0)",
	"playtime":   "playtime_s",
	"lastPlayed": "coalesce(last_played_at, '')",
}

// List returns one page of games that match the filter.
func (s *Store) List(ctx context.Context, f Filter) (Page, error) {
	where, args := buildWhere(f)
	limit := clamp(f.Limit, 1, maxLimit, defaultLimit)
	offset := max(f.Offset, 0)
	var page Page
	if offset == 0 || f.WantTotal {
		total, err := s.countGames(ctx, where, args)
		if err != nil {
			return Page{}, err
		}
		page.Total = &total
	}
	items, err := s.queryMany(ctx, pageQuery(where, f), append(args, limit, offset)...)
	if err != nil {
		return Page{}, err
	}
	page.Items = items
	return page, nil
}

func (s *Store) countGames(ctx context.Context, where string, args []any) (int, error) {
	var total int
	err := s.db.QueryRowContext(ctx, "select count(*) from games"+where, args...).Scan(&total)
	return total, err
}

// pageQuery selects one page of games. Its arguments are the where arguments, then limit and offset.
func pageQuery(where string, f Filter) string {
	return fmt.Sprintf("select %s from games%s order by %s limit ? offset ?", gameColumns, where, orderBy(f))
}

// orderBy returns the ORDER BY terms. The id breaks ties, so pages do not repeat or skip games.
func orderBy(f Filter) string {
	direction := sortDirection(f.Order)
	return sortExpression(f.Sort) + " " + direction + ", id " + direction
}

// TagCounts counts tags across visible games, most used first.
func (s *Store) TagCounts(ctx context.Context) ([]TagCount, error) {
	rows, err := s.db.QueryContext(ctx, `select tags from games where hidden = 0`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := map[string]int{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		for _, tag := range decodeStrings(raw) {
			counts[tag]++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]TagCount, 0, len(counts))
	for tag, count := range counts {
		out = append(out, TagCount{Tag: tag, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Tag < out[j].Tag
	})
	return out, nil
}

// Stats sums the visible games.
func (s *Store) Stats(ctx context.Context) (Stats, error) {
	var st Stats
	err := s.db.QueryRowContext(ctx, `select count(*),
		coalesce(sum(case when install_dir is not null then 1 else 0 end), 0),
		coalesce(sum(case when installed_version is not null and version is not null and version <> installed_version then 1 else 0 end), 0),
		coalesce(sum(coalesce(install_size, 0)), 0),
		coalesce(sum(playtime_s), 0)
		from games where hidden = 0`).Scan(&st.Total, &st.Installed, &st.Updates, &st.Bytes, &st.Playtime)
	return st, err
}

// buildWhere turns a filter into a WHERE clause and its arguments.
func buildWhere(f Filter) (string, []any) {
	var clauses []string
	var args []any
	if !f.Hidden {
		clauses = append(clauses, "hidden = 0")
	}
	if search := strings.TrimSpace(f.Search); search != "" {
		clause, searchArgs := searchClause(search)
		clauses = append(clauses, clause)
		args = append(args, searchArgs...)
	}
	if f.Source != "" {
		clauses = append(clauses, "source = ?")
		args = append(args, f.Source)
	}
	if len(f.Status) > 0 {
		clauses = append(clauses, "status in ("+placeholders(len(f.Status))+")")
		for _, status := range f.Status {
			args = append(args, status)
		}
	}
	if f.Favorite {
		clauses = append(clauses, "favorite = 1")
	}
	if f.Installed != nil {
		clauses = append(clauses, installedClause(*f.Installed))
	}
	if f.Updates {
		clauses = append(clauses, "installed_version is not null and version is not null and version <> installed_version")
	}
	for _, tag := range f.Tags {
		clauses = append(clauses, "tags like ?")
		args = append(args, fmt.Sprintf(`%%"%s"%%`, tag))
	}
	if len(clauses) == 0 {
		return "", nil
	}
	return " where " + strings.Join(clauses, " and "), args
}

// searchClause matches text inside the title, developer and tags. Three or more characters use the
// full-text index; one or two characters match the start of the title.
func searchClause(text string) (string, []any) {
	if utf8.RuneCountInString(text) >= trigramMinLength {
		return "id in (select rowid from games_fts where games_fts match ?)", []any{ftsPhrase(text)}
	}
	return `title like ? escape '\'`, []any{escapeLike(text) + "%"}
}

// ftsPhrase quotes text as one FTS5 phrase, so quotes and operators in it are plain text.
func ftsPhrase(text string) string {
	return `"` + strings.ReplaceAll(text, `"`, `""`) + `"`
}

// escapeLike escapes the LIKE wildcards and the escape character.
func escapeLike(text string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(text)
}

func installedClause(installed bool) string {
	if installed {
		return "install_dir is not null"
	}
	return "install_dir is null"
}

func sortExpression(name string) string {
	if column, ok := sortColumns[name]; ok {
		return column
	}
	return sortColumns["added"]
}

func sortDirection(order string) string {
	if strings.EqualFold(order, "asc") {
		return "ASC"
	}
	return "DESC"
}

func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

func clamp(value, low, high, fallback int) int {
	if value == 0 {
		return fallback
	}
	return min(max(value, low), high)
}
