package librarysvc

import (
	"context"
	"time"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/library"
)

const (
	syncEventChannel    = "library.sync"
	changedEventChannel = "library.changed"

	syncPhaseRunning = "running"
	syncPhaseDone    = "done"
	syncPhaseError   = "error"

	// syncEventInterval is the shortest time between two progress events of one sync.
	syncEventInterval = 500 * time.Millisecond
)

type syncSourcePayload struct {
	Source string `json:"source"`
}

type syncSourceResult struct {
	Started bool `json:"started"`
}

// syncEvent is the payload of library.sync.
type syncEvent struct {
	Source   string `json:"source"`
	Phase    string `json:"phase"`
	Imported int    `json:"imported"`
	Skipped  int    `json:"skipped"`
	Failed   int    `json:"failed"`
	Message  string `json:"message,omitempty"`
}

// syncSource starts a background import of the games of one source and returns at once.
func (s *service) syncSource(_ context.Context, in syncSourcePayload) (syncSourceResult, error) {
	reporter := &syncReporter{events: s.events, source: in.Source}
	if err := s.lib.StartImport(in.Source, reporter.progress, reporter.done); err != nil {
		return syncSourceResult{}, err
	}
	return syncSourceResult{Started: true}, nil
}

// syncReporter tells the UI how a sync goes. Progress events are throttled; the final event always goes out.
type syncReporter struct {
	events api.Events
	source string
	last   time.Time
}

func (r *syncReporter) progress(result library.ImportProgress) {
	if time.Since(r.last) < syncEventInterval {
		return
	}
	r.last = time.Now()
	r.send(syncPhaseRunning, result, "")
}

func (r *syncReporter) done(result library.ImportResult, err error) {
	if err != nil {
		r.send(syncPhaseError, result, err.Error())
		return
	}
	r.send(syncPhaseDone, result, "")
}

// send emits the sync event and tells the UI the library changed, since a sync writes games.
func (r *syncReporter) send(phase string, result library.ImportResult, message string) {
	r.events.Emit(syncEventChannel, syncEvent{
		Source:   r.source,
		Phase:    phase,
		Imported: result.Imported,
		Skipped:  result.Skipped,
		Failed:   result.Failed,
		Message:  message,
	})
	r.events.Emit(changedEventChannel, map[string]any{})
}
