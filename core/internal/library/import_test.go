package library

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"light-launcher/core/internal/contract"
)

const feedSource = "feed"

// feedWords builds titles, developers and tags for the fixture feed, so searches match a steady share of games.
var feedWords = []string{"Moonlit", "Harbor", "Rainy", "Coast", "Iron", "Sky", "Glass", "Garden", "Ember", "Frost",
	"Silver", "River", "Night", "Owl", "Stone", "Works", "Blue", "Fox", "Pixel", "Story"}

// feedGame is game i of the fixture feed. Its update stamp and version never change, so a repeat import skips it.
func feedGame(i int) contract.GameDTO {
	id := strconv.Itoa(i)
	return contract.GameDTO{
		Source: feedSource, SourceID: id, URL: "https://feed/" + id,
		Title:     fmt.Sprintf("%s %s %d", feedWords[i%20], feedWords[(i*7)%20], i),
		Developer: feedWords[(i*3)%20],
		Tags:      []string{feedWords[(i*11)%20]},
		Version:   "1",
		UpdatedAt: batchStamp,
	}
}

// pagedFeed serves total games the way a plugin's source.listGames does. The cursor is the offset of the next game.
func pagedFeed(total int) func(params any) (any, error) {
	return func(params any) (any, error) {
		page := params.(listGamesParams)
		start := 0
		if page.Cursor != "" {
			start, _ = strconv.Atoi(page.Cursor)
		}
		end := min(start+page.Limit, total)
		reply := listGamesReply{}
		for i := start; i < end; i++ {
			reply.Games = append(reply.Games, feedGame(i))
		}
		if end < total {
			reply.NextCursor = strconv.Itoa(end)
		}
		return reply, nil
	}
}

// feedSources is a fake host that offers the feed source with the listGames capability.
func feedSources(replies map[string]func(params any) (any, error)) *fakeSources {
	return &fakeSources{
		infos: []contract.SourceInfo{{Plugin: "feed", SourceID: feedSource, Label: "Feed",
			Capabilities: []string{CapabilityListGames}}},
		replies: replies,
	}
}

func listGamesOnly(total int) map[string]func(params any) (any, error) {
	return map[string]func(params any) (any, error){methodListGames: pagedFeed(total)}
}

func storedCount(t *testing.T, lib *Library) int {
	t.Helper()
	page, err := lib.List(context.Background(), Filter{Limit: 1, WantTotal: true})
	if err != nil {
		t.Fatal(err)
	}
	return totalOf(page)
}

func TestImportSourceSavesEveryPageAndSkipsRepeats(t *testing.T) {
	lib := openTestLibrary(t, Options{Sources: feedSources(listGamesOnly(250))})
	ctx := context.Background()

	var progress []ImportProgress
	result, err := lib.ImportSource(ctx, "feed", feedSource, func(p ImportProgress) { progress = append(progress, p) })
	if err != nil {
		t.Fatal(err)
	}
	if result != (ImportResult{Imported: 250}) {
		t.Errorf("result = %+v, want 250 imported", result)
	}
	if len(progress) != 3 || progress[0].Imported != 100 || progress[2].Imported != 250 {
		t.Errorf("progress = %+v, want one call per page of 100", progress)
	}
	if storedCount(t, lib) != 250 {
		t.Errorf("stored = %d, want 250", storedCount(t, lib))
	}

	again, err := lib.ImportSource(ctx, "feed", feedSource, nil)
	if err != nil {
		t.Fatal(err)
	}
	if again != (ImportResult{Skipped: 250}) {
		t.Errorf("repeat result = %+v, want 250 skipped", again)
	}
}

func TestImportSourceKeepsPagesSavedBeforeAFailedPage(t *testing.T) {
	replies := map[string]func(params any) (any, error){
		methodListGames: func(params any) (any, error) {
			if params.(listGamesParams).Cursor == "100" {
				return nil, errors.New("plugin is offline")
			}
			return pagedFeed(250)(params)
		},
	}
	lib := openTestLibrary(t, Options{Sources: feedSources(replies)})

	result, err := lib.ImportSource(context.Background(), "feed", feedSource, nil)
	if err == nil {
		t.Fatal("a failed page must end the import with an error")
	}
	if result.Imported != 100 || result.Failed != 1 {
		t.Errorf("result = %+v, want 100 imported and 1 failed", result)
	}
	if got := storedCount(t, lib); got != 100 {
		t.Errorf("stored = %d, want the first page kept", got)
	}
}

