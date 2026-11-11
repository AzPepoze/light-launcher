package appsvc

import (
	"light-launcher/core/internal/activity"
	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

type settingsStore interface {
	Load() types.AppSettings
	Save(types.AppSettings) error
	Update(change func(*types.AppSettings) error) error
}

type gameConfigStore interface {
	List() []types.LaunchOptions
	Find(executablePath string) *types.LaunchOptions
	Save(options *types.LaunchOptions) error
	Path(name, id string) string
}

type activityStore interface {
	List() []types.GameActivity
}

// service owns the stores the UI methods share.
type service struct {
	settings    settingsStore
	games       gameConfigStore
	activity    activityStore
	preferences preferenceStore
}

func newService(settings settingsStore, games gameConfigStore, activity activityStore, preferences preferenceStore) *service {
	return &service{settings: settings, games: games, activity: activity, preferences: preferences}
}

func newDefaultService() *service {
	settings := config.NewSettingsStore(config.SettingsFilePath())
	return newService(
		settings,
		config.NewGameStore(config.GetConfigDirectory()),
		activity.New(config.ActivityFilePath(), settings),
		newLauncherPreferences(launcherPreferencesPath()),
	)
}
