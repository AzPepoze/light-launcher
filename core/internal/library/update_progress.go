package library

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

// UpdateCheckChannel is the event name for check progress.
const UpdateCheckChannel = "library.updateCheck"

// The state of an UpdateCheckEvent.
const (
	UpdateCheckRunning = "running"
	UpdateCheckDone    = "done"
)

// OutcomeSkipped is the result of a game left unchecked because its source needs a sign-in.
const OutcomeSkipped = "skipped"

// IssueAuth is the kind of an UpdateCheckIssue: the source needs a sign-in or an API key.
const IssueAuth = "auth"

// UpdateCheckEvent is one progress push of a check run.
type UpdateCheckEvent struct {
	State   string              `json:"state"`
	Done    int                 `json:"done"`
	Total   int                 `json:"total"`
	Skipped int                 `json:"skipped"`
	Current *UpdateCheckCurrent `json:"current"`
	Result  *UpdateCheckOutcome `json:"result"`
	Issues  []UpdateCheckIssue  `json:"issues"`
}

// UpdateCheckCurrent is the game a check run is working on.
type UpdateCheckCurrent struct {
	GameID int64  `json:"gameId"`
	Title  string `json:"title"`
}

// UpdateCheckOutcome is the finished check of one game.
type UpdateCheckOutcome struct {
	GameID    int64   `json:"gameId"`
	Title     string  `json:"title"`
	From      *string `json:"from"`
	To        *string `json:"to"`
	Result    string  `json:"result"`
	Error     *string `json:"error"`
	CoverPath *string `json:"coverPath"`
	CoverURL  *string `json:"coverUrl"`
}

// UpdateCheckIssue is one source that could not be checked in a run. There is one per source.
type UpdateCheckIssue struct {
	Source  string `json:"source"`
	Label   string `json:"label"`
	Kind    string `json:"kind"`
	Message string `json:"message"`
	Games   int    `json:"games"`
}

// sourceGate lets the first check of a source run alone. Its answer decides whether the other games of the source are checked.
type sourceGate struct {
	probe   sync.Mutex
	settled atomic.Bool
	blocked atomic.Bool
}

// checkRun is one update check: its counts, its sign-in gates and the events it pushes.
// Events are sent under the lock so they arrive in order.
type checkRun struct {
	mu       sync.Mutex
	push     func(UpdateCheckEvent)
	total    int
	done     int
	skipped  int
	running  []Game
	gates    map[string]*sourceGate
	issues   []UpdateCheckIssue
	finished chan struct{}
	results  []CheckResult
}

func newCheckRun(total int, push func(UpdateCheckEvent)) *checkRun {
	return &checkRun{
		push:     push,
		total:    total,
		gates:    map[string]*sourceGate{},
		finished: make(chan struct{}),
	}
}

// gate returns the sign-in gate of a source, made on first use.
func (r *checkRun) gate(source string) *sourceGate {
	r.mu.Lock()
	defer r.mu.Unlock()
	gate, ok := r.gates[source]
	if !ok {
		gate = &sourceGate{}
		r.gates[source] = gate
	}
	return gate
}

// start pushes the first event, before any game is checked.
func (r *checkRun) start() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.push(r.eventLocked(UpdateCheckRunning))
}

// begin marks a game as being checked. It pushes nothing.
func (r *checkRun) begin(game Game) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.running = append(r.running, game)
}

// finish records a checked game and pushes its outcome.
func (r *checkRun) finish(game Game, result CheckResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.finishLocked(game, result, outcomeKind(result))
}

// signInNeeded records that a source asked for a sign-in. The game that asked counts as unchecked.
func (r *checkRun) signInNeeded(source, label string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.issues = append(r.issues, UpdateCheckIssue{
		Source:  source,
		Label:   label,
		Kind:    IssueAuth,
		Message: fmt.Sprintf("%s needs a sign-in before it can be checked", label),
		Games:   1,
	})
}