func TestImportSourceStopsWhenContextEnds(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	replies := map[string]func(params any) (any, error){
		methodListGames: func(params any) (any, error) {
			if params.(listGamesParams).Cursor == "100" {
				cancel()
				return nil, errors.New("request cancelled")
			}
			return pagedFeed(250)(params)
		},
	}
	lib := openTestLibrary(t, Options{Sources: feedSources(replies)})

	result, err := lib.ImportSource(ctx, "feed", feedSource, nil)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
	if result.Failed != 0 {
		t.Errorf("failed = %d, a cancelled import is not a failed page", result.Failed)
	}
	if got := storedCount(t, lib); got != 100 {
		t.Errorf("stored = %d, want the first page kept", got)
	}
}

func TestImportSourceStopsWhenThePluginRepeatsACursor(t *testing.T) {
	replies := map[string]func(params any) (any, error){
		methodListGames: func(params any) (any, error) {
			page := params.(listGamesParams)
			if page.Cursor == "" {
				return listGamesReply{Games: []contract.GameDTO{feedGame(0)}, NextCursor: "a"}, nil
			}
			return listGamesReply{Games: []contract.GameDTO{feedGame(1)}, NextCursor: "a"}, nil
		},
	}
	lib := openTestLibrary(t, Options{Sources: feedSources(replies)})

	result, err := lib.ImportSource(context.Background(), "feed", feedSource, nil)
	if err == nil {
		t.Fatal("a repeated cursor must end the import")
	}
	if result.Failed != 1 || result.Imported != 1 {
		t.Errorf("result = %+v, want 1 imported and the repeating page failed", result)
	}
	if got := storedCount(t, lib); got != 1 {
		t.Errorf("stored = %d, the repeating page must not be saved", got)
	}
}

func TestImportSourceRejectsGamesFromAnotherSource(t *testing.T) {
	replies := map[string]func(params any) (any, error){
		methodListGames: func(any) (any, error) {
			other := feedGame(0)
			other.Source = "elsewhere"
			return listGamesReply{Games: []contract.GameDTO{other}}, nil
		},
	}
	lib := openTestLibrary(t, Options{Sources: feedSources(replies)})

	if _, err := lib.ImportSource(context.Background(), "feed", feedSource, nil); err == nil {
		t.Fatal("a page with another source's game must fail")
	}
	if got := storedCount(t, lib); got != 0 {
		t.Errorf("stored = %d, want nothing saved", got)
	}
}

func TestStartImportRunsInTheBackgroundOneAtATime(t *testing.T) {
	release := make(chan struct{})
	replies := map[string]func(params any) (any, error){
		methodListGames: func(params any) (any, error) {
			<-release
			return pagedFeed(1)(params)
		},
	}
	lib := openTestLibrary(t, Options{Sources: feedSources(replies)})
	done := make(chan ImportResult, 2)
	onDone := func(result ImportResult, err error) {
		if err != nil {
			t.Errorf("background import: %v", err)
		}
		done <- result
	}

	if err := lib.StartImport(feedSource, nil, onDone); err != nil {
		t.Fatal(err)
	}
	if err := lib.StartImport(feedSource, nil, onDone); !errors.Is(err, ErrImportRunning) {
		t.Errorf("second start err = %v, want ErrImportRunning", err)
	}
	if _, err := lib.ImportSource(context.Background(), "feed", feedSource, nil); !errors.Is(err, ErrImportRunning) {
		t.Errorf("synchronous import during a sync err = %v, want ErrImportRunning", err)
	}

	close(release)
	select {
	case result := <-done:
		if result.Imported != 1 {
			t.Errorf("result = %+v, want 1 imported", result)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the background import never reported done")
	}
	if err := lib.StartImport(feedSource, nil, nil); err != nil {
		t.Errorf("a finished import must release its source: %v", err)
	}
}

func TestStartImportNeedsTheListGamesCapability(t *testing.T) {
	sources := &fakeSources{infos: []contract.SourceInfo{{Plugin: "f95", SourceID: "f95", Capabilities: []string{CapabilityBrowse}}}}
	lib := openTestLibrary(t, Options{Sources: sources})
	if err := lib.StartImport("f95", nil, nil); !errors.Is(err, ErrNoListGames) {
		t.Errorf("err = %v, want ErrNoListGames", err)
	}
	if err := lib.StartImport("missing", nil, nil); !errors.Is(err, ErrNoListGames) {
		t.Errorf("unknown source err = %v, want ErrNoListGames", err)
	}
}

func TestStoppingTheImporterCancelsARunningImport(t *testing.T) {
	replies := map[string]func(params any) (any, error){
		methodListGames: func(params any) (any, error) {
			time.Sleep(5 * time.Millisecond)
			return pagedFeed(100000)(params)
		},
	}
	lib := openTestLibrary(t, Options{Sources: feedSources(replies)})
	done := make(chan error, 1)
	if err := lib.StartImport(feedSource, nil, func(_ ImportResult, err error) { done <- err }); err != nil {
		t.Fatal(err)
	}
	lib.imports.stop()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("err = %v, want context.Canceled", err)
		}
	default:
		t.Fatal("stop must wait for the running import to report done")
	}
}
