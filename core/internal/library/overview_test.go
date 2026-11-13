package library

import (
	"context"
	"reflect"
	"testing"
	"time"

	"light-launcher/core/internal/contract"
)

func TestOverviewCountsSourcesTagsAndStats(t *testing.T) {
	sources := &fakeSources{infos: []contract.SourceInfo{
		{Plugin: "f95", SourceID: "f95", Label: "F95 Zone", Tile: "wide"},
		{Plugin: "itch", SourceID: "itch", Label: "itch.io"},
	}}
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

	overview, err := lib.Overview(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := []SourceCount{
		{ID: "f95", Label: "F95 Zone", Kind: "plugin", Plugin: "f95", Capabilities: []string{}, Tile: "wide", Cover: "16:9", Actions: []contract.SourceAction{}, Count: 1},
		{ID: "itch", Label: "itch.io", Kind: "plugin", Plugin: "itch", Capabilities: []string{}, Tile: "poster", Cover: "3:4", Actions: []contract.SourceAction{}, Count: 0},
		{ID: "manual", Label: "Local games", Kind: "local", Capabilities: []string{"launch", "scan"}, Tile: "square", Cover: "1:1", Actions: []contract.SourceAction{}, Count: 1},
	}
	if len(overview.Sources) != len(want) {
		t.Fatalf("sources = %+v, want %+v", overview.Sources, want)
	}
	for i := range want {
		if !reflect.DeepEqual(overview.Sources[i], want[i]) {
			t.Errorf("source %d = %+v, want %+v", i, overview.Sources[i], want[i])
		}
	}
	if len(overview.Tags) != 1 || overview.Tags[0] != (TagCount{Tag: "rpg", Count: 1}) {
		t.Errorf("tags = %+v, want rpg once; hidden games must not count", overview.Tags)
	}
	if overview.Stats.Total != 2 {
		t.Errorf("stats.total = %d, want 2", overview.Stats.Total)
	}
}

func TestOverviewCarriesCoverActionsAndCard(t *testing.T) {
	action := contract.SourceAction{ID: "sync", Label: "Sync", Method: "demo.sync", Scope: contract.ActionScopeLibrary}
	card := &contract.SourceCard{Title: "title", Badges: []string{"installed"}}
	sources := &fakeSources{infos: []contract.SourceInfo{
		{Plugin: "demo", SourceID: "demo", Label: "Demo", Tile: "wide", Cover: "4:3", Actions: []contract.SourceAction{action}, Card: card},
	}}
	lib := openTestLibrary(t, Options{Sources: sources})

	overview, err := lib.Overview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	demo := overview.Sources[0]
	if demo.Cover != "4:3" {
		t.Errorf("cover = %q, want 4:3 from the source", demo.Cover)
	}
	if !reflect.DeepEqual(demo.Actions, []contract.SourceAction{action}) {
		t.Errorf("actions = %+v, want the source action", demo.Actions)
	}
	if !reflect.DeepEqual(demo.Card, card) {
		t.Errorf("card = %+v, want the source card", demo.Card)
	}
}

func TestOverviewCoverFallsBackToTile(t *testing.T) {
	sources := &fakeSources{infos: []contract.SourceInfo{{Plugin: "demo", SourceID: "demo", Label: "Demo", Tile: "square"}}}
	lib := openTestLibrary(t, Options{Sources: sources})

	overview, err := lib.Overview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if demo := overview.Sources[0]; demo.Cover != "1:1" || demo.Card != nil {
		t.Errorf("source = %+v, want cover 1:1 from the tile and no card", demo)
	}
}

func TestOverviewCarriesTheSourceSidebar(t *testing.T) {
	sidebar := []contract.SidebarGroup{
		{Type: "search", ID: "text", Placeholder: "Search titles"},
		{Type: "multi", ID: "tags", Label: "Tags", Dynamic: true, Searchable: true},
	}
	sources := &fakeSources{infos: []contract.SourceInfo{
		{Plugin: "shop", SourceID: "shop", Label: "Shop", Sidebar: sidebar},
		{Plugin: "plain", SourceID: "plain", Label: "Plain"},
	}}
	lib := openTestLibrary(t, Options{Sources: sources})

	overview, err := lib.Overview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]SourceCount{}
	for _, source := range overview.Sources {
		byID[source.ID] = source
	}
	if !reflect.DeepEqual(byID["shop"].Sidebar, sidebar) {
		t.Errorf("shop sidebar = %+v, want %+v", byID["shop"].Sidebar, sidebar)
	}
	if len(byID["plain"].Sidebar) != 0 || len(byID["manual"].Sidebar) != 0 {
		t.Errorf("sources without a sidebar must carry none: plain=%+v manual=%+v",
			byID["plain"].Sidebar, byID["manual"].Sidebar)
	}
}

func TestOverviewSeesWritesAfterTheThrottleWindow(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	ctx := context.Background()
	upsertGame(t, lib, "first", "First")
	before, err := lib.Overview(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if before.Stats.Total != 1 {
		t.Fatalf("stats.total before = %d, want 1", before.Stats.Total)
	}

	upsertGame(t, lib, "second", "Second")
	start := time.Now()
	after, err := lib.Overview(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if after.Stats.Total != 2 {
		t.Errorf("stats.total after a write = %d, want 2", after.Stats.Total)
	}
	if waited := time.Since(start); waited > overviewMinInterval+500*time.Millisecond {
		t.Errorf("recount waited %v, want at most one interval", waited)
	}
}

func TestOverviewStopsWaitingWhenContextEnds(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	upsertGame(t, lib, "first", "First")
	if _, err := lib.Overview(context.Background()); err != nil {
		t.Fatal(err)
	}
	upsertGame(t, lib, "second", "Second")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := lib.Overview(ctx); err == nil {
		t.Error("a cancelled context must end the wait for a recount")
	}
}
