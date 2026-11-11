package library

import (
	"fmt"
	"sort"
	"strings"
)

// searchMigration adds the full-text index behind library search and the indexes behind the list sorts.
var searchMigration = migration{
	version: 4,
	statements: append([]string{
		// The trigram tokenizer matches any three or more characters inside the text, like a substring search.
		`create virtual table if not exists games_fts using fts5(title, developer, tags, content='games',
			content_rowid='id', tokenize='trigram')`,
		`create trigger if not exists games_fts_insert after insert on games begin
			insert into games_fts (rowid, title, developer, tags) values (new.id, new.title, new.developer, new.tags);
		end`,
		`create trigger if not exists games_fts_delete after delete on games begin
			insert into games_fts (games_fts, rowid, title, developer, tags)
			values ('delete', old.id, old.title, old.developer, old.tags);
		end`,
		`create trigger if not exists games_fts_update after update of title, developer, tags on games begin
			insert into games_fts (games_fts, rowid, title, developer, tags)
			values ('delete', old.id, old.title, old.developer, old.tags);
			insert into games_fts (rowid, title, developer, tags) values (new.id, new.title, new.developer, new.tags);
		end`,
		`insert into games_fts (games_fts) values ('rebuild')`,
	}, sortIndexStatements()...),
}

// sortIndexStatements creates two indexes per sort: one for all games and one for a source. Each starts
// with the equality columns of the list and then the sort expression, which must match the query text
// exactly for SQLite to use it.
func sortIndexStatements() []string {
	var statements []string
	for _, name := range sortNames() {
		key := strings.ToLower(name)
		expression := sortColumns[name]
		statements = append(statements,
			fmt.Sprintf(`create index if not exists games_sort_%s on games (hidden, %s)`, key, expression),
			fmt.Sprintf(`create index if not exists games_source_sort_%s on games (source, hidden, %s)`, key, expression),
		)
	}
	return statements
}

// sortNames lists the sort keys in a fixed order.
func sortNames() []string {
	names := make([]string, 0, len(sortColumns))
	for name := range sortColumns {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
