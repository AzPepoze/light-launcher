package appsvc

import (
	"fmt"
	"path/filepath"
	"testing"

	"light-launcher/core/internal/activity"
	"light-launcher/core/internal/config"
)

// isolateConfigHome points every config and home lookup at a temp folder.
func isolateConfigHome(t testing.TB) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("APPDATA", dir)
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}

// newTestService builds the service on temp folders instead of the user's config folder.
func newTestService(t testing.TB) *service {
	t.Helper()
	base := t.TempDir()
	settings := config.NewSettingsStore(filepath.Join(base, "settings.json"))
	return newService(
		settings,
		config.NewGameStore(filepath.Join(base, "config", "executables")),
		activity.New(filepath.Join(base, "activity.json"), settings),
		newLauncherPreferences(filepath.Join(base, "launcher-preferences.json")),
	)
}

// writeGameTree makes folders game folders under root. Each holds Game<i>.exe and a 32-bit Game<i>32.exe.
func writeGameTree(t testing.TB, root string, folders int) {
	t.Helper()
	for i := range folders {
		name := fmt.Sprintf("Game%04d", i)
		mustWrite(t, filepath.Join(root, name, name+".exe"))
		mustWrite(t, filepath.Join(root, name, name+"32.exe"))
	}
}
