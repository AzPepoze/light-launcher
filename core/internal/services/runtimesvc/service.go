package runtimesvc

import (
	"light-launcher/core/internal/activity"
	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

type settingsStore interface {
	Load() types.AppSettings
	Save(settings types.AppSettings) error
}

type gameConfigStore interface {
	Find(executablePath string) *types.LaunchOptions
	Save(options *types.LaunchOptions) error
	Path(name, id string) string
}

type activityStore interface {
	RecordLaunch(options types.LaunchOptions) error
	Sync(sessions []types.RunningSession) ([]types.GameActivity, error)
	List() []types.GameActivity
}

type prefixConfigStore interface {
	BaseDir() string
	Load(prefixName string) types.LaunchOptions
	Save(prefixName string, options types.LaunchOptions) error
}

// service owns the stores and session tracker the runtime methods share.
type service struct {
	settings settingsStore
	games    gameConfigStore
	activity activityStore
	prefixes prefixConfigStore
	sessions *sessionTracker
}

func newService(settings settingsStore, games gameConfigStore, activity activityStore, prefixes prefixConfigStore) *service {
	return &service{
		settings: settings,
		games:    games,
		activity: activity,
		prefixes: prefixes,
		sessions: newSessionTracker(activity),
	}
}

func newDefaultService() *service {
	settings := config.NewSettingsStore(config.SettingsFilePath())
	return newService(
		settings,
		config.NewGameStore(config.GetConfigDirectory()),
		activity.New(config.ActivityFilePath(), settings),
		config.NewPrefixStore(settings),
	)
}
