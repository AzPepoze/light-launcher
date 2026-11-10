package library

import (
	"context"
	"sort"
)

// SourceCount is a source with its label and how many visible games come from it.
type SourceCount struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

// Listing is one page of games plus the counts the library header and filters show.
// The counts cover the whole visible library, not only the page or the filter.
type Listing struct {
	Page
	Sources []SourceCount `json:"sources"`
	Tags    []TagCount    `json:"tags"`
	Stats   Stats         `json:"stats"`
}

// Listing returns one page of games with the sources, tags and stats of the library.
func (l *Library) Listing(ctx context.Context, f Filter) (Listing, error) {
	page, err := l.List(ctx, f)
	if err != nil {
		return Listing{}, err
	}
	sources, err := l.sourceCounts(ctx)
	if err != nil {
		return Listing{}, err
	}
	tags, err := l.TagCounts(ctx)
	if err != nil {
		return Listing{}, err
	}
	stats, err := l.Stats(ctx)
	if err != nil {
		return Listing{}, err
	}
	return Listing{Page: page, Sources: sources, Tags: tags, Stats: stats}, nil
}

// sourceCounts lists the configured sources and any other source that has games, sorted by id.
func (l *Library) sourceCounts(ctx context.Context) ([]SourceCount, error) {
	rows, err := l.db.QueryContext(ctx, `select source, count(*) from games where hidden = 0 group by source`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := map[string]int{}
	for rows.Next() {
		var source string
		var count int
		if err := rows.Scan(&source, &count); err != nil {
			return nil, err
		}
		counts[source] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	labels := map[string]string{}
	for _, info := range l.sources.Sources() {
		labels[info.SourceID] = info.Label
		if _, ok := counts[info.SourceID]; !ok {
			counts[info.SourceID] = 0
		}
	}
	out := make([]SourceCount, 0, len(counts))
	for id, count := range counts {
		label := labels[id]
		if label == "" {
			label = id
		}
		out = append(out, SourceCount{ID: id, Label: label, Count: count})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
