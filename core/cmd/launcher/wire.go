package main

import (
	"context"
	"database/sql"
	"sync"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/browserassist"
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
	lib, err := library.Open(ctx, libraryOptions(dirs, hub, sources, updates, archiveSink{installer: s.installer}, launcher))
	if err != nil {
		return err
	}
	s.library = lib
	updates.library = lib

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

	sched, err := newScheduler(ctx, s.library, host, sources, hub.Emit)
	if err != nil {
		return err
	}
	s.scheduler = sched

	s.registerMethods(router, hub, sources)
	s.startJobs(ctx)
	return nil
}

func (s *services) gameTitle(ctx context.Context, gameID int64) (string, error) {
	game, err := s.library.Get(ctx, gameID)
	return game.Title, err
}

func (s *services) registerMethods(router *api.Router, hub *api.Hub, sources contract.Sources) {
	appsvc.Register(router, hub)
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
	s.jobs.Add(3)
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
