package main

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"light-launcher/core/internal/activity"
	"light-launcher/core/internal/api"
	"light-launcher/core/internal/browserassist"
	"light-launcher/core/internal/config"
	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/downloads"
	"light-launcher/core/internal/library"
	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/plugins"
	"light-launcher/core/internal/scheduler"
	"light-launcher/core/internal/services/appsvc"
	"light-launcher/core/internal/services/downloadsvc"
	"light-launcher/core/internal/services/librarysvc"
	"light-launcher/core/internal/services/pluginsvc"
	"light-launcher/core/internal/services/runtimesvc"
	"light-launcher/core/internal/services/schedulersvc"
	"light-launcher/core/internal/services/systemsvc"
	"light-launcher/core/internal/types"
	"light-launcher/core/internal/uploads"
)

// services owns the long-lived parts of one launcher run: databases, download manager, plugins and jobs.
type services struct {
	library       *library.Library
	downloads     *downloads.Manager
	downloadsDB   *sql.DB
	host          *plugins.Host
	browserAssist *browserassist.Service
	installer     *downloads.FolderInstaller
	uploads       *uploads.Store
	scheduler     *scheduler.Scheduler
	rebuildJobs   func() error
	hub           *api.Hub
	stop          context.CancelFunc
	jobs          sync.WaitGroup
}

// wireServices builds every service and registers its methods on router. Call Close when done.
func wireServices(ctx context.Context, router *api.Router, hub *api.Hub) (*services, error) {
	dirs, err := prepareFolders()
	if err != nil {
		return nil, err
	}
	jobCtx, stop := context.WithCancel(ctx)
	s := &services{stop: stop}
	if err := s.open(jobCtx, router, hub, dirs); err != nil {
		s.Close()
		return nil, err
	}
	return s, nil
}

// open creates the services in dependency order. The library and the plugin host refer to each other,
// so the plugin source list is filled in after both exist.
func (s *services) open(ctx context.Context, router *api.Router, hub *api.Hub, dirs folders) error {
	uploadStore, err := uploads.Open(dirs.Uploads)
	if err != nil {
		return err
	}
	s.uploads = uploadStore
	s.installer = &downloads.FolderInstaller{Root: dirs.Games, Events: hub, Titles: s.gameTitle}
	if err := s.openDownloads(ctx, dirs, hub); err != nil {
		return err
	}
	sources := &lateSources{}
	updates := &updateInstaller{downloads: s.downloads}
	launcher := gameLauncher{router: router}
	lib, err := library.Open(ctx, libraryOptions(dirs, hub, sources, s.downloads, updates, archiveSink{installer: s.installer}, launcher))
	if err != nil {
		return err
	}
	s.library = lib
	updates.library = lib
	if err := backfillLocal(ctx, lib); err != nil {
		logger.Warn("main", "could not import local games into the library: %v", err)
	}

	browser, err := browserassist.New(browserassist.Options{
		ProfilesDir: dirs.BrowserProfiles,
		Enabled:     appsvc.BrowserAssistEnabled,
	})
	if err != nil {
		return err
	}
	s.browserAssist = browser
	host, err := newPluginHost(dirs, lib, s.downloads, browser)
	if err != nil {
		return err
	}
	s.host = host
	sources.registry = plugins.NewRegistry(host)

	sched, rebuild, err := newScheduler(ctx, s.library, host, sources, libraryImporter(s.library), hub.Emit)
	if err != nil {
		return err
	}
	s.scheduler = sched
	s.rebuildJobs = rebuild
	s.hub = hub

	s.registerMethods(router, hub, sources)
	s.startJobs(ctx)
	return nil
}

func (s *services) gameTitle(ctx context.Context, gameID int64) (string, error) {
	game, err := s.library.Get(ctx, gameID)
	return game.Title, err
}

