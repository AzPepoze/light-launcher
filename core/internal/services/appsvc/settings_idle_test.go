package appsvc

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/types"
)

func TestDefaultSettingsEnableIdleStop(t *testing.T) {
	settings := defaultSettings()
	if settings.PluginIdleTimeoutMinutes == nil || *settings.PluginIdleTimeoutMinutes != defaultPluginIdleMinutes {
		t.Fatalf("idle timeout = %v, want %d minutes", settings.PluginIdleTimeoutMinutes, defaultPluginIdleMinutes)
	}
}

func TestNormalizeFillsMissingIdleTimeout(t *testing.T) {
	settings := &types.AppSettings{}
	if !normalizeSettings(settings) {
		t.Fatal("normalization did not report the added idle timeout")
	}
	if settings.PluginIdleTimeoutMinutes == nil || *settings.PluginIdleTimeoutMinutes != defaultPluginIdleMinutes {
		t.Fatalf("idle timeout = %v", settings.PluginIdleTimeoutMinutes)
	}

	minutes := maxPluginIdleMinutes + 1
	settings.PluginIdleTimeoutMinutes = &minutes
	if !normalizePluginIdleTimeout(settings) {
		t.Fatal("an out-of-range idle timeout was not replaced")
	}
	if *settings.PluginIdleTimeoutMinutes != defaultPluginIdleMinutes {
		t.Fatalf("replaced idle timeout = %d", *settings.PluginIdleTimeoutMinutes)
	}
}

func TestIdleTimeoutUsesTheDocumentedJSONKey(t *testing.T) {
	minutes := 15
	data, err := json.Marshal(types.AppSettings{PluginIdleTimeoutMinutes: &minutes})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"pluginIdleTimeoutMinutes":15`) {
		t.Fatalf("settings JSON = %s", data)
	}
}

func TestSaveAppSettingsRejectsBadIdleTimeout(t *testing.T) {
	for _, minutes := range []int{-1, maxPluginIdleMinutes + 1} {
		value := minutes
		err := validateAppSettings(types.AppSettings{PluginIdleTimeoutMinutes: &value})
		if err == nil {
			t.Fatalf("%d minutes was accepted", minutes)
		}
		var fieldsErr *api.FieldsError
		if !errors.As(err, &fieldsErr) || len(fieldsErr.Fields) != 1 || fieldsErr.Fields[0].Field != "pluginIdleTimeoutMinutes" {
			t.Fatalf("error for %d = %v", minutes, err)
		}
	}
	for _, minutes := range []int{0, 5, maxPluginIdleMinutes} {
		value := minutes
		if err := validateAppSettings(types.AppSettings{PluginIdleTimeoutMinutes: &value}); err != nil {
			t.Fatalf("%d minutes was rejected: %v", minutes, err)
		}
	}
}

func TestPluginIdleTimeoutReadsTheSetting(t *testing.T) {
	isolateConfigHome(t)
	zero := 0
	if err := saveAppSettings(types.AppSettings{PluginIdleTimeoutMinutes: &zero}); err != nil {
		t.Fatal(err)
	}
	timeout := PluginIdleTimeout()
	if timeout == nil || *timeout != 0 {
		t.Fatalf("PluginIdleTimeout = %v, want zero", timeout)
	}

	minutes := 15
	if err := saveAppSettings(types.AppSettings{PluginIdleTimeoutMinutes: &minutes}); err != nil {
		t.Fatal(err)
	}
	timeout = PluginIdleTimeout()
	if timeout == nil || *timeout != 15*time.Minute {
		t.Fatalf("PluginIdleTimeout = %v, want 15m", timeout)
	}
}
