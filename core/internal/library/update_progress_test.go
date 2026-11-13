package library

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"light-launcher/core/internal/contract"
)

// eventLog collects the events a check run pushes. It is safe for the pool's goroutines.
type eventLog struct {
	mu     sync.Mutex
	events []UpdateCheckEvent
}

func (e *eventLog) notify(channel string, data any) {
	if channel != UpdateCheckChannel {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.events = append(e.events, data.(UpdateCheckEvent))
}

func (e *eventLog) all() []UpdateCheckEvent {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]UpdateCheckEvent{}, e.events...)
}

// f95Games adds count games of the f95 source, each with version 1.0.
func f95Games(t *testing.T, lib *Library, count int) []Game {
	t.Helper()
	var games []Game
	for i := range count {
		id, err := lib.Upsert(context.Background(), contract.GameDTO{
			Source: "f95", SourceID: string(rune('a' + i)), URL: "https://f95/" + string(rune('a'+i)),
			Title: "Harbor " + string(rune('a'+i)), Version: "1.0",
		})
		if err != nil {
			t.Fatal(err)
		}
		game, err := lib.Get(context.Background(), id)
		if err != nil {
			t.Fatal(err)
		}
		games = append(games, game)
	}
	return games
}

func newlyUpdated(any) (any, error) {
	return map[string]any{"update": map[string]string{"version": "2.0"}}, nil
}

func TestCheckAllPushesProgressInOrder(t *testing.T) {
	var log eventLog
	lib := openTestLibrary(t, Options{Sources: updateSources(newlyUpdated), Notify: log.notify})
	f95Games(t, lib, 4)

	if _, err := lib.CheckAll(context.Background(), nil, false); err != nil {
		t.Fatal(err)
	}
	events := log.all()
	if len(events) != 6 {
		t.Fatalf("events = %d, want start, four results and done", len(events))
	}
	start := events[0]
	if start.State != UpdateCheckRunning || start.Done != 0 || start.Total != 4 || start.Result != nil {
		t.Errorf("start = %+v", start)
	}
	for i, event := range events[1:5] {
		if event.State != UpdateCheckRunning || event.Done != i+1 || event.Total != 4 {
			t.Errorf("event %d = %+v, want running with done %d", i+1, event, i+1)
		}
		if event.Result == nil || event.Result.Result != ResultUpdate || event.Result.To == nil {
			t.Errorf("event %d result = %+v, want an update", i+1, event.Result)
		}
	}
	end := events[5]
	if end.State != UpdateCheckDone || end.Done != 4 || end.Total != 4 || end.Current != nil {
		t.Errorf("end = %+v", end)
	}
}

