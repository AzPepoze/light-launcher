package library

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"light-launcher/core/internal/contract"
)

// fakeClock is the library clock of a test. It only moves when the test advances it.
type fakeClock struct {
	now time.Time
}

func newFakeClock(lib *Library) *fakeClock {
	clock := &fakeClock{now: time.Now().Truncate(time.Second)}
	lib.now = func() time.Time { return clock.now }
	return clock
}

func (c *fakeClock) advance(d time.Duration) {
	c.now = c.now.Add(d)
}

// setChecked stores a last check time for a game.
func setChecked(t *testing.T, lib *Library, id int64, at time.Time) {
	t.Helper()
	value := stamp(at)
	if _, err := lib.Patch(context.Background(), id, Patch{LastCheckedAt: &value}); err != nil {
		t.Fatal(err)
	}
}

// sampleItchGame adds a game of the itch source.
func sampleItchGame(t *testing.T, lib *Library, sourceID string) Game {
	t.Helper()
	id, err := lib.Upsert(context.Background(), contract.GameDTO{Source: "itch", SourceID: sourceID, URL: "https://itch/" + sourceID, Title: "Sample Game", Version: "1"})
	if err != nil {
		t.Fatal(err)
	}
	game, err := lib.Get(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return game
}

func targetIDs(games []Game) map[int64]bool {
	ids := map[int64]bool{}
	for _, game := range games {
		ids[game.ID] = true
	}
	return ids
}

func TestDueIsDecidedPerGame(t *testing.T) {
	lib := openTestLibrary(t, Options{UpdateRule: recheckEvery(24 * time.Hour)})
	clock := newFakeClock(lib)
	never := f95Game(t, lib, "1.0")
	recent := sampleItchGame(t, lib, "9")
	old := sampleItchGame(t, lib, "10")
	setChecked(t, lib, recent.ID, clock.now.Add(-3*time.Hour))
	setChecked(t, lib, old.ID, clock.now.Add(-25*time.Hour))

	targets, err := lib.checkTargets(context.Background(), nil, true)
	if err != nil {
		t.Fatal(err)
	}
	ids := targetIDs(targets)
	if !ids[never.ID] || ids[recent.ID] || !ids[old.ID] {
		t.Errorf("due games = %v, want never and old, not recent (checked 3 h ago)", ids)
	}
}

func TestDueUsesTheRuleOfEachSource(t *testing.T) {
	rule := func(source string) UpdateRule {
		if source == "itch" {
			return UpdateRule{Auto: true, Recheck: time.Hour}
		}
		return UpdateRule{Auto: true, Recheck: 24 * time.Hour}
	}
	lib := openTestLibrary(t, Options{UpdateRule: rule})
	clock := newFakeClock(lib)
	f95 := f95Game(t, lib, "1.0")
	itch := sampleItchGame(t, lib, "9")
	setChecked(t, lib, f95.ID, clock.now.Add(-3*time.Hour))
	setChecked(t, lib, itch.ID, clock.now.Add(-3*time.Hour))

	targets, err := lib.checkTargets(context.Background(), nil, true)
	if err != nil {
		t.Fatal(err)
	}
	ids := targetIDs(targets)
	if ids[f95.ID] || !ids[itch.ID] {
		t.Errorf("due games = %v, want only the itch game with a 1 h recheck", ids)
	}
}

func TestOffSourceIsNeverDueButManualCheckWorks(t *testing.T) {
	rule := func(string) UpdateRule { return UpdateRule{Auto: false, Recheck: time.Hour} }
	lib := openTestLibrary(t, Options{UpdateRule: rule})
	game := f95Game(t, lib, "1.0")

	scheduled, err := lib.checkTargets(context.Background(), nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(scheduled) != 0 {
		t.Errorf("scheduled targets = %d, want 0 for an off source", len(scheduled))
	}
	manual, err := lib.checkTargets(context.Background(), nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(manual) != 1 || manual[0].ID != game.ID {
		t.Errorf("manual targets = %+v, want the game", manual)
	}
	byID, err := lib.checkTargets(context.Background(), []int64{game.ID}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(byID) != 1 {
		t.Errorf("a game picked by id = %d targets, want 1 even for an off source", len(byID))
	}
}

func TestZeroRecheckDuesEveryGame(t *testing.T) {
	lib := openTestLibrary(t, Options{UpdateRule: recheckEvery(0)})
	clock := newFakeClock(lib)
	game := f95Game(t, lib, "1.0")
	setChecked(t, lib, game.ID, clock.now.Add(-time.Minute))

	targets, err := lib.checkTargets(context.Background(), nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 {
		t.Errorf("targets = %d, want the game due with recheck 0", len(targets))
	}
}

func TestDueAllowsOneTickOfLateness(t *testing.T) {
	lib := openTestLibrary(t, Options{UpdateRule: recheckEvery(24 * time.Hour)})
	clock := newFakeClock(lib)
	game := f95Game(t, lib, "1.0")
	setChecked(t, lib, game.ID, clock.now.Add(-(24*time.Hour - 10*time.Minute)))

	targets, err := lib.checkTargets(context.Background(), nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 {
		t.Errorf("targets = %d, want the game due within one scheduler tick", len(targets))
	}
}

func TestFailedCheckCountsAsChecked(t *testing.T) {
	sources := updateSources(func(any) (any, error) { return nil, errors.New("login expired") })
	lib := openTestLibrary(t, Options{Sources: sources})
	game := f95Game(t, lib, "1.0")

	if _, err := lib.CheckGame(context.Background(), game.ID); err != nil {
		t.Fatal(err)
	}
	stored, err := lib.Get(context.Background(), game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.LastCheckedAt == nil {
		t.Error("a failed check left last_checked_at empty, so a scheduled run would retry it every tick")
	}
}

func TestScheduledCheckFollowsTheInterval(t *testing.T) {
	var calls atomic.Int32
	sources := updateSources(func(any) (any, error) {
		calls.Add(1)
		return nil, nil
	})
	lib := openTestLibrary(t, Options{
		Sources:        sources,
		UpdateRule:     recheckEvery(24 * time.Hour),
		UpdateSchedule: func() UpdateSchedule { return UpdateSchedule{Enabled: true, Interval: 24 * time.Hour} },
	})
	clock := newFakeClock(lib)
	f95Game(t, lib, "1.0")
	ctx := context.Background()

	steps := []struct {
		name  string
		after time.Duration
		want  bool
	}{
		{"first tick starts the clock", 0, false},
		{"23 h later is too early", 23 * time.Hour, false},
		{"25 h later runs", 2 * time.Hour, true},
		{"1 h later is too early", time.Hour, false},
		{"24 h later runs again", 24 * time.Hour, true},
	}
	for _, step := range steps {
		clock.advance(step.after)
		ran, err := lib.ScheduledCheck(ctx, false)
		if err != nil {
			t.Fatal(err)
		}
		if ran != step.want {
			t.Fatalf("%s: ran = %v, want %v", step.name, ran, step.want)
		}
	}
	if calls.Load() != 2 {
		t.Errorf("source calls = %d, want 2 (one per run that had a due game)", calls.Load())
	}
}

func TestScheduledCheckSkipsGameCheckedByHand(t *testing.T) {
	var calls atomic.Int32
	sources := updateSources(func(any) (any, error) {
		calls.Add(1)
		return nil, nil
	})
	lib := openTestLibrary(t, Options{
		Sources:        sources,
		UpdateRule:     recheckEvery(24 * time.Hour),
		UpdateSchedule: func() UpdateSchedule { return UpdateSchedule{Enabled: true, OnStartup: true, Interval: time.Hour} },
	})
	clock := newFakeClock(lib)
	game := f95Game(t, lib, "1.0")
	setChecked(t, lib, game.ID, clock.now.Add(-3*time.Hour))

	ran, err := lib.ScheduledCheck(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}
	if !ran || calls.Load() != 0 {
		t.Errorf("ran = %v, source calls = %d; want the run to start and skip the game checked 3 h ago", ran, calls.Load())
	}
}

func TestScheduledCheckRespectsOffSwitches(t *testing.T) {
	tests := []struct {
		name      string
		schedule  UpdateSchedule
		startup   bool
		advance   time.Duration
		wantRunAt bool
	}{
		{"disabled never runs", UpdateSchedule{Enabled: false, OnStartup: true, Interval: time.Hour}, true, 0, false},
		{"no startup run when off", UpdateSchedule{Enabled: true, OnStartup: false, Interval: time.Hour}, true, 0, false},
		{"interval 0 never runs on a schedule", UpdateSchedule{Enabled: true, OnStartup: false, Interval: 0}, false, 1000 * time.Hour, false},
		{"startup still runs with interval 0", UpdateSchedule{Enabled: true, OnStartup: true, Interval: 0}, true, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lib := openTestLibrary(t, Options{
				UpdateRule:     recheckEvery(24 * time.Hour),
				UpdateSchedule: func() UpdateSchedule { return tt.schedule },
			})
			clock := newFakeClock(lib)
			f95Game(t, lib, "1.0")
			// The first tick starts the clock, so a later periodic tick is the one under test.
			if _, err := lib.ScheduledCheck(context.Background(), false); err != nil {
				t.Fatal(err)
			}
			clock.advance(tt.advance)
			ran, err := lib.ScheduledCheck(context.Background(), tt.startup)
			if err != nil {
				t.Fatal(err)
			}
			if ran != tt.wantRunAt {
				t.Errorf("ran = %v, want %v", ran, tt.wantRunAt)
			}
		})
	}
}

func TestNextRunAtIsEmptyWhenNothingIsScheduled(t *testing.T) {
	lib := openTestLibrary(t, Options{
		UpdateRule:     recheckEvery(24 * time.Hour),
		UpdateSchedule: func() UpdateSchedule { return UpdateSchedule{Enabled: true, Interval: 0} },
	})
	newFakeClock(lib)
	f95Game(t, lib, "1.0")

	status, err := lib.UpdateStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.NextRunAt != nil {
		t.Errorf("nextRunAt = %v, want nil with interval 0", *status.NextRunAt)
	}
}

func TestNextRunAtWaitsForTheLaterOfScheduleAndDueGame(t *testing.T) {
	schedule := UpdateSchedule{Enabled: true, Interval: 24 * time.Hour}
	lib := openTestLibrary(t, Options{
		UpdateRule:     recheckEvery(24 * time.Hour),
		UpdateSchedule: func() UpdateSchedule { return schedule },
	})
	clock := newFakeClock(lib)
	game := f95Game(t, lib, "1.0")
	setChecked(t, lib, game.ID, clock.now.Add(-25*time.Hour))

	status, err := lib.UpdateStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.NextRunAt == nil || *status.NextRunAt != stamp(clock.now.Add(24*time.Hour)) {
		t.Errorf("nextRunAt = %v, want now plus the interval (the game is already due)", status.NextRunAt)
	}

	schedule.Interval = time.Hour
	setChecked(t, lib, game.ID, clock.now)
	if _, err := lib.ScheduledCheck(context.Background(), false); err != nil {
		t.Fatal(err)
	}
	clock.advance(time.Minute)
	status, err = lib.UpdateStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	wantDue := clock.now.Add(-time.Minute).Add(24*time.Hour - UpdateTick)
	if status.NextRunAt == nil || *status.NextRunAt != stamp(wantDue) {
		t.Errorf("nextRunAt = %v, want %v (the game is due last, after the 1 h schedule)", status.NextRunAt, stamp(wantDue))
	}
}

func TestSourceIntervalDelaysDueGames(t *testing.T) {
	now := time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)
	last := stamp(now.Add(-2 * time.Hour))
	rule := UpdateRule{Auto: true, Recheck: time.Hour, Interval: 48 * time.Hour}
	if rule.due(&last, now) {
		t.Error("a game checked 2h ago is due again under a 48h source interval")
	}
	old := stamp(now.Add(-49 * time.Hour))
	if !rule.due(&old, now) {
		t.Error("a game checked 49h ago is due under a 48h source interval")
	}
}
