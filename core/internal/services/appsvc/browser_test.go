package appsvc

import (
	"testing"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

func TestBrowserAssistEnabledDefaultsOn(t *testing.T) {
	isolateConfigHome(t)
	if !BrowserAssistEnabled() {
		t.Fatal("browser assist is off by default")
	}
}

func TestBrowserAssistEnabledCanBeTurnedOff(t *testing.T) {
	isolateConfigHome(t)
	off := false
	settings := config.NewSettingsStore(config.SettingsFilePath())
	if err := settings.Save(types.AppSettings{BrowserAssistEnabled: &off}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if BrowserAssistEnabled() {
		t.Fatal("browser assist kill switch was ignored")
	}
}