func TestCheckAllKeepsThreeGamesInFlight(t *testing.T) {
	var inFlight, peak atomic.Int32
	sources := updateSources(func(any) (any, error) {
		now := inFlight.Add(1)
		defer inFlight.Add(-1)
		for {
			old := peak.Load()
			if now <= old || peak.CompareAndSwap(old, now) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		return map[string]any{"update": map[string]string{"version": "2.0"}}, nil
	})
	lib := openTestLibrary(t, Options{Sources: sources})
	games := f95Games(t, lib, 9)

	ctx := context.Background()
	targets, err := lib.checkTargets(ctx, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	results, err := lib.CheckAll(ctx, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if peak.Load() > checkWorkers || peak.Load() < 2 {
		t.Errorf("peak in flight = %d, want 2 to %d", peak.Load(), checkWorkers)
	}
	if len(results) != len(games) || len(targets) != len(games) {
		t.Fatalf("results = %d, want %d", len(results), len(games))
	}
	for i, result := range results {
		if result.ID != targets[i].ID {
			t.Errorf("result %d is game %d, want %d: order must follow the games", i, result.ID, targets[i].ID)
		}
	}
}

func TestCheckAllReportsErrorsWithoutStopping(t *testing.T) {
	var log eventLog
	sources := updateSources(func(any) (any, error) { return nil, errors.New("login expired") })
	lib := openTestLibrary(t, Options{Sources: sources, Notify: log.notify})
	f95Games(t, lib, 3)

	results, err := lib.CheckAll(context.Background(), nil, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range results {
		if result.Error != "login expired" {
			t.Errorf("result = %+v, want the error", result)
		}
	}
	events := log.all()
	last := events[len(events)-2]
	if last.Done != 3 || last.Result == nil || last.Result.Result != ResultError || last.Result.Error == nil {
		t.Errorf("last result event = %+v", last)
	}
}

func TestCheckAllSkipsSourceThatNeedsSignIn(t *testing.T) {
	var log eventLog
	var itchCalls, f95Calls atomic.Int32
	sources := &fakeSources{
		infos: []contract.SourceInfo{
			{Plugin: "itchplugin", SourceID: "itch", Label: "itch.io", Capabilities: []string{CapabilityUpdateCheck}},
			{Plugin: "f95plugin", SourceID: "f95", Label: "F95zone", Capabilities: []string{CapabilityUpdateCheck}},
		},
		replies: map[string]func(any) (any, error){
			methodCheckUpdate: func(params any) (any, error) {
				if params.(gameParams).Source == "itch" {
					itchCalls.Add(1)
					return nil, contract.ErrAuthRequired
				}
				f95Calls.Add(1)
				return map[string]any{"update": map[string]string{"version": "2.0"}}, nil
			},
		},
	}
	lib := openTestLibrary(t, Options{Sources: sources, Notify: log.notify})
	ctx := context.Background()
	for i := range 3 {
		if _, err := lib.Upsert(ctx, contract.GameDTO{Source: "itch", SourceID: string(rune('a' + i)), URL: "https://itch/" + string(rune('a'+i)), Title: "Sample", Version: "1.0"}); err != nil {
			t.Fatal(err)
		}
	}
	f95Games(t, lib, 2)

	results, err := lib.CheckAll(ctx, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if itchCalls.Load() != 1 {
		t.Errorf("itch calls = %d, want 1: only the first game asks", itchCalls.Load())
	}
	if f95Calls.Load() != 2 {
		t.Errorf("f95 calls = %d, want 2: other sources still checked", f95Calls.Load())
	}
	if len(results) != 5 {
		t.Fatalf("results = %d, want one per game", len(results))
	}

	history, err := lib.UpdateHistory(ctx, 0, 50, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 3 {
		t.Errorf("history = %d rows, want 3: one itch failure and two f95 checks", len(history))
	}

	events := log.all()
	end := events[len(events)-1]
	if end.State != UpdateCheckDone || end.Done != 5 || end.Total != 5 || end.Skipped != 2 {
		t.Fatalf("end = %+v, want done 5 of 5 with 2 skipped", end)
	}
	if len(end.Issues) != 1 {
		t.Fatalf("issues = %+v, want one per source", end.Issues)
	}
	issue := end.Issues[0]
	if issue.Source != "itch" || issue.Label != "itch.io" || issue.Kind != IssueAuth || issue.Games != 3 || issue.Message == "" {
		t.Errorf("issue = %+v", issue)
	}
}

func TestUpdateStatusReadsWithoutChecking(t *testing.T) {
	calls := atomic.Int32{}
	sources := updateSources(func(any) (any, error) {
		calls.Add(1)
		return map[string]any{"update": map[string]string{"version": "2.0"}}, nil
	})
	lib := openTestLibrary(t, Options{Sources: sources})
	ctx := context.Background()
	game := f95Games(t, lib, 1)[0]
	// Only installed games take part in the updates filter.
	installed := "1.0"
	if _, err := lib.Patch(ctx, game.ID, Patch{InstallDir: strPtr("/games/harbor"), InstalledVersion: &installed}); err != nil {
		t.Fatal(err)
	}

	before, err := lib.UpdateStatus(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if before.LastCheckedAt != nil || before.Running || before.Count != 0 || len(before.Games) != 0 {
		t.Errorf("status before any check = %+v", before)
	}

	if _, err := lib.CheckAll(ctx, nil, false); err != nil {
		t.Fatal(err)
	}
	status, err := lib.UpdateStatus(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Errorf("source calls = %d, want only the check's 1", calls.Load())
	}
	if status.LastCheckedAt == nil || status.Running || status.Count != 1 || len(status.Games) != 1 {
		t.Errorf("status = %+v, want a saved run and one game with an update", status)
	}
}

func TestSecondCheckJoinsRunningCheck(t *testing.T) {
	release := make(chan struct{})
	var calls atomic.Int32
	sources := updateSources(func(any) (any, error) {
		calls.Add(1)
		<-release
		return map[string]any{"update": map[string]string{"version": "2.0"}}, nil
	})
	lib := openTestLibrary(t, Options{Sources: sources})
	ctx := context.Background()
	f95Games(t, lib, 1)

	first := make(chan []CheckResult, 1)
	go func() {
		results, _ := lib.CheckAll(ctx, nil, false)
		first <- results
	}()
	for lib.activeCheck.current() == nil {
		time.Sleep(time.Millisecond)
	}
	status, err := lib.UpdateStatus(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Running || status.Progress == nil || status.Progress.Total != 1 {
		t.Fatalf("status during run = %+v, want running with 1 game", status)
	}

	second := make(chan []CheckResult, 1)
	go func() {
		results, _ := lib.CheckAll(ctx, nil, false)
		second <- results
	}()
	time.Sleep(50 * time.Millisecond)
	close(release)

	if got := len(<-first); got != 1 {
		t.Errorf("first results = %d, want 1", got)
	}
	if got := len(<-second); got != 1 {
		t.Errorf("joined results = %d, want the running check's 1", got)
	}
	if calls.Load() != 1 {
		t.Errorf("source calls = %d, want 1: the second check must not run again", calls.Load())
	}
}
