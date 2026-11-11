package librarysvc

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/library"
)

// listingSources offers one source with listGames and serves it in two pages of one game each.
type listingSources struct{}

func (listingSources) Sources() []contract.SourceInfo {
	return []contract.SourceInfo{{Plugin: "feedplugin", SourceID: "feed", Label: "Feed",
		Capabilities: []string{library.CapabilityListGames}}}
}

func (listingSources) Call(_ context.Context, _, method string, params any, out any) error {
	if method != "source.listGames" {
		return errors.New("unexpected method " + method)
	}
	raw, err := json.Marshal(params)
	if err != nil {
		return err
	}
	var request struct {
		Cursor string `json:"cursor"`
	}
	if err := json.Unmarshal(raw, &request); err != nil {
		return err
	}
	reply := map[string]any{"games": []contract.GameDTO{{Source: "feed", SourceID: "1", URL: "https://feed/1", Title: "First", UpdatedAt: "2026-01-01T00:00:00.000Z"}}, "nextCursor": "2"}
	if request.Cursor != "" {
		reply = map[string]any{"games": []contract.GameDTO{{Source: "feed", SourceID: "2", URL: "https://feed/2", Title: "Second", UpdatedAt: "2026-01-01T00:00:00.000Z"}}}
	}
	raw, err = json.Marshal(reply)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, out)
}

func (listingSources) StartDownload(context.Context, string, any, int64) (int64, error) {
	return 0, errors.New("downloads are not faked here")
}

// emitted is one event the service sent.
type emitted struct {
	channel string
	data    any
}

// eventLog keeps every event the service emits, in order.
type eventLog struct {
	mu     sync.Mutex
	events []emitted
}

func (e *eventLog) Emit(channel string, data any) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.events = append(e.events, emitted{channel: channel, data: data})
}

func (e *eventLog) snapshot() []emitted {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]emitted(nil), e.events...)
}

// waitForSyncPhase polls the log until a library.sync event with the phase arrives.
func waitForSyncPhase(t *testing.T, log *eventLog, phase string) syncEvent {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		for _, event := range log.snapshot() {
			sync, ok := event.data.(syncEvent)
			if event.channel == syncEventChannel && ok && sync.Phase == phase {
				return sync
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("no library.sync event with phase %q", phase)
	return syncEvent{}
}

func newSyncRouter(t *testing.T, events api.Events) *api.Router {
	t.Helper()
	lib, err := library.Open(context.Background(), library.Options{
		DBPath:   filepath.Join(t.TempDir(), "library.db"),
		Sources:  listingSources{},
		MediaDir: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { lib.Close() })
	router := api.NewRouter()
	Register(router, events, lib, listingSources{})
	return router
}

func TestLibrarySyncSourceRunsInTheBackgroundAndReports(t *testing.T) {
	log := &eventLog{}
	router := newSyncRouter(t, log)

	started := call(t, router, "LibrarySyncSource", `{"source":"feed"}`).(syncSourceResult)
	if !started.Started {
		t.Fatal("started = false")
	}

	done := waitForSyncPhase(t, log, syncPhaseDone)
	if done.Source != "feed" || done.Imported != 2 || done.Failed != 0 {
		t.Errorf("done event = %+v, want 2 imported from feed", done)
	}
	sawChanged := false
	for _, event := range log.snapshot() {
		if event.channel == changedEventChannel {
			sawChanged = true
		}
	}
	if !sawChanged {
		t.Error("the sync must tell the UI the library changed")
	}
	if page := call(t, router, "LibraryList", `{}`).(library.Page); page.Total == nil || *page.Total != 2 {
		t.Errorf("library total after sync = %v, want 2", page.Total)
	}
}

func TestLibrarySyncSourceRejectsSourceWithoutListGames(t *testing.T) {
	router, _ := setup(t)
	if _, err := router.Dispatch(context.Background(), "LibrarySyncSource", json.RawMessage(`{"source":"f95"}`)); err == nil {
		t.Error("a source without listGames must be rejected")
	}
}

func TestLibraryOverviewReturnsSourcesTagsAndStats(t *testing.T) {
	router, _ := setup(t)
	createManual(t, router)

	overview := call(t, router, "LibraryOverview", ``).(library.Overview)
	if overview.Stats.Total != 1 {
		t.Errorf("stats.total = %d, want 1", overview.Stats.Total)
	}
	for _, source := range overview.Sources {
		if source.ID == "manual" && (source.Label != "Local games" || source.Tile != "square" || source.Count != 1) {
			t.Errorf("manual source = %+v, want Local games, square, 1 game", source)
		}
	}
}
