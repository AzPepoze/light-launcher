package library

import (
	"context"
	"time"
)

// lastScheduledRunKey is the meta key of the last background run that started.
const lastScheduledRunKey = "updates.lastScheduledRun"

// UpdateTick is how often the scheduler looks at the schedule. A game is due this much early,
// so a run that starts one tick late still checks it.
const UpdateTick = 15 * time.Minute

// UpdateRule is how background checks treat the games of one source.
type UpdateRule struct {
	// Auto is false when the games are never checked in the background. Manual checks still work.
	Auto bool
	// Recheck skips a game checked within this time. Zero means every game is due.
	Recheck time.Duration
	// Interval is the source's own check gap. The longer of Interval and Recheck applies.
	Interval time.Duration
}

// UpdateSchedule is when background runs start.
type UpdateSchedule struct {
	Enabled   bool
	OnStartup bool
	// Interval is the minimum gap between two scheduled runs. Zero means no scheduled runs.
	Interval time.Duration
}

// dueAt returns when a game last checked at lastChecked is due again. The zero time means due now.
// The bool is false when the source is never checked in the background.
func (r UpdateRule) dueAt(lastChecked *string) (time.Time, bool) {
	if !r.Auto {
		return time.Time{}, false
	}
	gap := r.Recheck
	if r.Interval > gap {
		gap = r.Interval
	}
	if gap <= 0 || lastChecked == nil {
		return time.Time{}, true
	}
	checked, err := time.Parse(time.RFC3339Nano, *lastChecked)
	if err != nil {
		return time.Time{}, true
	}
	return checked.Add(gap - UpdateTick), true
}

// due reports whether a game last checked at lastChecked is due at now.
func (r UpdateRule) due(lastChecked *string, now time.Time) bool {
	at, ok := r.dueAt(lastChecked)
	return ok && !now.Before(at)
}

// ruleLookup returns a function that gives the rule of a source. The host is asked once per source.
func (l *Library) ruleLookup() func(source string) UpdateRule {
	cache := map[string]UpdateRule{}
	return func(source string) UpdateRule {
		rule, ok := cache[source]
		if !ok {
			rule = l.updateRule(source)
			cache[source] = rule
		}
		return rule
	}
}

func (l *Library) updateRule(source string) UpdateRule {
	if l.opts.UpdateRule == nil {
		return UpdateRule{Auto: true}
	}
	return l.opts.UpdateRule(source)
}

func (l *Library) updateSchedule() UpdateSchedule {
	if l.opts.UpdateSchedule == nil {
		return UpdateSchedule{}
	}
	return l.opts.UpdateSchedule()
}

// ScheduledCheck runs a background check when the schedule allows one and reports whether it ran.
// The startup run happens once after launch, when OnStartup is on. Other runs wait for the interval.
func (l *Library) ScheduledCheck(ctx context.Context, startup bool) (bool, error) {
	schedule := l.updateSchedule()
	if !schedule.Enabled {
		return false, nil
	}
	now := l.now()
	last, found, err := l.lastScheduledRun(ctx)
	if err != nil {
		return false, err
	}
	switch {
	case startup:
		if !schedule.OnStartup {
			return false, nil
		}
	case !found:
		// The first tick only starts the clock, so the first scheduled run waits one interval.
		return false, l.saveScheduledRun(ctx, now)
	case schedule.Interval <= 0 || now.Sub(last) < schedule.Interval:
		return false, nil
	}
	if err := l.saveScheduledRun(ctx, now); err != nil {
		return false, err
	}
	_, err = l.CheckAll(ctx, nil, true)
	return true, err
}

// lastScheduledRun reads when the last background run started. The bool is false when none did.
func (l *Library) lastScheduledRun(ctx context.Context) (time.Time, bool, error) {
	raw, found, err := l.GetMeta(ctx, lastScheduledRunKey)
	if err != nil || !found {
		return time.Time{}, false, err
	}
	last, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return time.Time{}, false, nil
	}
	return last, true, nil
}

func (l *Library) saveScheduledRun(ctx context.Context, at time.Time) error {
	return l.SetMeta(ctx, lastScheduledRunKey, stamp(at))
}

// nextRunAt returns when the next background check starts, or nil when none is scheduled.
// A scheduled run that finds no due game checks nothing, so the answer is the later of the
// next scheduled run and the earliest due game.
func (l *Library) nextRunAt(ctx context.Context, now time.Time) (*string, error) {
	schedule := l.updateSchedule()
	if !schedule.Enabled || schedule.Interval <= 0 {
		return nil, nil
	}
	earliest, found, err := l.earliestDueGame(ctx, now)
	if err != nil || !found {
		return nil, err
	}
	base := now
	last, lastFound, err := l.lastScheduledRun(ctx)
	if err != nil {
		return nil, err
	}
	if lastFound {
		base = last
	}
	next := later(later(base.Add(schedule.Interval), earliest), now)
	text := stamp(next)
	return &text, nil
}

// later returns the time that comes last.
func later(a, b time.Time) time.Time {
	if b.After(a) {
		return b
	}
	return a
}

// earliestDueGame returns when the first visible game in a checked source becomes due.
func (l *Library) earliestDueGame(ctx context.Context, now time.Time) (time.Time, bool, error) {
	games, err := l.All(ctx)
	if err != nil {
		return time.Time{}, false, err
	}
	rule := l.ruleLookup()
	var earliest time.Time
	found := false
	for _, game := range games {
		if game.Hidden != 0 {
			continue
		}
		at, ok := rule(game.Source).dueAt(game.LastCheckedAt)
		if !ok {
			continue
		}
		at = later(at, now)
		if !found || at.Before(earliest) {
			earliest = at
			found = true
		}
	}
	return earliest, found, nil
}
