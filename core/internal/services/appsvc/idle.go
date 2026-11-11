package appsvc

import (
	"time"

	"light-launcher/core/internal/config"
)

// PluginIdleTimeout returns the idle stop time from the app settings. Nil means the host default.
func PluginIdleTimeout() *time.Duration {
	settings := config.NewSettingsStore(config.SettingsFilePath()).Load()
	if settings.PluginIdleTimeoutMinutes == nil {
		return nil
	}
	timeout := time.Duration(*settings.PluginIdleTimeoutMinutes) * time.Minute
	return &timeout
}
