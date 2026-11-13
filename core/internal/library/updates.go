package library

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/logger"
)

const (
	logCategory = "library"
	// notModified is the error text a plugin uses when the source answered "unchanged".
	notModified = "not-modified"
	// checkWorkers is how many games CheckAll checks at once.
	checkWorkers = 3
)

// CheckResult is the outcome of one update check.
type CheckResult struct {
	ID        int64   `json:"id"`
	Title     string  `json:"title"`
	From      *string `json:"from"`
	To        *string `json:"to"`
	HasUpdate bool    `json:"hasUpdate"`
	Error     string  `json:"error,omitempty"`
}

// lastRunKey is the meta key that keeps the last update check.
const lastRunKey = "lastUpdateCheck"

// UpdateRun is one update check: when it ran and what each game answered.
type UpdateRun struct {
	CheckedAt string        `json:"checkedAt"`
	Results   []CheckResult `json:"results"`
}

// UpdatesSummary is what the updates view shows before the user checks again.
type UpdatesSummary struct {
	Unseen int        `json:"unseen"`
	Last   *UpdateRun `json:"last"`
}

// UpdatesSummary returns the number of unseen updates and the last check, if there was one.
func (l *Library) UpdatesSummary(ctx context.Context) (UpdatesSummary, error) {
	unseen, err := l.UnseenCount(ctx)
	if err != nil {
		return UpdatesSummary{}, err
	}
	summary := UpdatesSummary{Unseen: unseen}
	raw, found, err := l.GetMeta(ctx, lastRunKey)
	if err != nil {
		return UpdatesSummary{}, err
	}
	if found {
		var run UpdateRun
		if err := json.Unmarshal([]byte(raw), &run); err == nil {
			summary.Last = &run
		}
	}
	return summary, nil
}

// CheckGame asks the game's source for a newer version and records what it finds.
func (l *Library) CheckGame(ctx context.Context, id int64) (CheckResult, error) {
	result, _, err := l.checkGame(ctx, id)
	return result, err
}

// checkGame is CheckGame. It also reports whether the source asked for a sign-in.
func (l *Library) checkGame(ctx context.Context, id int64) (CheckResult, bool, error) {
	game, err := l.Get(ctx, id)
	if err != nil {
		return CheckResult{}, false, err
	}
	info, ok := l.pluginFor(game.Source)
	if !ok || !hasCapability(info, CapabilityUpdateCheck) {
		return noUpdate(game), false, nil
	}

	checkedAt := stamp(l.now())
	var reply updateReply
	params := gameParams{Source: info.SourceID, Game: gameRefOf(game)}
	if err := l.sources.Call(ctx, info.Plugin, methodCheckUpdate, params, &reply); err != nil {
		result, handleErr := l.handleCheckError(ctx, game, checkedAt, err)
		return result, errors.Is(err, contract.ErrAuthRequired), handleErr
	}
	if _, err := l.Patch(ctx, id, Patch{LastCheckedAt: &checkedAt}); err != nil {
		return CheckResult{}, false, err
	}
	if reply.Update == nil {
		result := noUpdate(game)
		l.recordCheck(ctx, game, derefString(result.From), derefString(result.To), "", ResultCurrent, "")
		return result, false, nil
	}
	result, err := l.applyUpdate(ctx, game, *reply.Update)
	return result, false, err
}

func (l *Library) handleCheckError(ctx context.Context, game Game, checkedAt string, err error) (CheckResult, error) {
	if strings.Contains(err.Error(), notModified) {
		if _, patchErr := l.Patch(ctx, game.ID, Patch{LastCheckedAt: &checkedAt}); patchErr != nil {
			return CheckResult{}, patchErr
		}
		result := noUpdate(game)
		l.recordCheck(ctx, game, derefString(result.From), derefString(result.To), "", ResultCurrent, "")
		return result, nil
	}
	logger.Warn(logCategory, "update check failed for %s: %v", game.Title, err)
	// A failed check still counts as a check, so a scheduled run does not retry it every tick.
	if _, patchErr := l.Patch(ctx, game.ID, Patch{LastCheckedAt: &checkedAt}); patchErr != nil {
		return CheckResult{}, patchErr
	}
	result := noUpdate(game)
	result.Error = err.Error()
	l.recordCheck(ctx, game, derefString(result.From), "", "", ResultError, err.Error())
	return result, nil
}

