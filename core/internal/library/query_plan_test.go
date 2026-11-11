package library

import (
	"context"
	"strings"
	"testing"
)

// explainList returns the SQLite query plan of the page query that List runs for a filter.
func explainList(t *testing.T, lib *Library, f Filter) string {
	t.Helper()
	where, args := buildWhere(f)
	args = append(args, defaultLimit, 0)
	rows, err := lib.db.QueryContext(context.Background(), "explain query plan "+pageQuery(where, f), args...)
	if err != nil {
		t.Fatalf("explain %+v: %v", f, err)
	}
	defer rows.Close()
	var lines []string
	for rows.Next() {
		var id, parent, unused int
		var detail string
		if err := rows.Scan(&id, &parent, &unused, &detail); err != nil {
			t.Fatal(err)
		}
		lines = append(lines, detail)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return strings.Join(lines, "\n")
}

// Every sort must read its index in order, so a page never sorts the whole library.
func TestListSortsUseTheirIndexes(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	for _, name := range sortNames() {
		for _, order := range []string{"asc", "desc"} {
			for _, source := range []string{"", "f95"} {
				f := Filter{Sort: name, Order: order, Source: source}
				index := "games_sort_" + strings.ToLower(name)
				if source != "" {
					index = "games_source_sort_" + strings.ToLower(name)
				}
				t.Run(name+"/"+order+"/source="+source, func(t *testing.T) {
					plan := explainList(t, lib, f)
					if !strings.Contains(plan, "USING INDEX "+index) {
						t.Errorf("plan does not use %s:\n%s", index, plan)
					}
					if strings.Contains(plan, "TEMP B-TREE") {
						t.Errorf("plan sorts a temporary table:\n%s", plan)
					}
				})
			}
		}
	}
}

// The full-text search filters through the FTS index, not a scan of every title.
func TestSearchUsesTheFullTextIndex(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	plan := explainList(t, lib, Filter{Search: "harbor"})
	if !strings.Contains(plan, "games_fts") {
		t.Errorf("search plan does not use games_fts:\n%s", plan)
	}
}
