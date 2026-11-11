package contract

import (
	"errors"
	"time"
)

// Time limits for the browse calls. A browse page can be slow on a busy site, a list of filter values should be quick.
const (
	BrowseTimeout      = 20 * time.Second
	ListFiltersTimeout = 10 * time.Second
)

// ErrInvalidCursor means the source does not know the page cursor, usually because the filters or sort changed.
var ErrInvalidCursor = errors.New("the list changed; start again from the first page")

// BrowseRequest asks a source for one page of its games, filtered and sorted. It never saves games.
type BrowseRequest struct {
	Source  string         `json:"source"`
	Query   string         `json:"query,omitempty"`
	Filters map[string]any `json:"filters,omitempty"`
	Sort    string         `json:"sort,omitempty"`
	Order   string         `json:"order,omitempty"`
	Cursor  string         `json:"cursor,omitempty"`
	Limit   int            `json:"limit"`
}

// BrowsePage is one page of a browse. Total is the filtered count, when the source knows it.
type BrowsePage struct {
	Games      []GameDTO `json:"games"`
	NextCursor string    `json:"nextCursor,omitempty"`
	Total      *int      `json:"total,omitempty"`
}

// FiltersRequest asks a source for the values of one dynamic sidebar group.
type FiltersRequest struct {
	Source string `json:"source"`
	Group  string `json:"group"`
	Query  string `json:"query,omitempty"`
	Limit  int    `json:"limit"`
}

// FiltersPage lists the values of one dynamic sidebar group.
type FiltersPage struct {
	Options []FilterOption `json:"options"`
}

// FilterOption is one value of a dynamic group. Count is optional and counts the whole source.
type FilterOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Count *int   `json:"count,omitempty"`
}
