package main

import (
	"os"
	"path/filepath"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/plugins"
)

// folders are the places the launcher reads and writes. They all sit under the config folder.
type folders struct {
	Config          string // databases and the lock file
	Games           string // installed games, one folder per game
	Downloads       string // finished downloads
	Incoming        string // watch folder for files the user drops in
	Media           string // cached covers and screenshots
	PluginData      string // data kept by plugins
	BrowserProfiles string // persistent browser profiles, one folder per plugin
	Uploads         string // plugin files sent by the UI, cleared on start and after install
}

func prepareFolders() (folders, error) {
	base := config.GetBaseDirectory()
	dirs := folders{
		Config:          base,
		Games:           filepath.Join(base, "games"),
		Downloads:       filepath.Join(base, "downloads"),
		Incoming:        filepath.Join(base, "incoming"),
		Media:           filepath.Join(base, "media"),
		PluginData:      plugins.DefaultLayout().Data,
		BrowserProfiles: filepath.Join(base, "browser-profiles"),
		Uploads:         filepath.Join(base, "temp", "uploads"),
	}
	for _, dir := range []string{dirs.Config, dirs.Games, dirs.Downloads, dirs.Incoming, dirs.Media} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return dirs, err
		}
	}
	return dirs, nil
}
