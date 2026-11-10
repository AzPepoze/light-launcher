package main

import (
	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/plugins"
)

// newPluginHost creates the plugin host. Plugins may ask for these folders by name in plugin.json.
// Secrets use the host default: the keyring when available, otherwise an encrypted file.
func newPluginHost(dirs folders, library contract.Library, downloads contract.Downloads) (*plugins.Host, error) {
	layout := plugins.DefaultLayout()
	return plugins.NewHost(plugins.HostOptions{
		Layout:      layout,
		HostVersion: version,
		Library:     library,
		Downloads:   downloads,
		Folders: map[string]string{
			"$GAMES":     dirs.Games,
			"$DOWNLOADS": dirs.Downloads,
			"$DATA":      dirs.PluginData,
		},
	})
}
