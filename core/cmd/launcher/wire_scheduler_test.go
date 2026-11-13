package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/library"
	"light-launcher/core/internal/plugins"
	"light-launcher/core/internal/scheduler"
)

func TestSyncCapable(t *testing.T) {
	cases := []struct {
		name   string
		source contract.SourceInfo
		want   bool
	}{
		{"games", contract.SourceInfo{Capabilities: []string{"browse", "sync:games"}}, true},
		{"index", contract.SourceInfo{Capabilities: []string{"sync:index"}}, true},
		{"list only", contract.SourceInfo{Capabilities: []string{"browse", "listGames"}}, true},
		{"none", contract.SourceInfo{Capabilities: []string{"browse", "updateCheck"}}, false},
	}
	for _, tc := range cases {
		if got := syncCapable(tc.source); got != tc.want {
			t.Errorf("%s: syncCapable = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// fakeSources answers the source calls of a real library from a fake plugin host.
type fakeSources struct {
	infos   []contract.SourceInfo
	replies map[string]func(params any) (any, error)
	calls   []string
}

func (f *fakeSources) Sources() []contract.SourceInfo { return f.infos }

func (f *fakeSources) Call(_ context.Context, plugin, method string, params, out any) error {
	f.calls = append(f.calls, plugin+" "+method)
	reply, ok := f.replies[method]
	if !ok {
		return fmt.Errorf("no fake for %s", method)
	}
	value, err := reply(params)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, out)
}

func (f *fakeSources) StartDownload(context.Context, string, any, int64) (int64, error) {
	return 0, errors.New("not used")
}

func (f *fakeSources) Browse(context.Context, string, contract.BrowseRequest) (contract.BrowsePage, error) {
	return contract.BrowsePage{}, errors.New("not used")
}

func (f *fakeSources) CallAction(context.Context, string, string, any, any) error {
	return errors.New("not used")
}

func (f *fakeSources) ListFilters(context.Context, string, contract.FiltersRequest) (contract.FiltersPage, error) {
	return contract.FiltersPage{}, errors.New("not used")
}

// fakeImporter records the sources a scheduled job sends to the library.
type fakeImporter struct {
	calls []string
	err   error
}

func (f *fakeImporter) importSource(_ context.Context, plugin, source string) error {
	f.calls = append(f.calls, plugin+" "+source)
	return f.err
}

func TestRunSourceSyncSendsEverySourceToTheLibrary(t *testing.T) {
	cases := []struct {
		name   string
		source contract.SourceInfo
	}{
		{"listGames", contract.SourceInfo{SourceID: "mine", Capabilities: []string{"listGames"}}},
		{"listGames wins over sync:games", contract.SourceInfo{SourceID: "mine", Capabilities: []string{"sync:games", "listGames"}}},
		{"sync:games", contract.SourceInfo{SourceID: "shop", Capabilities: []string{"sync:games"}}},
		{"sync:index", contract.SourceInfo{SourceID: "shop", Capabilities: []string{"sync:index"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			importer := &fakeImporter{}
			if err := runSourceSync(t.Context(), importer.importSource, "demo", tc.source); err != nil {
				t.Fatalf("runSourceSync: %v", err)
			}
			want := []string{"demo " + tc.source.SourceID}
			if !slices.Equal(importer.calls, want) {
				t.Errorf("imports = %v, want %v", importer.calls, want)
			}
		})
	}
}

func TestRunSourceSyncReportsFailures(t *testing.T) {
	boom := errors.New("boom")
	listed := contract.SourceInfo{SourceID: "mine", Capabilities: []string{"listGames"}}
	syncOnly := contract.SourceInfo{SourceID: "shop", Capabilities: []string{"sync:games"}}

	importer := &fakeImporter{err: boom}
	err := runSourceSync(t.Context(), importer.importSource, "demo", listed)
	if !errors.Is(err, boom) || !strings.Contains(err.Error(), "import games of source mine") {
		t.Errorf("import failure = %v, want it to wrap %v as an import", err, boom)
	}
	err = runSourceSync(t.Context(), importer.importSource, "demo", syncOnly)
	if !errors.Is(err, boom) || !strings.Contains(err.Error(), "sync source shop") {
		t.Errorf("sync failure = %v, want it to wrap %v as a sync", err, boom)
	}
}

// openSchedulerLibrary opens a real library that stores its games in a temp database.
func openSchedulerLibrary(t *testing.T, sources contract.Sources) *library.Library {
	t.Helper()
	dir := t.TempDir()
	lib, err := library.Open(t.Context(), library.Options{
		DBPath:   filepath.Join(dir, "library.db"),
		MediaDir: filepath.Join(dir, "media"),
		Sources:  sources,
	})
	if err != nil {
		t.Fatalf("open library: %v", err)
	}
	t.Cleanup(func() {
		if err := lib.Close(); err != nil {
			t.Errorf("close library: %v", err)
		}
	})
	return lib
}

// storedGames counts the games a scheduled run saved.
func storedGames(t *testing.T, lib *library.Library) int {
	t.Helper()
	page, err := lib.List(t.Context(), library.Filter{Limit: 1, WantTotal: true})
	if err != nil {
		t.Fatalf("list library: %v", err)
	}
	if page.Total == nil {
		t.Fatal("the library page has no total")
	}
	return *page.Total
}

// syncJob builds the one scheduled sync job of a fake plugin host.
func syncJob(t *testing.T, plugin string, sources *fakeSources, importSource func(context.Context, string, string) error) scheduler.Job {
	t.Helper()
	host := &fakeSchedulerHost{infos: []plugins.PluginInfo{{
		ID: plugin, Enabled: true, Status: "stopped", Sources: sources.infos,
	}}}
	jobs := pluginSyncJobs(host, importSource)
	if len(jobs) != 1 {
		t.Fatalf("jobs = %d, want the one sync job of %s", len(jobs), plugin)
	}
	return jobs[0]
}

// scheduledGames builds the games a fake source returns, stamped so a repeat run could skip them.
func scheduledGames(source string, titles ...string) []contract.GameDTO {
	games := make([]contract.GameDTO, 0, len(titles))
	for i, title := range titles {
		id := fmt.Sprintf("%d", i+1)
		games = append(games, contract.GameDTO{
			Source: source, SourceID: id, URL: "https://" + source + "/" + id,
			Title: title, Version: "1", UpdatedAt: "2026-11-10T08:00:00Z",
		})
	}
	return games
}

func TestScheduledSyncSavesGamesOfSyncOnlySource(t *testing.T) {
	sources := &fakeSources{
		infos: []contract.SourceInfo{{Plugin: "itch", SourceID: "itch", Label: "itch.io", Capabilities: []string{capabilitySyncGames}}},
		replies: map[string]func(any) (any, error){
			"source.sync": func(any) (any, error) { return scheduledGames("itch", "First", "Second"), nil },
		},
	}
	lib := openSchedulerLibrary(t, sources)
	job := syncJob(t, "itch", sources, libraryImporter(lib))

	if err := job.Run(t.Context()); err != nil {
		t.Fatalf("scheduled sync: %v", err)
	}
	if got := storedGames(t, lib); got != 2 {
		t.Errorf("stored = %d, want both games of source.sync", got)
	}
	if want := []string{"itch source.sync"}; !slices.Equal(sources.calls, want) {
		t.Errorf("source calls = %v, want %v", sources.calls, want)
	}
}

func TestScheduledSyncAcceptsStatusWithoutGames(t *testing.T) {
	sources := &fakeSources{
		infos: []contract.SourceInfo{{Plugin: "itch", SourceID: "itch", Capabilities: []string{capabilitySyncIndex}}},
		replies: map[string]func(any) (any, error){
			"source.sync": func(any) (any, error) {
				return map[string]any{"count": 1250, "syncedAt": "2026-11-10T08:00:00Z", "running": false}, nil
			},
		},
	}
	lib := openSchedulerLibrary(t, sources)
	job := syncJob(t, "itch", sources, libraryImporter(lib))

	if err := job.Run(t.Context()); err != nil {
		t.Fatalf("scheduled sync: %v", err)
	}
	if got := storedGames(t, lib); got != 0 {
		t.Errorf("stored = %d, want a status reply to store nothing", got)
	}
}

func TestScheduledImportStillSavesListedGames(t *testing.T) {
	sources := &fakeSources{
		infos: []contract.SourceInfo{{Plugin: "feed", SourceID: "feed", Capabilities: []string{capabilityListGames}}},
		replies: map[string]func(any) (any, error){
			"source.listGames": func(any) (any, error) {
				return map[string]any{"games": scheduledGames("feed", "First", "Second")}, nil
			},
		},
	}
	lib := openSchedulerLibrary(t, sources)
	job := syncJob(t, "feed", sources, libraryImporter(lib))

	if err := job.Run(t.Context()); err != nil {
		t.Fatalf("scheduled import: %v", err)
	}
	if got := storedGames(t, lib); got != 2 {
		t.Errorf("stored = %d, want both listed games", got)
	}
	if want := []string{"feed source.listGames"}; !slices.Equal(sources.calls, want) {
		t.Errorf("source calls = %v, want %v", sources.calls, want)
	}
}

func TestSyncIntervalFromSettings(t *testing.T) {
	withSchema := func(value any) plugins.SettingsView {
		return plugins.SettingsView{
			Schema: map[string]any{"properties": map[string]any{"syncIntervalHours": map[string]any{"type": "integer"}}},
			Values: map[string]any{"syncIntervalHours": value},
		}
	}
	cases := []struct {
		name string
		view plugins.SettingsView
		want time.Duration
	}{
		{"saved hours", withSchema(float64(6)), 6 * time.Hour},
		{"saved zero", withSchema(float64(0)), syncIntervalDefault},
		{"wrong type", withSchema("6"), syncIntervalDefault},
		{
			name: "schema has no setting",
			view: plugins.SettingsView{
				Schema: map[string]any{"properties": map[string]any{"other": map[string]any{}}},
				Values: map[string]any{"syncIntervalHours": float64(6)},
			},
			want: syncIntervalDefault,
		},
		{"no schema", plugins.SettingsView{}, syncIntervalDefault},
	}
	for _, tc := range cases {
		if got := syncIntervalFromSettings(tc.view); got != tc.want {
			t.Errorf("%s: interval = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestPluginRunnable(t *testing.T) {
	cases := []struct {
		name string
		info plugins.PluginInfo
		want bool
	}{
		{"stopped", plugins.PluginInfo{ID: "f95", Enabled: true, Status: "stopped"}, true},
		{"running", plugins.PluginInfo{ID: "f95", Enabled: true, Status: "running"}, true},
		{"disabled", plugins.PluginInfo{ID: "f95", Enabled: false, Status: "disabled"}, false},
		{"broken", plugins.PluginInfo{ID: "f95", Enabled: true, Status: "broken"}, false},
		{"missing", plugins.PluginInfo{}, false},
	}
	for _, tc := range cases {
		if got := pluginRunnable(tc.info); got != tc.want {
			t.Errorf("%s: pluginRunnable = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// fakeSchedulerHost stands in for the plugin host while testing the schedule.
type fakeSchedulerHost struct {
	infos      []plugins.PluginInfo
	checkErr   error
	checkedIDs []string
}

func (f *fakeSchedulerHost) List() []plugins.PluginInfo { return f.infos }

func (f *fakeSchedulerHost) SettingsGet(string) (plugins.SettingsView, error) {
	return plugins.SettingsView{}, nil
}

func (f *fakeSchedulerHost) CheckUpdates(_ context.Context, id string) (plugins.PluginUpdatesStatus, error) {
	f.checkedIDs = append(f.checkedIDs, id)
	if f.checkErr != nil {
		return plugins.PluginUpdatesStatus{}, f.checkErr
	}
	return plugins.PluginUpdatesStatus{}, nil
}

// fakeSchedulerLibrary stands in for the library's schedule and last-run store.
type fakeSchedulerLibrary struct {
	meta map[string]string
}

func (f *fakeSchedulerLibrary) GetMeta(_ context.Context, key string) (string, bool, error) {
	value, ok := f.meta[key]
	return value, ok, nil
}

func (f *fakeSchedulerLibrary) SetMeta(_ context.Context, key, value string) error {
	if f.meta == nil {
		f.meta = map[string]string{}
	}
	f.meta[key] = value
	return nil
}

func (f *fakeSchedulerLibrary) ScheduledCheck(context.Context, bool) (bool, error) {
	return false, nil
}

func TestPluginUpdateCheckJobSpec(t *testing.T) {
	job := pluginUpdateCheckJob(&fakeSchedulerHost{})
	if job.Name != "plugins.update-check" {
		t.Errorf("name = %q, want plugins.update-check", job.Name)
	}
	if job.Interval != 24*time.Hour {
		t.Errorf("interval = %v, want 24h", job.Interval)
	}
	if job.Timeout <= 0 {
		t.Errorf("timeout = %v, want a bound", job.Timeout)
	}
	if job.Jitter <= 0 {
		t.Errorf("jitter = %v, want a delay", job.Jitter)
	}
	if job.Run == nil {
		t.Error("job has no run function")
	}
}

func TestPluginUpdateCheckJobCallsHost(t *testing.T) {
	host := &fakeSchedulerHost{}
	if err := pluginUpdateCheckJob(host).Run(t.Context()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !slices.Equal(host.checkedIDs, []string{""}) {
		t.Errorf("checked ids = %q, want one full check", host.checkedIDs)
	}

	boom := errors.New("github is down")
	host = &fakeSchedulerHost{checkErr: boom}
	err := pluginUpdateCheckJob(host).Run(t.Context())
	if !errors.Is(err, boom) {
		t.Fatalf("failed check = %v, want it to wrap %v", err, boom)
	}
	if !slices.Equal(host.checkedIDs, []string{""}) {
		t.Errorf("checked ids after a failure = %q, want one full check", host.checkedIDs)
	}
}

func TestSchedulerStatusKeepsDailyPluginCheck(t *testing.T) {
	host := &fakeSchedulerHost{infos: []plugins.PluginInfo{{
		ID:      "demo",
		Enabled: true,
		Status:  "stopped",
		Sources: []contract.SourceInfo{{SourceID: "mine", Capabilities: []string{"listGames"}}},
	}}}
	sched, rebuild, err := newScheduler(t.Context(), &fakeSchedulerLibrary{}, host, nil, nil, nil)
	if err != nil {
		t.Fatalf("newScheduler: %v", err)
	}

	status := sched.Status()
	want := []string{"library.update-check", "plugin-sync:demo:mine", "plugins.update-check"}
	if names := jobNames(status); !slices.Equal(names, want) {
		t.Fatalf("jobs = %v, want %v", names, want)
	}
	if got := jobInterval(status, "plugins.update-check"); got != int64((24*time.Hour)/time.Second) {
		t.Errorf("plugin check interval = %ds, want %ds", got, int64((24*time.Hour)/time.Second))
	}

	// Installing or enabling a plugin rebuilds the jobs; the daily check must stay.
	host.infos = append(host.infos, plugins.PluginInfo{
		ID:      "extra",
		Enabled: true,
		Status:  "stopped",
		Sources: []contract.SourceInfo{{SourceID: "shop", Capabilities: []string{"sync:games"}}},
	})
	if err := rebuild(); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	status = sched.Status()
	want = []string{"library.update-check", "plugin-sync:demo:mine", "plugin-sync:extra:shop", "plugins.update-check"}
	if names := jobNames(status); !slices.Equal(names, want) {
		t.Fatalf("jobs after a plugin was installed = %v, want %v", names, want)
	}

	// Disabling every plugin removes the sync jobs but not the daily check.
	host.infos = nil
	if err := rebuild(); err != nil {
		t.Fatalf("rebuild without plugins: %v", err)
	}
	status = sched.Status()
	want = []string{"library.update-check", "plugins.update-check"}
	if names := jobNames(status); !slices.Equal(names, want) {
		t.Fatalf("jobs without plugins = %v, want %v", names, want)
	}
}

func jobNames(status scheduler.Status) []string {
	names := make([]string, 0, len(status.Jobs))
	for _, job := range status.Jobs {
		names = append(names, job.Name)
	}
	return names
}

func jobInterval(status scheduler.Status, name string) int64 {
	for _, job := range status.Jobs {
		if job.Name == name {
			return job.IntervalSeconds
		}
	}
	return 0
}

func TestPluginUpdateCheckFailureKeepsSchedulerRunning(t *testing.T) {
	host := &fakeSchedulerHost{checkErr: errors.New("github is down")}
	sched, _, err := newScheduler(t.Context(), &fakeSchedulerLibrary{}, host, nil, nil, nil)
	if err != nil {
		t.Fatalf("newScheduler: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		defer close(done)
		sched.Run(ctx)
	}()
	defer func() {
		cancel()
		<-done
	}()

	select {
	case <-sched.Ready():
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler did not become ready")
	}
	started, err := sched.RunNow("plugins.update-check")
	if err != nil || !started {
		t.Fatalf("RunNow = %v, %v; want true, nil", started, err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		job, ok := jobStatusByName(sched.Status(), "plugins.update-check")
		if !ok {
			t.Fatal("the daily plugin check is missing")
		}
		if strings.Contains(job.LastError, "github is down") {
			if job.NextRunAt == nil {
				t.Error("a failed job must stay scheduled")
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("failure was not recorded: %+v", job)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func jobStatusByName(status scheduler.Status, name string) (scheduler.JobStatus, bool) {
	for _, job := range status.Jobs {
		if job.Name == name {
			return job, true
		}
	}
	return scheduler.JobStatus{}, false
}
