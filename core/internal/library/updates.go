package library

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"light-launcher/core/internal/logger"
)

const (
	logCategory = "library"
	// notModified is the error text a plugin uses when the source answered "unchanged".
	notModified = "not-modified"
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
	game, err := l.Get(ctx, id)
	if err != nil {
		return CheckResult{}, err
	}
	info, ok := l.pluginFor(game.Source)
	if !ok || !hasCapability(info, CapabilityUpdateCheck) {
		return noUpdate(game), nil
	}

	stamp := nowStamp()
	var reply updateReply
	params := gameParams{Source: info.SourceID, Game: gameRefOf(game)}
	if err := l.sources.Call(ctx, info.Plugin, methodCheckUpdate, params, &reply); err != nil {
		return l.handleCheckError(ctx, game, stamp, err)
	}
	if _, err := l.Patch(ctx, id, Patch{LastCheckedAt: &stamp}); err != nil {
		return CheckResult{}, err
	}
	if reply.Update == nil {
		return noUpdate(game), nil
	}
	return l.applyUpdate(ctx, game, *reply.Update)
}

func (l *Library) handleCheckError(ctx context.Context, game Game, stamp string, err error) (CheckResult, error) {
	if strings.Contains(err.Error(), notModified) {
		if _, patchErr := l.Patch(ctx, game.ID, Patch{LastCheckedAt: &stamp}); patchErr != nil {
			return CheckResult{}, patchErr
		}
		return noUpdate(game), nil
	}
	logger.Warn(logCategory, "update check failed for %s: %v", game.Title, err)
	result := noUpdate(game)
	result.Error = err.Error()
	return result, nil
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
		return result, nil
	}
	if err := l.AddCheck(ctx, game.ID, current, latest, info.Changelog); err != nil {
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
// With skipRecent, games checked within UpdateRecheck are left out.
func (l *Library) CheckAll(ctx context.Context, ids []int64, skipRecent bool) ([]CheckResult, error) {
	targets, err := l.checkTargets(ctx, ids, skipRecent)
	if err != nil {
		return nil, err
	}
	results := make([]CheckResult, 0, len(targets))
	for _, game := range targets {
		result, err := l.CheckGame(ctx, game.ID)
		if err != nil {
			result = CheckResult{ID: game.ID, Title: game.Title, From: game.Version, To: game.Version, Error: err.Error()}
		}
		results = append(results, result)
	}
	l.saveRun(ctx, results)
	count, err := l.UnseenCount(ctx)
	if err == nil {
		l.notify("library.updates", map[string]any{"count": count})
	}
	return results, nil
}

// saveRun keeps the results of a check for the updates view. A failed save is logged only.
func (l *Library) saveRun(ctx context.Context, results []CheckResult) {
	run := UpdateRun{CheckedAt: nowStamp(), Results: results}
	if err := l.setMetaJSON(ctx, lastRunKey, run); err != nil {
		logger.Warn(logCategory, "could not save the update check: %v", err)
	}
}

func (l *Library) checkTargets(ctx context.Context, ids []int64, skipRecent bool) ([]Game, error) {
	var candidates []Game
	if len(ids) > 0 {
		for _, id := range ids {
			game, err := l.Get(ctx, id)
			if err == nil {
				candidates = append(candidates, game)
			}
		}
	} else {
		all, err := l.All(ctx)
		if err != nil {
			return nil, err
		}
		for _, game := range all {
			if game.Hidden == 0 {
				candidates = append(candidates, game)
			}
		}
	}
	if !skipRecent {
		return candidates, nil
	}
	return dueForCheck(candidates, sinceOrZero(time.Now(), l.opts.UpdateRecheck)), nil
}

// dueForCheck keeps games never checked, or last checked before the cutoff.
func dueForCheck(games []Game, cutoff time.Time) []Game {
	if cutoff.IsZero() {
		return games
	}
	var due []Game
	for _, game := range games {
		if game.LastCheckedAt == nil {
			due = append(due, game)
			continue
		}
		checked, err := time.Parse(time.RFC3339Nano, *game.LastCheckedAt)
		if err != nil || checked.Before(cutoff) {
			due = append(due, game)
		}
	}
	return due
}

// sinceOrZero returns the cutoff before which a check is stale. A zero recheck means every game is due.
func sinceOrZero(now time.Time, recheck time.Duration) time.Time {
	if recheck <= 0 {
		return time.Time{}
	}
	return now.Add(-recheck)
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
