package library

import (
	"context"
	"path/filepath"
	"sort"
	"testing"

	"light-launcher/core/internal/contract"
)

// searchFixture stores three games that differ in where their words sit: title, developer and tags.
func searchFixture(t *testing.T, lib *Library) {
	t.Helper()
	games := []contract.GameDTO{
		{Source: "manual", SourceID: "a", URL: "u", Title: "Moonlit Harbor", Developer: "Blue Fox", Tags: []string{"pixel-art"}},
		{Source: "manual", SourceID: "b", URL: "u", Title: "Rainy Coast", Developer: "Night Owl", Tags: []string{"rpg"}},
		{Source: "manual", SourceID: "c", URL: "u", Title: "Iron Sky", Developer: "Stone Works", Tags: []string{"strategy"}},
	}
	for _, game := range games {
		if _, err := lib.Upsert(context.Background(), game); err != nil {
			t.Fatal(err)
		}
	}
}

// titlesOf lists the sorted titles of a filtered list.
func titlesOf(t *testing.T, lib *Library, f Filter) []string {
	t.Helper()
	page, err := lib.List(context.Background(), f)
	if err != nil {
		t.Fatalf("list %+v: %v", f, err)
	}
	var titles []string
	for _, game := range page.Items {
		titles = append(titles, game.Title)
	}
	sort.Strings(titles)
	return titles
}

func TestSearchMatchesThreeOrMoreCharactersAnywhere(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	searchFixture(t, lib)

	cases := []struct {
		name  string
		query string
		want  []string
	}{
		{"substring of title", "arbo", []string{"Moonlit Harbor"}},
		{"case is ignored", "HARBOR", []string{"Moonlit Harbor"}},
		{"developer", "blue fox", []string{"Moonlit Harbor"}},
		{"tag with a dash", "ixel-a", []string{"Moonlit Harbor"}},
		{"tag", "rpg", []string{"Rainy Coast"}},
		{"surrounding spaces are trimmed", "   arbo  ", []string{"Moonlit Harbor"}},
		{"quotes are plain text", `a"b`, nil},
		{"no match", "zzz", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := titlesOf(t, lib, Filter{Search: tc.query})
			if !equalStrings(got, tc.want) {
				t.Errorf("search %q = %v, want %v", tc.query, got, tc.want)
			}
		})
	}
}

func TestSearchOfOneOrTwoCharactersMatchesTitleStart(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	searchFixture(t, lib)

	cases := []struct {
		name  string
		query string
		want  []string
	}{
		{"two letters match the start of a title", "mo", []string{"Moonlit Harbor"}},
		{"one letter", "r", []string{"Rainy Coast"}},
		{"a word inside a title does not match", "ar", nil},
		{"a developer does not match", "bl", nil},
		{"percent is a plain character", "%", nil},
		{"underscore is a plain character", "_", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := titlesOf(t, lib, Filter{Search: tc.query})
			if !equalStrings(got, tc.want) {
				t.Errorf("search %q = %v, want %v", tc.query, got, tc.want)
			}
		})
	}
}

func TestSearchIndexFollowsUpdatesAndDeletes(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	searchFixture(t, lib)
	ctx := context.Background()
	coast := mustGameByTitle(t, lib, "Rainy Coast")

	newTitle := "Glass Garden"
	if _, err := lib.Patch(ctx, coast.ID, Patch{Title: &newTitle}); err != nil {
		t.Fatal(err)
	}
	if got := titlesOf(t, lib, Filter{Search: "rainy"}); got != nil {
		t.Errorf("old title still found: %v", got)
	}
	if got := titlesOf(t, lib, Filter{Search: "garden"}); !equalStrings(got, []string{"Glass Garden"}) {
		t.Errorf("new title not found: %v", got)
	}

	newTags := []string{"puzzle"}
	if _, err := lib.Patch(ctx, coast.ID, Patch{Tags: &newTags}); err != nil {
		t.Fatal(err)
	}
	if got := titlesOf(t, lib, Filter{Search: "puzzle"}); !equalStrings(got, []string{"Glass Garden"}) {
		t.Errorf("new tag not found: %v", got)
	}

	if err := lib.Remove(ctx, coast.ID); err != nil {
		t.Fatal(err)
	}
	if got := titlesOf(t, lib, Filter{Search: "garden"}); got != nil {
		t.Errorf("removed game still found: %v", got)
	}
}

func TestSearchIndexIsBuiltForGamesStoredBeforeTheMigration(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "library.db")
	ctx := context.Background()
	lib, err := Open(ctx, Options{DBPath: path, MediaDir: filepath.Join(dir, "media")})
	if err != nil {
		t.Fatal(err)
	}
	upsertGame(t, lib, "old", "Harbor Before")
	for _, statement := range []string{
		`drop trigger games_fts_insert`, `drop trigger games_fts_delete`, `drop trigger games_fts_update`,
		`drop table games_fts`, `pragma user_version = 3`,
	} {
		if _, err := lib.db.ExecContext(ctx, statement); err != nil {
			t.Fatalf("%s: %v", statement, err)
		}
	}
	if err := lib.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(ctx, Options{DBPath: path, MediaDir: filepath.Join(dir, "media")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { reopened.Close() })
	if got := titlesOf(t, reopened, Filter{Search: "arbo"}); !equalStrings(got, []string{"Harbor Before"}) {
		t.Errorf("search after migration = %v, want the stored game", got)
	}
}

func mustGameByTitle(t *testing.T, lib *Library, title string) Game {
	t.Helper()
	page, err := lib.List(context.Background(), Filter{Search: title})
	if err != nil || len(page.Items) != 1 {
		t.Fatalf("find %q: %v, %d items", title, err, len(page.Items))
	}
	return page.Items[0]
}

// equalStrings compares two lists, treating nil and empty as equal.
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
