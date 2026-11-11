package library

import (
	"context"
	"errors"
	"testing"

	"light-launcher/core/internal/contract"
)

// browseSources records the browse calls and answers them with a fixed page or error.
type browseSources struct {
	fakeSources
	page        contract.BrowsePage
	filtersPage contract.FiltersPage
	err         error
	browses     []contract.BrowseRequest
	filters     []contract.FiltersRequest
}

func (b *browseSources) Browse(_ context.Context, plugin string, request contract.BrowseRequest) (contract.BrowsePage, error) {
	b.browses = append(b.browses, request)
	return b.page, b.err
}

func (b *browseSources) ListFilters(_ context.Context, plugin string, request contract.FiltersRequest) (contract.FiltersPage, error) {
	b.filters = append(b.filters, request)
	return b.filtersPage, b.err
}

func browseTestSources() *browseSources {
	return &browseSources{fakeSources: fakeSources{infos: []contract.SourceInfo{
		{Plugin: "shop", SourceID: "shop", Capabilities: []string{CapabilityBrowse, CapabilityBrowseFilters}},
		{Plugin: "plain", SourceID: "plain", Capabilities: []string{CapabilityBrowse}},
	}}}
}

func TestBrowseChecksTheRequest(t *testing.T) {
	cases := []struct {
		name    string
		request contract.BrowseRequest
		wantErr error
	}{
		{"unknown source", contract.BrowseRequest{Source: "nope", Limit: 10}, ErrUnknownSource},
		{"no browse:filters", contract.BrowseRequest{Source: "plain", Limit: 10}, ErrNoBrowseFilters},
		{"limit zero", contract.BrowseRequest{Source: "shop", Limit: 0}, nil},
		{"limit above 100", contract.BrowseRequest{Source: "shop", Limit: 101}, nil},
		{"order not asc or desc", contract.BrowseRequest{Source: "shop", Limit: 10, Order: "up"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sources := browseTestSources()
			lib := openTestLibrary(t, Options{Sources: sources})
			_, err := lib.Browse(context.Background(), tc.request)
			if err == nil {
				t.Fatalf("expected an error")
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if len(sources.browses) != 0 {
				t.Fatalf("plugin was called for a bad request: %+v", sources.browses)
			}
		})
	}
}

func TestBrowseForwardsTheRequestAndSavesNothing(t *testing.T) {
	sources := browseTestSources()
	total := 1
	sources.page = contract.BrowsePage{
		Games: []contract.GameDTO{{Source: "shop", SourceID: "g1", URL: "https://shop/g1", Title: "Game One"}},
		Total: &total,
	}
	lib := openTestLibrary(t, Options{Sources: sources})

	request := contract.BrowseRequest{Source: "shop", Query: "one", Sort: "date", Order: "desc", Limit: 50}
	page, err := lib.Browse(context.Background(), request)
	if err != nil {
		t.Fatalf("Browse: %v", err)
	}
	if len(page.Games) != 1 || page.Games[0].Title != "Game One" {
		t.Fatalf("page = %+v", page)
	}
	if len(sources.browses) != 1 || sources.browses[0].Query != "one" || sources.browses[0].Order != "desc" {
		t.Fatalf("plugin got %+v", sources.browses)
	}
	saved, err := lib.List(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(saved.Items) != 0 {
		t.Fatalf("browse saved %d games, want none", len(saved.Items))
	}
}

func TestBrowseReturnsInvalidCursorAsItsOwnError(t *testing.T) {
	sources := browseTestSources()
	sources.err = contract.ErrInvalidCursor
	lib := openTestLibrary(t, Options{Sources: sources})

	_, err := lib.Browse(context.Background(), contract.BrowseRequest{Source: "shop", Cursor: "old", Limit: 10})
	if !errors.Is(err, contract.ErrInvalidCursor) {
		t.Fatalf("err = %v, want invalid cursor", err)
	}
}

func TestListFiltersChecksTheRequest(t *testing.T) {
	cases := []struct {
		name    string
		request contract.FiltersRequest
		wantErr error
	}{
		{"unknown source", contract.FiltersRequest{Source: "nope", Group: "tags", Limit: 10}, ErrUnknownSource},
		{"no browse:filters", contract.FiltersRequest{Source: "plain", Group: "tags", Limit: 10}, ErrNoBrowseFilters},
		{"no group", contract.FiltersRequest{Source: "shop", Limit: 10}, nil},
		{"limit zero", contract.FiltersRequest{Source: "shop", Group: "tags", Limit: 0}, nil},
		{"limit above 200", contract.FiltersRequest{Source: "shop", Group: "tags", Limit: 201}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sources := browseTestSources()
			lib := openTestLibrary(t, Options{Sources: sources})
			_, err := lib.ListFilters(context.Background(), tc.request)
			if err == nil {
				t.Fatalf("expected an error")
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if len(sources.filters) != 0 {
				t.Fatalf("plugin was called for a bad request: %+v", sources.filters)
			}
		})
	}
}

func TestListFiltersForwardsTheGroup(t *testing.T) {
	sources := browseTestSources()
	count := 7
	sources.filtersPage = contract.FiltersPage{Options: []contract.FilterOption{{Value: "rpg", Label: "RPG", Count: &count}}}
	lib := openTestLibrary(t, Options{Sources: sources})

	page, err := lib.ListFilters(context.Background(), contract.FiltersRequest{Source: "shop", Group: "tags", Query: "r", Limit: 20})
	if err != nil {
		t.Fatalf("ListFilters: %v", err)
	}
	if len(page.Options) != 1 || page.Options[0].Value != "rpg" {
		t.Fatalf("page = %+v", page)
	}
	if len(sources.filters) != 1 || sources.filters[0].Group != "tags" || sources.filters[0].Query != "r" {
		t.Fatalf("plugin got %+v", sources.filters)
	}
}
