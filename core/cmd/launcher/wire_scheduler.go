package main

import (
	"context"
	"fmt"
	"slices"
	"time"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/library"
	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/plugins"
	"light-launcher/core/internal/scheduler"
)

const (
	syncIntervalDefault = 24 * time.Hour
	syncJobTimeout      = 10 * time.Minute
	syncJobJitter       = 5 * time.Minute
	updateCheckTimeout  = 10 * time.Minute

	pluginUpdateCheckJobName  = "plugins.update-check"
	pluginUpdateCheckInterval = 24 * time.Hour
	pluginUpdateCheckTimeout  = 10 * time.Minute
	pluginUpdateCheckJitter   = 5 * time.Minute

	capabilitySyncGames = "sync:games"
	capabilitySyncIndex = "sync:index"
	capabilityListGames = "listGames"
)

// schedulerHost is the part of the plugin host the schedule needs.
type schedulerHost interface {
	List() []plugins.PluginInfo
	SettingsGet(id string) (plugins.SettingsView, error)
	CheckUpdates(ctx context.Context, id string) (plugins.PluginUpdatesStatus, error)
}

// schedulerLibrary is the part of the library the schedule needs. The library also keeps the last
// run times, so the scheduler does not check GitHub or the sources again after a restart too early.
type schedulerLibrary interface {
	scheduler.MetaStore
	ScheduledCheck(ctx context.Context, startup bool) (bool, error)
}

// newScheduler builds the plugin sync jobs, the daily plugin release check and the library update
// check. Every source goes through importSource, so the library owns the plugin calls and a source
// that syncs its own games is stored like a manual sync. The sources argument is unused: the library
// already holds the source list, but the launcher wiring still passes it.
func newScheduler(ctx context.Context, lib schedulerLibrary, host schedulerHost, _ contract.Sources, importSource func(ctx context.Context, plugin, source string) error, notify func(string, any)) (*scheduler.Scheduler, func() error, error) {
	build := func() []scheduler.Job {
		jobs := pluginSyncJobs(host, importSource)
		jobs = append(jobs, pluginUpdateCheckJob(host))
		jobs = append(jobs, scheduler.Job{
			Name:     "library.update-check",
			Interval: library.UpdateTick,
			Timeout:  updateCheckTimeout,
			Run:      updateCheckRun(lib),
		})
		return jobs
	}
	sched, err := scheduler.New(ctx, scheduler.Options{
		Jobs:   build(),
		State:  lib,
		Notify: notify,
	})
	if err != nil {
		return nil, nil, err
	}
	return sched, func() error { return sched.SetJobs(build()) }, nil
}

// pluginUpdateCheckJob checks the GitHub releases of every installed plugin once a day. The startup
// check and "Check now" stay as they are. A failed check is logged and retried on the next run.
func pluginUpdateCheckJob(host schedulerHost) scheduler.Job {
	return scheduler.Job{
		Name:     pluginUpdateCheckJobName,
		Interval: pluginUpdateCheckInterval,
		Timeout:  pluginUpdateCheckTimeout,
		Jitter:   pluginUpdateCheckJitter,
		Run: func(ctx context.Context) error {
			if _, err := host.CheckUpdates(ctx, ""); err != nil {
				return fmt.Errorf("check plugin updates: %w", err)
			}
			return nil
		},
	}
}

// updateCheckRun returns the run function of the background update check. The first call is the
// run after launch. Later calls run when the schedule allows it, and the scheduler skips the rest.
func updateCheckRun(lib schedulerLibrary) func(ctx context.Context) error {
	first := true
	return func(ctx context.Context) error {
		startup := first
		first = false
		ran, err := lib.ScheduledCheck(ctx, startup)
		if err != nil {
			return err
		}
		if !ran {
			return scheduler.ErrSkipped
		}
		return nil
	}
}

func pluginSyncJobs(host schedulerHost, importSource func(ctx context.Context, plugin, source string) error) []scheduler.Job {
	var jobs []scheduler.Job
	for _, info := range host.List() {
		if !pluginRunnable(info) {
			continue
		}
		interval := pluginSyncInterval(host, info.ID)
		for _, source := range info.Sources {
			if !syncCapable(source) {
				continue
			}
			jobs = append(jobs, pluginSyncJob(host, importSource, info.ID, source, interval))
		}
	}
	return jobs
}

func pluginSyncJob(host schedulerHost, importSource func(ctx context.Context, plugin, source string) error, pluginID string, source contract.SourceInfo, interval time.Duration) scheduler.Job {
	return scheduler.Job{
		Name:           fmt.Sprintf("plugin-sync:%s:%s", pluginID, source.SourceID),
		PluginID:       pluginID,
		Interval:       interval,
		Jitter:         syncJobJitter,
		Timeout:        syncJobTimeout,
		ChangesLibrary: true,
		Run: func(ctx context.Context) error {
			if !pluginRunnable(findPlugin(host, pluginID)) {
				return scheduler.ErrSkipped
			}
			return runSourceSync(ctx, importSource, pluginID, source)
		},
	}
}

// runSourceSync fills the library from one source. Both routes go through the library: it imports a
// source that lists games page by page, and saves the games that a sync-only source returns from
// source.sync. Only the error text differs, so an operator can tell the two apart.
func runSourceSync(ctx context.Context, importSource func(ctx context.Context, plugin, source string) error, pluginID string, source contract.SourceInfo) error {
	err := importSource(ctx, pluginID, source.SourceID)
	if err == nil {
		return nil
	}
	if listsGames(source) {
		return fmt.Errorf("import games of source %s: %w", source.SourceID, err)
	}
	return fmt.Errorf("sync source %s: %w", source.SourceID, err)
}

// libraryImporter adapts the library sync to the scheduler's job signature. A progress callback is not
// needed for a scheduled run: the library still sends its throttled change events.
func libraryImporter(lib *library.Library) func(ctx context.Context, plugin, source string) error {
	return func(ctx context.Context, plugin, source string) error {
		_, err := lib.SyncSource(ctx, plugin, source)
		return err
	}
}

func pluginSyncInterval(host schedulerHost, id string) time.Duration {
	view, err := host.SettingsGet(id)
	if err != nil {
		logger.Warn("scheduler", "settings of %s: %v", id, err)
		return syncIntervalDefault
	}
	return syncIntervalFromSettings(view)
}

func syncIntervalFromSettings(view plugins.SettingsView) time.Duration {
	properties, _ := view.Schema["properties"].(map[string]any)
	if _, ok := properties["syncIntervalHours"]; !ok {
		return syncIntervalDefault
	}
	hours, ok := view.Values["syncIntervalHours"].(float64)
	if !ok || hours <= 0 {
		return syncIntervalDefault
	}
	return time.Duration(hours * float64(time.Hour))
}

func findPlugin(host schedulerHost, id string) plugins.PluginInfo {
	for _, info := range host.List() {
		if info.ID == id {
			return info
		}
	}
	return plugins.PluginInfo{}
}

func pluginRunnable(info plugins.PluginInfo) bool {
	return info.ID != "" && info.Enabled && info.Status != "broken"
}

func syncCapable(source contract.SourceInfo) bool {
	return listsGames(source) ||
		slices.Contains(source.Capabilities, capabilitySyncGames) ||
		slices.Contains(source.Capabilities, capabilitySyncIndex)
}

func listsGames(source contract.SourceInfo) bool {
	return slices.Contains(source.Capabilities, capabilityListGames)
}
