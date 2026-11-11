package main

import (
	"path/filepath"
	"time"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/library"
)

const (
	watchInterval  = 2 * time.Second  // how often the watch folder is polled
	updateInterval = 30 * time.Minute // how often due games are checked for updates
	updateRecheck  = 6 * time.Hour    // a game checked more recently than this is skipped
)

// libraryOptions wires the library to its neighbours. Notifications go to the UI event stream.
func libraryOptions(dirs folders, hub *api.Hub, sources contract.Sources, downloads contract.Downloads, updates library.Installer, sink library.ArchiveSink, launcher library.Launcher) library.Options {
	return library.Options{
		DBPath:        filepath.Join(dirs.Config, "library.db"),
		MediaDir:      dirs.Media,
		DownloadsDir:  dirs.Downloads,
		GamesDir:      func() string { return dirs.Games },
		WatchFolder:   func() string { return dirs.Incoming },
		Sources:       sources,
		Downloads:     downloads,
		Launcher:      launcher,
		Installer:     updates,
		Sink:          sink,
		UpdateRecheck: updateRecheck,
		Notify:        hub.Emit,
	}
}
