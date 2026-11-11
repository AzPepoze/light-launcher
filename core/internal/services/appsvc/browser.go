package appsvc

import "light-launcher/core/internal/config"

// BrowserAssistEnabled reports whether plugins may open the embedded browser window.
// It is on unless settings.json turns it off.
func BrowserAssistEnabled() bool {
	settings := config.NewSettingsStore(config.SettingsFilePath()).Load()
	return settings.BrowserAssistEnabled == nil || *settings.BrowserAssistEnabled
}
