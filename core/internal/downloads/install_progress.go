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
	Title      string   `json:"title,omitempty"` // the game's title, when the library knows it
	State      string   `json:"state"`           // active, done or error
	Phase      string   `json:"phase"`           // extracting, unpacking, finalizing or done
	Done       int64    `json:"done"`
	Total      int64    `json:"total"`
	Progress   *float64 `json:"progress"`
	Speed      float64  `json:"speed"`
	ETA        *float64 `json:"eta"` // seconds left at the current speed; null while unknown
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
	running    *runningInstalls
	state      InstallProgress
	phaseStart time.Time
	lastSent   time.Time
}

func newInstallReporter(events api.Events, running *runningInstalls, downloadID, gameID int64) *installReporter {
	return &installReporter{
		events:  events,
		running: running,
		state:   InstallProgress{DownloadID: downloadID, GameID: gameID, State: "active"},
	}
}

// titled sets the game title that the next snapshot carries.
func (r *installReporter) titled(title string) {
	r.state.Title = title
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

// send records the current snapshot for ActiveInstalls and emits it, unless it is too soon after the last one.
func (r *installReporter) send(force bool) {
	now := time.Now()
	snapshot := r.snapshot(now)
	if r.running != nil {
		r.running.update(snapshot)
	}
	if r.events == nil || (!force && now.Sub(r.lastSent) < installTick) {
		return
	}
	r.lastSent = now
	r.events.Emit(EventInstall, snapshot)
}

func (r *installReporter) snapshot(now time.Time) InstallProgress {
	snapshot := r.state
	snapshot.Speed = r.speed(now)
	snapshot.Progress = fraction(snapshot.Done, snapshot.Total)
	snapshot.ETA = etaSeconds(snapshot.Done, snapshot.Total, snapshot.Speed)
	if r.state.State == "done" {
		one := 1.0
		snapshot.Progress = &one
	}
	return snapshot
}

// etaSeconds is the time left at the current speed, or nil while the total or the speed is unknown.
func etaSeconds(done, total int64, speed float64) *float64 {
	if total <= 0 || speed <= 0 || done >= total {
		return nil
	}
	eta := float64(total-done) / speed
	return &eta
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
