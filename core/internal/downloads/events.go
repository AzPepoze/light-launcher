package downloads

import (
	"context"
	"time"
)

// EventProgress is the channel that carries download state and progress to the UI.
const EventProgress = "download-progress"

// ProgressEvent is pushed on EventProgress whenever a state changes and about four times a second while running.
type ProgressEvent struct {
	ID       int64   `json:"id"`
	State    State   `json:"state"`
	Received int64   `json:"received"`
	Total    int64   `json:"total"`
	Speed    float64 `json:"speed"`
	Error    string  `json:"error,omitempty"`
}

// View is one download as the UI lists it. Request headers are left out on purpose.
type View struct {
	ID          int64   `json:"id"`
	GameID      int64   `json:"gameId"`
	URL         string  `json:"url"`
	Filename    string  `json:"filename"`
	Dest        string  `json:"dest"`
	State       State   `json:"state"`
	Error       string  `json:"error,omitempty"`
	Received    int64   `json:"received"`
	Total       int64   `json:"total"`
	Speed       float64 `json:"speed"`
	Connections int     `json:"connections"`
	Install     bool    `json:"install"`
	// RateLimited is true while a total speed limit applies to a running download.
	RateLimited bool       `json:"rateLimited"`
	CreatedAt   time.Time  `json:"createdAt"`
	FinishedAt  *time.Time `json:"finishedAt,omitempty"`
}

// viewLocked builds the view of one row, using live progress when the download is running.
func (m *Manager) viewLocked(d Download, now time.Time) View {
	v := View{
		ID:          d.ID,
		GameID:      d.GameID,
		URL:         d.URL,
		Filename:    d.Filename,
		Dest:        d.Dest,
		State:       d.State,
		Error:       d.Error,
		Received:    d.BytesDone,
		Total:       d.Size,
		Connections: d.Connections,
		Install:     d.Install,
		CreatedAt:   d.CreatedAt,
	}
	if !d.FinishedAt.IsZero() {
		finished := d.FinishedAt
		v.FinishedAt = &finished
	}
	v.RateLimited = d.State == StateDownloading && m.settings.SpeedLimitKBps > 0
	if j := m.active[d.ID]; j != nil {
		p := j.progress(now)
		v.Received, v.Total, v.Speed = p.received, p.total, p.speed
		if p.connections > 0 {
			v.Connections = p.connections
		}
	}
	return v
}

// publishStateLocked sends the current state of a download to the UI.
func (m *Manager) publishStateLocked(ctx context.Context, id int64) {
	d, err := m.store.Get(ctx, id)
	if err != nil {
		return
	}
	v := m.viewLocked(d, time.Now())
	m.publish(ProgressEvent{
		ID:       id,
		State:    d.State,
		Received: v.Received,
		Total:    v.Total,
		Speed:    v.Speed,
		Error:    d.Error,
	})
}

func (m *Manager) publish(e ProgressEvent) {
	if m.events != nil {
		m.events.Emit(EventProgress, e)
	}
}
