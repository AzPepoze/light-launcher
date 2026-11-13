package library

import "context"

// UpdateProgress is how far a running check has got.
type UpdateProgress struct {
	Done  int `json:"done"`
	Total int `json:"total"`
}

// UpdateStatus is what the Updates page reads without starting a check.
type UpdateStatus struct {
	LastCheckedAt *string         `json:"lastCheckedAt"`
	Running       bool            `json:"running"`
	Progress      *UpdateProgress `json:"progress"`
	Count         int             `json:"count"`
	// NextRunAt is when the next background check starts, or null when none is scheduled.
	NextRunAt *string `json:"nextRunAt"`
	Games     []Game  `json:"games"`
}

// UpdateStatus returns the last saved check, whether a check runs now, and the games with a newer version.
// It never starts a check.
func (l *Library) UpdateStatus(ctx context.Context) (UpdateStatus, error) {
	summary, err := l.UpdatesSummary(ctx)
	if err != nil {
		return UpdateStatus{}, err
	}
	games, err := l.updateGames(ctx)
	if err != nil {
		return UpdateStatus{}, err
	}
	nextRun, err := l.nextRunAt(ctx, l.now())
	if err != nil {
		return UpdateStatus{}, err
	}
	status := UpdateStatus{Count: summary.Unseen, NextRunAt: nextRun, Games: games}
	if summary.Last != nil {
		status.LastCheckedAt = &summary.Last.CheckedAt
	}
	if run := l.activeCheck.current(); run != nil {
		done, total := run.progress()
		status.Running = true
		status.Progress = &UpdateProgress{Done: done, Total: total}
	}
	return status, nil
}

// updateGames reads every game with a newer version, one page at a time.
func (l *Library) updateGames(ctx context.Context) ([]Game, error) {
	games := []Game{}
	for offset := 0; ; offset += maxLimit {
		page, err := l.List(ctx, Filter{Updates: true, Sort: "added", Order: "desc", Limit: maxLimit, Offset: offset})
		if err != nil {
			return nil, err
		}
		games = append(games, page.Items...)
		if len(page.Items) < maxLimit {
			return games, nil
		}
	}
}