// skipSignIn leaves out a game whose source needs a sign-in, and pushes its outcome.
func (r *checkRun) skipSignIn(game Game) CheckResult {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := noUpdate(game)
	r.skipped++
	if issue := r.issueLocked(game.Source); issue != nil {
		issue.Games++
	}
	r.finishLocked(game, result, OutcomeSkipped)
	return result
}

// finishLocked counts a game as done and pushes its outcome. The caller holds the lock.
func (r *checkRun) finishLocked(game Game, result CheckResult, kind string) {
	r.running = withoutGame(r.running, game.ID)
	r.done++
	event := r.eventLocked(UpdateCheckRunning)
	event.Result = outcomeOf(game, result, kind)
	r.push(event)
}

// complete stores the results, pushes the last event and wakes the callers that joined the run.
func (r *checkRun) complete(results []CheckResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.results = results
	event := r.eventLocked(UpdateCheckDone)
	event.Issues = append([]UpdateCheckIssue{}, r.issues...)
	r.push(event)
	close(r.finished)
}

// wait blocks until the run ends and returns its results. A canceled ctx stops the wait early.
func (r *checkRun) wait(ctx context.Context) ([]CheckResult, error) {
	select {
	case <-r.finished:
		return r.results, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// progress returns how many games are finished and how many there are in all.
func (r *checkRun) progress() (done, total int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.done, r.total
}

// eventLocked builds an event from the counts. The caller holds the lock.
func (r *checkRun) eventLocked(state string) UpdateCheckEvent {
	return UpdateCheckEvent{
		State:   state,
		Done:    r.done,
		Total:   r.total,
		Skipped: r.skipped,
		Current: currentOf(r.running),
		Issues:  []UpdateCheckIssue{},
	}
}

// issueLocked returns the issue of a source, or nil. The caller holds the lock.
func (r *checkRun) issueLocked(source string) *UpdateCheckIssue {
	for i := range r.issues {
		if r.issues[i].Source == source {
			return &r.issues[i]
		}
	}
	return nil
}

// withoutGame returns the games without the one with the given id.
func withoutGame(games []Game, id int64) []Game {
	kept := games[:0]
	for _, game := range games {
		if game.ID != id {
			kept = append(kept, game)
		}
	}
	return kept
}

// currentOf names the game that started first among those still running.
func currentOf(running []Game) *UpdateCheckCurrent {
	if len(running) == 0 {
		return nil
	}
	return &UpdateCheckCurrent{GameID: running[0].ID, Title: running[0].Title}
}

func outcomeOf(game Game, result CheckResult, kind string) *UpdateCheckOutcome {
	return &UpdateCheckOutcome{
		GameID:    game.ID,
		Title:     game.Title,
		From:      result.From,
		To:        result.To,
		Result:    kind,
		Error:     stringOrNil(result.Error),
		CoverPath: game.CoverPath,
		CoverURL:  game.CoverURL,
	}
}

// outcomeKind names what a check found: an error, an update, or nothing new.
func outcomeKind(result CheckResult) string {
	switch {
	case result.Error != "":
		return ResultError
	case result.HasUpdate:
		return ResultUpdate
	default:
		return ResultCurrent
	}
}

// activeCheck holds the run in progress, if any, so a second check joins it.
type activeCheck struct {
	mu  sync.Mutex
	run *checkRun
}

// claim starts a run of total games. When a run is already going, it returns that run instead.
func (a *activeCheck) claim(total int, push func(UpdateCheckEvent)) (run *checkRun, running *checkRun) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.run != nil {
		return nil, a.run
	}
	a.run = newCheckRun(total, push)
	return a.run, nil
}

// release forgets the run once it has stopped checking.
func (a *activeCheck) release() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.run = nil
}

// current returns the run in progress, or nil.
func (a *activeCheck) current() *checkRun {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.run
}
