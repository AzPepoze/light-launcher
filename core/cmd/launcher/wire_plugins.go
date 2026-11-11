package main

import (
	"context"
	"time"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/plugins"
)

// pluginUpdateCheckDelay leaves the first start quiet while the launcher is still opening.
const pluginUpdateCheckDelay = 60 * time.Second

// newPluginHost creates the plugin host. Plugins may ask for these folders by name in plugin.json.
// Secrets use the host default: the keyring when available, otherwise an encrypted file.
func newPluginHost(dirs folders, library contract.Library, downloads contract.Downloads) (*plugins.Host, error) {
	layout := plugins.DefaultLayout()
	return plugins.NewHost(plugins.HostOptions{
		Layout:      layout,
		HostVersion: hostVersion(version),
		Library:     library,
		Downloads:   downloads,
		Folders: map[string]string{
			"$GAMES":     dirs.Games,
			"$DOWNLOADS": dirs.Downloads,
			"$DATA":      dirs.PluginData,
		},
	})
}

// runPluginUpdateCheck checks GitHub once after start. The UI calls the same host method for "Check now".
func (s *services) runPluginUpdateCheck(ctx context.Context) {
	timer := time.NewTimer(pluginUpdateCheckDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return
	case <-timer.C:
	}
	if _, err := s.host.CheckUpdates(ctx, ""); err != nil {
		logger.Warn("plugins", "check updates: %v", err)
	}
}
