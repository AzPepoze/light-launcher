package appsvc

import (
	"path/filepath"
	"testing"

	"light-launcher/core/internal/activity"
	"light-launcher/core/internal/config"
)

// isolateConfigHome points every config and home lookup at a temp folder.
func isolateConfigHome(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("APPDATA", dir)
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}

// newTestService builds the service on temp folders instead of the user's config folder.
func newTestService(t *testing.T) *service {
	t.Helper()
	base := t.TempDir()
	settings := config.NewSettingsStore(filepath.Join(base, "settings.json"))
	return newService(
		settings,
		config.NewGameStore(filepath.Join(base, "config", "executables")),
		activity.New(filepath.Join(base, "activity.json"), settings),
	)
}
