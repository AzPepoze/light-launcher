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

// writeGameTree makes categories category folders under root, each with gamesPerCategory game
// folders. Each game folder holds a launcher and a 32-bit copy of it.
func writeGameTree(t testing.TB, root string, categories, gamesPerCategory int) {
	t.Helper()
	for c := range categories {
		for g := range gamesPerCategory {
			name := fmt.Sprintf("Game%02d-%03d", c, g)
			dir := filepath.Join(root, fmt.Sprintf("Category%02d", c), name)
			mustWrite(t, filepath.Join(dir, name+".exe"))
			mustWrite(t, filepath.Join(dir, name+"32.exe"))
		}
	}
}