func (s *services) registerMethods(router *api.Router, hub *api.Hub, sources contract.Sources) {
	appsvc.RegisterWithLocalView(router, hub, func(ctx context.Context, groups []types.ScannedFolderGroup) error {
		_, err := s.library.SyncScanned(ctx, groups)
		return err
	}, s.library)
	runtimesvc.Register(router, hub)
	systemsvc.Register(router, hub)
	pluginsvc.Register(router, hub, s.host, s.uploads)
	librarysvc.Register(router, hub, s.library, sources)
	downloadsvc.Register(router, hub, s.downloads)
	schedulersvc.Register(router, hub, s.scheduler)
	logger.Info("api", "registered %d methods", len(router.Methods()))
}

// startJobs runs the watch folder poller, the background scheduler and the plugin release check until Close.
func (s *services) startJobs(ctx context.Context) {
	s.jobs.Add(4)
	go func() {
		defer s.jobs.Done()
		s.library.RunWatcher(ctx, watchInterval)
	}()
	go func() {
		defer s.jobs.Done()
		s.scheduler.Run(ctx)
	}()
	go func() {
		defer s.jobs.Done()
		s.runPluginUpdateCheck(ctx)
	}()
	go func() {
		defer s.jobs.Done()
		s.watchJobChanges(ctx)
	}()
}

// watchJobChanges rebuilds the scheduler jobs when a plugin is installed, updated, enabled or its
// settings change. A short debounce collapses the burst of events one install sends.
func (s *services) watchJobChanges(ctx context.Context) {
	if s.hub == nil || s.rebuildJobs == nil {
		return
	}
	events, cancel := s.hub.Subscribe()
	defer cancel()
	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		<-timer.C
	}
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			if event.Channel != "plugin-state" && event.Channel != "plugin-settings" {
				continue
			}
			timer.Reset(time.Second)
		case <-timer.C:
			if err := s.rebuildJobs(); err != nil {
				logger.Warn("scheduler", "could not apply plugin job changes: %v", err)
			}
		}
	}
}

// Close stops the jobs, the browser window, the plugins and the downloads, then closes the databases.
func (s *services) Close() {
	if s.stop != nil {
		s.stop()
	}
	s.jobs.Wait()
	if s.browserAssist != nil {
		s.browserAssist.Close()
	}
	if s.host != nil {
		s.host.Shutdown()
	}
	if s.downloads != nil {
		s.closeDownloads()
	}
	if s.library != nil {
		if err := s.library.Close(); err != nil {
			logger.Warn("main", "close library: %v", err)
		}
	}
}

// lateSources is the plugin source list. The library needs it before the plugin host exists.
type lateSources struct {
	registry contract.Sources
}

func (l *lateSources) Sources() []contract.SourceInfo {
	return l.registry.Sources()
}

func (l *lateSources) Call(ctx context.Context, plugin, method string, params any, out any) error {
	return l.registry.Call(ctx, plugin, method, params, out)
}

func (l *lateSources) StartDownload(ctx context.Context, plugin string, params any, gameID int64) (int64, error) {
	return l.registry.StartDownload(ctx, plugin, params, gameID)
}

func (l *lateSources) Browse(ctx context.Context, plugin string, request contract.BrowseRequest) (contract.BrowsePage, error) {
	return l.registry.Browse(ctx, plugin, request)
}

func (l *lateSources) ListFilters(ctx context.Context, plugin string, request contract.FiltersRequest) (contract.FiltersPage, error) {
	return l.registry.ListFilters(ctx, plugin, request)
}

func (l *lateSources) CallAction(ctx context.Context, plugin, method string, params any, out any) error {
	return l.registry.CallAction(ctx, plugin, method, params, out)
}

// backfillLocal imports the old JSON game configs, ignored paths and playtime into the library once.
func backfillLocal(ctx context.Context, lib *library.Library) error {
	settings := config.NewSettingsStore(config.SettingsFilePath())
	current := settings.Load()
	configs := config.NewGameStore(config.GetConfigDirectory()).List()
	activityStore := activity.New(config.ActivityFilePath(), settings)
	_, err := lib.BackfillLocal(ctx, library.LocalBackfillInput{
		Configs:     configs,
		ScanFolders: current.ScanFolders,
		Blacklist:   current.Blacklist,
		Activity:    activityStore.List(),
	})
	return err
}
