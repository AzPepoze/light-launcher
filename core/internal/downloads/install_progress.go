package downloads

import (
	"time"

	"light-launcher/core/internal/api"
)

const (
	// EventInstall carries unpacking progress for one install.
	EventInstall = "install-progress"
	// EventToast asks the UI to show a short message.
	EventToast = "toast"

	installTick    = 100 * time.Millisecond
	minSpeedWindow = 250 * time.Millisecond
)

// InstallProgress is pushed on EventInstall. Progress is nil while the total size is unknown.
type InstallProgress struct {
	DownloadID int64    `json:"downloadId"`
	GameID     int64    `json:"gameId"`
	State      string   `json:"state"` // active, done or error
	Phase      string   `json:"phase"` // extracting, unpacking, finalizing or done
	Done       int64    `json:"done"`
	Total      int64    `json:"total"`
	Progress   *float64 `json:"progress"`
	Speed      float64  `json:"speed"`
	Engine     string   `json:"engine,omitempty"`
	InstallDir string   `json:"installDir,omitempty"`
	ExePath    string   `json:"exePath,omitempty"`
	Error      string   `json:"error,omitempty"`
}

// Toast is a short message for the user.
type Toast struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

// installReporter sends install progress, throttled to about ten updates a second.
// It is used from one goroutine only.
type installReporter struct {
	events     api.Events
	state      InstallProgress
	phaseStart time.Time
	lastSent   time.Time
}

func newInstallReporter(events api.Events, downloadID, gameID int64) *installReporter {
	return &installReporter{
		events: events,
		state:  InstallProgress{DownloadID: downloadID, GameID: gameID, State: "active"},
	}
}

// phase starts a new phase; progress counts restart from zero.
func (r *installReporter) phase(name string) {
	r.state.Phase = name
	r.state.Done, r.state.Total = 0, 0
	r.phaseStart = time.Now()
	r.send(true)
}

func (r *installReporter) progress(done, total int64) {
	r.state.Done, r.state.Total = done, total
	r.send(false)
}

func (r *installReporter) engine(name string) {
	r.state.Engine = name
	r.send(true)
}

func (r *installReporter) done(res installResult) {
	r.state.State = "done"
	r.state.Phase = "done"
	r.state.InstallDir = res.dir
	r.state.ExePath = res.executable
	r.state.Done, r.state.Total = res.size, res.size
	r.send(true)
}

func (r *installReporter) fail(err error) {
	r.state.State = "error"
	r.state.Error = err.Error()
	r.send(true)
}

func (r *installReporter) send(force bool) {
	if r.events == nil {
		return
	}
	now := time.Now()
	if !force && now.Sub(r.lastSent) < installTick {
		return
	}
	r.lastSent = now
	snapshot := r.state
	snapshot.Speed = r.speed(now)
	snapshot.Progress = fraction(snapshot.Done, snapshot.Total)
	if r.state.State == "done" {
		one := 1.0
		snapshot.Progress = &one
	}
	r.events.Emit(EventInstall, snapshot)
}

// speed is the average rate in the current phase, once there is enough time to measure it.
func (r *installReporter) speed(now time.Time) float64 {
	elapsed := now.Sub(r.phaseStart)
	if elapsed < minSpeedWindow || r.state.Done == 0 {
		return 0
	}
	return float64(r.state.Done) / elapsed.Seconds()
}

// fraction is done/total, or nil when the total is unknown.
func fraction(done, total int64) *float64 {
	if total <= 0 {
		return nil
	}
	p := min(float64(done)/float64(total), 1)
	return &p
}
