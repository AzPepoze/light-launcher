package library

import (
	"context"
	"testing"

	"light-launcher/core/internal/contract"
)

func TestListingCountsSourcesTagsAndStats(t *testing.T) {
	sources := &fakeSources{infos: []contract.SourceInfo{{Plugin: "f95", SourceID: "f95", Label: "F95 Zone"}}}
	lib := openTestLibrary(t, Options{Sources: sources})
	ctx := context.Background()
	upsertGame(t, lib, "list-1", "One")
	if _, err := lib.Upsert(ctx, contract.GameDTO{Source: "f95", SourceID: "2", URL: "https://f95/2",
		Title: "Two", Tags: []string{"rpg"}}); err != nil {
		t.Fatal(err)
	}
	hiddenID, err := lib.Upsert(ctx, contract.GameDTO{Source: "f95", SourceID: "3", URL: "https://f95/3",
		Title: "Three", Tags: []string{"hidden-only"}})
	if err != nil {
		t.Fatal(err)
	}
	hidden := true
	if _, err := lib.Patch(ctx, hiddenID, Patch{Hidden: &hidden}); err != nil {
		t.Fatal(err)
	}

	listing, err := lib.Listing(ctx, Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if listing.Total != 2 || len(listing.Items) != 2 {
		t.Errorf("total = %d, items = %d; hidden games must not be listed", listing.Total, len(listing.Items))
	}
	if len(listing.Sources) != 2 || listing.Sources[0].ID != "f95" || listing.Sources[0].Label != "F95 Zone" {
		t.Fatalf("sources = %+v", listing.Sources)
	}
	if listing.Sources[0].Count != 1 || listing.Sources[1].ID != "manual" || listing.Sources[1].Count != 1 {
		t.Errorf("source counts = %+v, want f95 with 1 and manual with 1", listing.Sources)
	}
	if len(listing.Tags) != 1 || listing.Tags[0].Tag != "rpg" || listing.Tags[0].Count != 1 {
		t.Errorf("tags = %+v, want rpg once", listing.Tags)
	}
	if listing.Stats.Total != 2 {
		t.Errorf("stats.total = %d, want 2", listing.Stats.Total)
	}
}
