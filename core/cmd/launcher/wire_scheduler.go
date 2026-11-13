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
	capabilitySyncGames = "sync:games"
	capabilitySyncIndex = "sync:index"
	capabilityListGames = "listGames"
	methodSourceSync    = "source.sync"
)

// newScheduler builds the plugin sync jobs and the library update check. The scheduler itself
// knows nothing about plugins or the library. importSource imports the games of a source that lists them.
func newScheduler(ctx context.Context, lib *library.Library, host *plugins.Host, sources contract.Sources, importSource func(ctx context.Context, plugin, source string) error, notify func(string, any)) (*scheduler.Scheduler, func() error, error) {
	build := func() []scheduler.Job {
		jobs := pluginSyncJobs(host, sources, importSource)
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

// updateCheckRun returns the run function of the background update check. The first call is the
// run after launch. Later calls run when the schedule allows it, and the scheduler skips the rest.
func updateCheckRun(lib *library.Library) func(ctx context.Context) error {
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

func pluginSyncJobs(host *plugins.Host, sources contract.Sources, importSource func(ctx context.Context, plugin, source string) error) []scheduler.Job {
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
			jobs = append(jobs, pluginSyncJob(host, sources, importSource, info.ID, source, interval))
		}
	}
	return jobs
}

func pluginSyncJob(host *plugins.Host, sources contract.Sources, importSource func(ctx context.Context, plugin, source string) error, pluginID string, source contract.SourceInfo, interval time.Duration) scheduler.Job {
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
			return runSourceSync(ctx, sources, importSource, pluginID, source)
		},
	}
}

// runSourceSync imports the games of a source that lists them. Other sources run their own source.sync.
func runSourceSync(ctx context.Context, sources contract.Sources, importSource func(ctx context.Context, plugin, source string) error, pluginID string, source contract.SourceInfo) error {
	if listsGames(source) {
		return importListedGames(ctx, importSource, pluginID, source.SourceID)
	}
	if err := sources.Call(ctx, pluginID, methodSourceSync, map[string]string{"source": source.SourceID}, nil); err != nil {
		return fmt.Errorf("sync source %s: %w", source.SourceID, err)
	}
	return nil
}

// libraryImporter adapts the library import to the scheduler's job signature. A progress callback is not
// needed for a scheduled run: the library still sends its throttled change events.
func libraryImporter(lib *library.Library) func(ctx context.Context, plugin, source string) error {
	return func(ctx context.Context, plugin, source string) error {
		_, err := lib.ImportSource(ctx, plugin, source, nil)
		return err
	}
}

func importListedGames(ctx context.Context, importSource func(ctx context.Context, plugin, source string) error, pluginID, sourceID string) error {
	if err := importSource(ctx, pluginID, sourceID); err != nil {
		return fmt.Errorf("import games of source %s: %w", sourceID, err)
	}
	return nil
}

func pluginSyncInterval(host *plugins.Host, id string) time.Duration {
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

func findPlugin(host *plugins.Host, id string) plugins.PluginInfo {
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
