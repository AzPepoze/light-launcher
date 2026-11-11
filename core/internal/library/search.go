package library

import (
	"context"
	"sort"
	"sync"

	"light-launcher/core/internal/contract"
)

// SearchResult is one source's answer to a search. Games are sorted by score, best first.
type SearchResult struct {
	Plugin string       `json:"plugin"`
	Source string       `json:"source"`
	Games  []SearchGame `json:"games"`
	Error  string       `json:"error,omitempty"`
}

// SearchGame is a game a source returned, with how well its title matches the query.
// The game fields are flattened into the JSON object.
type SearchGame struct {
	contract.GameDTO
	Score float64 `json:"score"`
}

// SearchSources asks every source that can browse for games matching the query.
// Sources run in parallel, and one failing source does not hide the others.
func SearchSources(ctx context.Context, sources contract.Sources, query string) []SearchResult {
	var targets []contract.SourceInfo
	seen := map[string]bool{}
	for _, info := range sources.Sources() {
		if hasCapability(info, CapabilityBrowse) && !seen[info.SourceID] {
			seen[info.SourceID] = true
			targets = append(targets, info)
		}
	}
	results := make([]SearchResult, len(targets))
	var wg sync.WaitGroup
	for i, info := range targets {
		wg.Add(1)
		go func(i int, info contract.SourceInfo) {
			defer wg.Done()
			results[i] = searchPlugin(ctx, sources, info, query)
		}(i, info)
	}
	wg.Wait()
	return results
}

func searchPlugin(ctx context.Context, sources contract.Sources, info contract.SourceInfo, query string) SearchResult {
	result := SearchResult{Plugin: info.Plugin, Source: info.SourceID, Games: []SearchGame{}}
	var reply searchReply
	if err := sources.Call(ctx, info.Plugin, methodSearch, searchParams{Source: info.SourceID, Query: query}, &reply); err != nil {
		result.Error = err.Error()
		return result
	}
	result.Games = rankGames(query, reply.Games)
	return result
}

// rankGames scores each game title against the query and sorts the best matches first.
func rankGames(query string, games []contract.GameDTO) []SearchGame {
	ranked := make([]SearchGame, 0, len(games))
	for _, game := range games {
		ranked = append(ranked, SearchGame{GameDTO: game, Score: scoreMatch(query, game.Title)})
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].Score > ranked[j].Score })
	return ranked
}
