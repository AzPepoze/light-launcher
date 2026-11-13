package main

import (
	"path/filepath"
	"time"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/config"
	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/library"
	"light-launcher/core/internal/types"
)

const watchInterval = 2 * time.Second // how often the watch folder is polled

// libraryOptions wires the library to its neighbours. Notifications go to the UI event stream.
func libraryOptions(dirs folders, hub *api.Hub, sources contract.Sources, downloads contract.Downloads, updates library.Installer, sink library.ArchiveSink, launcher interface {
	library.Launcher
	library.ConfigLauncher
}) library.Options {
	return library.Options{
		DBPath:         filepath.Join(dirs.Config, "library.db"),
		MediaDir:       dirs.Media,
		DownloadsDir:   dirs.Downloads,
		GamesDir:       func() string { return dirs.Games },
		WatchFolder:    func() string { return dirs.Incoming },
		Sources:        sources,
		Downloads:      downloads,
		Launcher:       launcher,
		ConfigLauncher: launcher,
		Installer:      updates,
		Sink:           sink,
		UpdateRule:     updateRule,
		UpdateSchedule: updateSchedule,
		Notify:         hub.Emit,
	}
}

// updateRule returns how background checks treat a source. It reads the saved settings on every call.
func updateRule(source string) library.UpdateRule {
	settings := loadSettings()
	enabled, recheck := types.ResolveUpdateCheck(settings, source)
	return library.UpdateRule{Auto: enabled, Recheck: recheck, Interval: types.ResolveSourceInterval(settings, source)}
}

// updateSchedule returns when background runs start. It reads the saved settings on every call.
func updateSchedule() library.UpdateSchedule {
	enabled, onStartup, interval := types.ResolveSchedule(loadSettings())
	return library.UpdateSchedule{Enabled: enabled, OnStartup: onStartup, Interval: interval}
}

func loadSettings() types.AppSettings {
	return config.NewSettingsStore(config.SettingsFilePath()).Load()
}
