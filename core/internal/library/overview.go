package library

import (
	"context"
	"sort"
	"sync"
	"time"

	"light-launcher/core/internal/contract"
)

const (
	// manualSourceID is the source of games added by hand, linked folders and scanned folders.
	manualSourceID    = "manual"
	manualSourceLabel = "Local games"

	// Card shapes a source can declare. Poster is the default.
	tilePoster = "poster"
	tileWide   = "wide"
	tileSquare = "square"

	// overviewMinInterval is the shortest time between two counts of the library.
	overviewMinInterval = time.Second
)

// SourceCount is a source with its label, card box and how many visible games it has.
type SourceCount struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	// Kind is local for games from folders or by hand, plugin for imported source games.
	Kind string `json:"kind"`
	// Plugin is the plugin that supplies a plugin source. Empty for local games.
	Plugin string `json:"plugin,omitempty"`
	// Capabilities are the source features the UI may show: browse, listGames, sync:*,
	// updateCheck, resolveFiles, auth, browse:filters.
	Capabilities []string `json:"capabilities"`
	// Sync says the source can import or sync the user's games.
	Sync bool `json:"sync"`
	// UpdateCheck says the source can look for newer game versions.
	UpdateCheck bool `json:"updateCheck"`
	// Tile is the older card shape name, kept for consumers that predate Cover.
	Tile string `json:"tile"`
	// Cover is the card box as "W:H".
	Cover string `json:"cover"`
	Count int    `json:"count"`
	// Actions are the buttons the source adds to the Library. Never null.
	Actions []contract.SourceAction `json:"actions"`
	// Card names the game fields the source's cards show. Null uses the defaults.
	Card *contract.SourceCard `json:"card"`
	// Sidebar holds the filter groups the source declares for the Library sidebar. Empty for local games.
	Sidebar []contract.SidebarGroup `json:"sidebar,omitempty"`
}

// Overview is the summary the library shows beside its list. It covers the whole visible library, not a filter.
type Overview struct {
	Sources []SourceCount `json:"sources"`
	Tags    []TagCount    `json:"tags"`
	Stats   Stats         `json:"stats"`
}

// overviewCounts is what the database says about the library. Labels come from the plugins on each call.
type overviewCounts struct {
	sources map[string]int
	tags    []TagCount
	stats   Stats
}

// overviewCache keeps the last counts. It recounts after a write, at most once per overviewMinInterval.
type overviewCache struct {
	mu       sync.Mutex
	valid    bool
	writes   uint64
	computed time.Time
	counts   overviewCounts
}

// Overview returns the sources, tags and stats of the visible library.
func (l *Library) Overview(ctx context.Context) (Overview, error) {
	counts, err := l.cachedCounts(ctx)
	if err != nil {
		return Overview{}, err
	}
	return Overview{Sources: l.sourceList(counts.sources), Tags: counts.tags, Stats: counts.stats}, nil
}

// cachedCounts returns the counts, recounting only after a write. A recount that comes sooner than
// overviewMinInterval after the last one waits for it.
func (l *Library) cachedCounts(ctx context.Context) (overviewCounts, error) {
	cache := &l.overview
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.valid && cache.writes == l.writeCount() {
		return cache.counts, nil
	}
	if cache.valid {
		if err := sleepUntil(ctx, cache.computed.Add(overviewMinInterval)); err != nil {
			return overviewCounts{}, err
		}
	}
	writes := l.writeCount()
	counts, err := l.countLibrary(ctx)
	if err != nil {
		return overviewCounts{}, err
	}
	cache.valid, cache.writes, cache.computed, cache.counts = true, writes, time.Now(), counts
	return counts, nil
}

// sleepUntil waits until t, or stops early when ctx ends.
func sleepUntil(ctx context.Context, t time.Time) error {
	timer := time.NewTimer(time.Until(t))
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *Library) countLibrary(ctx context.Context) (overviewCounts, error) {
	sources, err := l.sourceCounts(ctx)
	if err != nil {
		return overviewCounts{}, err
	}
	tags, err := l.TagCounts(ctx)
	if err != nil {
		return overviewCounts{}, err
	}
	stats, err := l.Stats(ctx)
	if err != nil {
		return overviewCounts{}, err
	}
	return overviewCounts{sources: sources, tags: tags, stats: stats}, nil
}

// sourceCounts counts the visible games of each source.
func (l *Library) sourceCounts(ctx context.Context) (map[string]int, error) {
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
	return counts, rows.Err()
}

// sourceList joins the counted sources with the sources the plugins offer, so empty sources show too.
func (l *Library) sourceList(counts map[string]int) []SourceCount {
	infos := map[string]contract.SourceInfo{}
	for _, info := range l.sources.Sources() {
		infos[info.SourceID] = info
	}
	ids := map[string]bool{}
	for id := range counts {
		ids[id] = true
	}
	for id := range infos {
		ids[id] = true
	}
	out := make([]SourceCount, 0, len(ids))
	for id := range ids {
		out = append(out, sourceCountOf(id, counts[id], infos[id]))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// sourceCountOf builds the entry for one source. The local source has a fixed label and a square card.
func sourceCountOf(id string, count int, info contract.SourceInfo) SourceCount {
	if id == manualSourceID {
		return SourceCount{ID: id, Label: manualSourceLabel, Kind: "local", Tile: tileSquare,
			Cover: contract.NormalizeCover("", tileSquare), Count: count, Actions: []contract.SourceAction{},
			Capabilities: []string{"launch", "scan"}}
	}
	label := info.Label
	if label == "" {
		label = id
	}
	return SourceCount{
		ID:           id,
		Label:        label,
		Kind:         "plugin",
		Plugin:       info.Plugin,
		Capabilities: append([]string{}, info.Capabilities...),
		Sync:         hasCapability(info, CapabilityListGames) || hasCapability(info, "sync:games") || hasCapability(info, "sync:index"),
		UpdateCheck:  hasCapability(info, CapabilityUpdateCheck),
		Tile:         cardShape(info.Tile),
		Cover:        contract.NormalizeCover(info.Cover, info.Tile),
		Count:        count,
		Actions:      actionList(info.Actions),
		Card:         info.Card,
		Sidebar:      info.Sidebar,
	}
}

// actionList returns the actions, or an empty list, so the JSON array is never null.
func actionList(actions []contract.SourceAction) []contract.SourceAction {
	if actions == nil {
		return []contract.SourceAction{}
	}
	return actions
}

// cardShape returns the declared card shape, or the poster shape when the source declares none it knows.
func cardShape(tile string) string {
	switch tile {
	case tileWide, tileSquare:
		return tile
	default:
		return tilePoster
	}
}
