package library

import (
	"context"
	"errors"
	"fmt"

	"light-launcher/core/internal/contract"
)

// Limits of the browse calls, as the plugin SDK states them.
const (
	browseMaxLimit  = 100
	filtersMaxLimit = 200
)

var (
	// ErrUnknownSource means no plugin offers the source id.
	ErrUnknownSource = errors.New("unknown source")
	// ErrNoBrowseFilters means the source does not declare the browse:filters capability.
	ErrNoBrowseFilters = errors.New("source does not offer browse filters")
)

// Browse asks a source for one page of games with the sidebar filters, sort and cursor. It never saves
// the games: the answer is for display only.
func (l *Library) Browse(ctx context.Context, request contract.BrowseRequest) (contract.BrowsePage, error) {
	info, err := l.browseSource(request.Source)
	if err != nil {
		return contract.BrowsePage{}, err
	}
	if request.Limit < 1 || request.Limit > browseMaxLimit {
		return contract.BrowsePage{}, fmt.Errorf("limit must be 1 to %d, not %d", browseMaxLimit, request.Limit)
	}
	if !validOrder(request.Order) {
		return contract.BrowsePage{}, fmt.Errorf("order must be asc or desc, not %q", request.Order)
	}
	return l.sources.Browse(ctx, info.Plugin, request)
}

// ListFilters returns the values of one dynamic sidebar group of a source, such as its tags.
func (l *Library) ListFilters(ctx context.Context, request contract.FiltersRequest) (contract.FiltersPage, error) {
	info, err := l.browseSource(request.Source)
	if err != nil {
		return contract.FiltersPage{}, err
	}
	if request.Group == "" {
		return contract.FiltersPage{}, errors.New("group is required")
	}
	if request.Limit < 1 || request.Limit > filtersMaxLimit {
		return contract.FiltersPage{}, fmt.Errorf("limit must be 1 to %d, not %d", filtersMaxLimit, request.Limit)
	}
	return l.sources.ListFilters(ctx, info.Plugin, request)
}

// browseSource finds the source and checks that it offers browse filters.
func (l *Library) browseSource(sourceID string) (contract.SourceInfo, error) {
	info, ok := l.pluginFor(sourceID)
	if !ok {
		return contract.SourceInfo{}, ErrUnknownSource
	}
	if !hasCapability(info, CapabilityBrowseFilters) {
		return contract.SourceInfo{}, ErrNoBrowseFilters
	}
	return info, nil
}

// validOrder accepts an empty order, which the plugin reads as its default, or asc or desc.
func validOrder(order string) bool {
	return order == "" || order == "asc" || order == "desc"
}
