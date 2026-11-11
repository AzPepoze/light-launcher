package appsvc

import (
	"testing"
	"time"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

func TestPluginIdleTimeoutReadsTheSetting(t *testing.T) {
	isolateConfigHome(t)
	settings := config.NewSettingsStore(config.SettingsFilePath())
	zero := 0
	if err := settings.Save(types.AppSettings{PluginIdleTimeoutMinutes: &zero}); err != nil {
		t.Fatal(err)
	}
	timeout := PluginIdleTimeout()
	if timeout == nil || *timeout != 0 {
		t.Fatalf("PluginIdleTimeout = %v, want zero", timeout)
	}

	minutes := 15
	if err := settings.Save(types.AppSettings{PluginIdleTimeoutMinutes: &minutes}); err != nil {
		t.Fatal(err)
	}
	timeout = PluginIdleTimeout()
	if timeout == nil || *timeout != 15*time.Minute {
		t.Fatalf("PluginIdleTimeout = %v, want 15m", timeout)
	}
}
