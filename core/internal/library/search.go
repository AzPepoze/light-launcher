package library

import (
	"context"
	"sync"

	"light-launcher/core/internal/contract"
)

// SearchResult is one source's answer to a search.
type SearchResult struct {
	Plugin string             `json:"plugin"`
	Source string             `json:"source"`
	Games  []contract.GameDTO `json:"games"`
	Error  string             `json:"error,omitempty"`
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
	result := SearchResult{Plugin: info.Plugin, Source: info.SourceID, Games: []contract.GameDTO{}}
	var reply searchReply
	if err := sources.Call(ctx, info.Plugin, methodSearch, searchParams{Source: info.SourceID, Query: query}, &reply); err != nil {
		result.Error = err.Error()
		return result
	}
	if reply.Games != nil {
		result.Games = reply.Games
	}
	return result
}