// recordCheck keeps a check in the history. A failed write is logged only: it must not fail the check.
func (l *Library) recordCheck(ctx context.Context, game Game, from, to, changelog, result, failure string) {
	if err := l.AddCheckResult(ctx, game.ID, from, to, changelog, result, failure); err != nil {
		logger.Warn(logCategory, "record update check for %s: %v", game.Title, err)
	}
}

// applyUpdate compares the source answer with the stored version and saves a change when there is one.
func (l *Library) applyUpdate(ctx context.Context, game Game, info UpdateInfo) (CheckResult, error) {
	current := derefString(game.Version)
	latest := firstNonEmpty(info.Version, current)
	newerVersion := isNewer(latest, current)
	newerEdit := info.UpdatedAt != "" && info.UpdatedAt != derefString(game.SourceUpdatedAt)
	hasUpdate := newerVersion || (newerEdit && latest != "")

	result := CheckResult{ID: game.ID, Title: game.Title, From: game.Version, To: stringOrNil(latest), HasUpdate: hasUpdate}
	if !hasUpdate {
		l.recordCheck(ctx, game, current, latest, info.Changelog, ResultCurrent, "")
		return result, nil
	}
	if err := l.AddCheckResult(ctx, game.ID, current, latest, info.Changelog, ResultUpdate, ""); err != nil {
		return result, err
	}
	patch := Patch{}
	if info.Version != "" {
		patch.Version = &info.Version
	}
	if info.UpdatedAt != "" {
		patch.SourceUpdatedAt = &info.UpdatedAt
	}
	if _, err := l.Patch(ctx, game.ID, patch); err != nil {
		return result, err
	}
	l.notifyUpdates(ctx, game.ID)
	l.autoInstall(ctx, game, latest)
	return result, nil
}

func (l *Library) notifyUpdates(ctx context.Context, gameID int64) {
	count, err := l.UnseenCount(ctx)
	if err != nil {
		logger.Warn(logCategory, "count updates: %v", err)
	}
	l.notify("library.updates", map[string]any{"count": count})
	l.notify("library.game", map[string]any{"id": gameID})
}

// autoInstall hands the update to the installer when the host asks for it for this game.
func (l *Library) autoInstall(ctx context.Context, game Game, version string) {
	if l.opts.Installer == nil || l.opts.AutoUpdate == nil || !l.opts.AutoUpdate(game) {
		return
	}
	if err := l.opts.Installer.InstallUpdate(ctx, game, version); err != nil {
		logger.Warn(logCategory, "auto update failed for %s: %v", game.Title, err)
	}
}

// CheckAll checks the given games, or every visible game when ids is empty.
// Given ids are always checked. With skipRecent and no ids, only the games that are due are checked.
// A check that is already running is joined: its results are returned and nothing is checked twice.
// Progress is pushed on UpdateCheckChannel as games finish.
func (l *Library) CheckAll(ctx context.Context, ids []int64, skipRecent bool) ([]CheckResult, error) {
	targets, err := l.checkTargets(ctx, ids, skipRecent)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, nil
	}
	run, running := l.activeCheck.claim(len(targets), l.pushUpdateCheck)
	if running != nil {
		return running.wait(ctx)
	}
	run.start()
	results := l.checkGames(ctx, targets, run)
	l.saveRun(ctx, results)
	l.activeCheck.release()
	run.complete(results)
	count, err := l.UnseenCount(ctx)
	if err == nil {
		l.notify("library.updates", map[string]any{"count": count})
	}
	return results, nil
}

func (l *Library) pushUpdateCheck(event UpdateCheckEvent) {
	l.notify(UpdateCheckChannel, event)
}

// checkGames checks the games with a pool of checkWorkers. The results keep the order of games.
func (l *Library) checkGames(ctx context.Context, games []Game, run *checkRun) []CheckResult {
	results := make([]CheckResult, len(games))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range min(checkWorkers, len(games)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				results[i] = l.checkOne(ctx, games[i], run)
			}
		}()
	}
	for i := range games {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return results
}

