package main

import (
	"context"
	"fmt"
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
	updateCheckJitter   = 2 * time.Minute
	updateCheckTimeout  = 10 * time.Minute
	capabilitySyncGames = "sync:games"
	capabilitySyncIndex = "sync:index"
	methodSourceSync    = "source.sync"
)

// newScheduler builds the plugin sync jobs and the library update check. The scheduler itself
// knows nothing about plugins or the library.
func newScheduler(ctx context.Context, lib *library.Library, host *plugins.Host, sources contract.Sources, notify func(string, any)) (*scheduler.Scheduler, error) {
	jobs := pluginSyncJobs(host, sources)
	jobs = append(jobs, scheduler.Job{
		Name:     "library.update-check",
		Interval: updateInterval,
		Jitter:   updateCheckJitter,
		Timeout:  updateCheckTimeout,
		Run: func(ctx context.Context) error {
			_, err := lib.CheckAll(ctx, nil, true)
			return err
		},
	})
	return scheduler.New(ctx, scheduler.Options{
		Jobs:   jobs,
		State:  lib,
		Notify: notify,
	})
}

func pluginSyncJobs(host *plugins.Host, sources contract.Sources) []scheduler.Job {
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
			pluginID, sourceID := info.ID, source.SourceID
			jobs = append(jobs, scheduler.Job{
				Name:           fmt.Sprintf("plugin-sync:%s:%s", pluginID, sourceID),
				PluginID:       pluginID,
				Interval:       interval,
				Jitter:         syncJobJitter,
				Timeout:        syncJobTimeout,
				ChangesLibrary: true,
				Run: func(ctx context.Context) error {
					if !pluginRunnable(findPlugin(host, pluginID)) {
						return scheduler.ErrSkipped
					}
					if err := sources.Call(ctx, pluginID, methodSourceSync, map[string]string{"source": sourceID}, nil); err != nil {
						return fmt.Errorf("sync source %s: %w", sourceID, err)
					}
					return nil
				},
			})
		}
	}
	return jobs
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
	for _, capability := range source.Capabilities {
		if capability == capabilitySyncGames || capability == capabilitySyncIndex {
			return true
		}
	}
	return false
}