// checkOne checks one game. When its source needs a sign-in, the other games of that source are skipped.
func (l *Library) checkOne(ctx context.Context, game Game, run *checkRun) CheckResult {
	gate := run.gate(game.Source)
	if !gate.settled.Load() {
		if result, probed := l.probe(ctx, game, run, gate); probed {
			return result
		}
	}
	if gate.blocked.Load() {
		return run.skipSignIn(game)
	}
	return l.checkNow(ctx, game, run)
}

// probe runs the first check of a source while the other games of it wait. It returns false when
// another game already got the answer, so the caller goes on with that answer.
func (l *Library) probe(ctx context.Context, game Game, run *checkRun, gate *sourceGate) (CheckResult, bool) {
	gate.probe.Lock()
	defer gate.probe.Unlock()
	if gate.settled.Load() {
		return CheckResult{}, false
	}
	run.begin(game)
	result, needsSignIn := l.checkResult(ctx, game)
	if needsSignIn {
		gate.blocked.Store(true)
		run.signInNeeded(game.Source, l.sourceLabel(game.Source))
	}
	gate.settled.Store(true)
	run.finish(game, result)
	return result, true
}

// checkNow checks a game whose source has already answered.
func (l *Library) checkNow(ctx context.Context, game Game, run *checkRun) CheckResult {
	run.begin(game)
	result, _ := l.checkResult(ctx, game)
	run.finish(game, result)
	return result
}

// checkResult runs CheckGame for a game. A failure becomes the result's error, so the run goes on.
func (l *Library) checkResult(ctx context.Context, game Game) (CheckResult, bool) {
	result, needsSignIn, err := l.checkGame(ctx, game.ID)
	if err != nil {
		return CheckResult{ID: game.ID, Title: game.Title, From: game.Version, To: game.Version, Error: err.Error()}, false
	}
	return result, needsSignIn
}

// sourceLabel returns the name a source shows in the UI, or its id when it has none.
func (l *Library) sourceLabel(source string) string {
	if info, ok := l.pluginFor(source); ok && info.Label != "" {
		return info.Label
	}
	return source
}

// saveRun keeps the results of a check for the updates view. A failed save is logged only.
func (l *Library) saveRun(ctx context.Context, results []CheckResult) {
	run := UpdateRun{CheckedAt: nowStamp(), Results: results}
	if err := l.setMetaJSON(ctx, lastRunKey, run); err != nil {
		logger.Warn(logCategory, "could not save the update check: %v", err)
	}
}

func (l *Library) checkTargets(ctx context.Context, ids []int64, skipRecent bool) ([]Game, error) {
	if len(ids) > 0 {
		return l.gamesByID(ctx, ids), nil
	}
	all, err := l.All(ctx)
	if err != nil {
		return nil, err
	}
	var visible []Game
	for _, game := range all {
		// Local games have no source to ask; they are never part of a check run.
		if game.Hidden == 0 && game.SourceKind != "local" {
			visible = append(visible, game)
		}
	}
	if !skipRecent {
		return visible, nil
	}
	return dueForCheck(visible, l.ruleLookup(), l.now()), nil
}

// gamesByID returns the games that exist. Unknown ids are skipped.
func (l *Library) gamesByID(ctx context.Context, ids []int64) []Game {
	var games []Game
	for _, id := range ids {
		game, err := l.Get(ctx, id)
		if err == nil {
			games = append(games, game)
		}
	}
	return games
}

// dueForCheck keeps the games that are due at now, each under the rule of its own source.
func dueForCheck(games []Game, rule func(source string) UpdateRule, now time.Time) []Game {
	var due []Game
	for _, game := range games {
		if rule(game.Source).due(game.LastCheckedAt, now) {
			due = append(due, game)
		}
	}
	return due
}

func noUpdate(game Game) CheckResult {
	return CheckResult{ID: game.ID, Title: game.Title, From: game.Version, To: game.Version}
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func stringOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
