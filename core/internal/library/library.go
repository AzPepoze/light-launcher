package library

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/types"
)

// Launcher starts a game and blocks until it exits. args is the raw argument line stored on the game.
type Launcher interface {
	Launch(ctx context.Context, exePath, args, name string) error
}

// ConfigLauncher starts a local game with its saved prefix, Proton and argument settings.
// It blocks until the game exits, so the library can add the playtime.
type ConfigLauncher interface {
	LaunchConfig(ctx context.Context, options types.LaunchOptions) error
}

// Installer installs a newer version of a game. The download side is implemented by the host.
type Installer interface {
	InstallUpdate(ctx context.Context, game Game, version string) error
}

// ArchiveSink takes a finished archive that the watch folder matched to a game.
// The host downloader implements it.
type ArchiveSink interface {
	Accept(ctx context.Context, gameID int64, path string) error
}

// Options configures a Library. Optional fields may be left zero.
type Options struct {
	DBPath       string
	MediaDir     string
	DownloadsDir string
	// WatchFolder returns the current watch folder. It is read on every poll, so changes apply without a restart.
	WatchFolder func() string
	// GamesDir returns the folder that holds installed games. Only install folders inside it are ever deleted.
	GamesDir func() string
	Sources  contract.Sources
	// Downloads queues saved links of games whose source cannot prepare them.
	Downloads contract.Downloads
	Launcher  Launcher
	// ConfigLauncher starts local rows that carry LaunchOptions. Nil means local rows cannot start.
	ConfigLauncher ConfigLauncher
	Installer      Installer
	Sink           ArchiveSink
	// AutoUpdate reports whether a game should be installed as soon as an update is found.
	AutoUpdate func(Game) bool
	// UpdateRule returns how background checks treat the games of a source. It is read on every run,
	// so Settings changes apply without a restart. Nil checks every game of every source.
	UpdateRule func(source string) UpdateRule
	// UpdateSchedule returns when background runs start. It is read on every tick. Nil means no runs.
	UpdateSchedule func() UpdateSchedule
	// Notify pushes background changes to the UI, such as new updates or watched files.
	Notify func(channel string, data any)
}

// Library is the game library: storage plus the background jobs around it.
// Store methods are promoted, so callers use lib.List, lib.Get and so on.
type Library struct {
	*Store
	db       *sql.DB
	opts     Options
	sources  contract.Sources
	media    *mediaCache
	launches *launchSet
	watch    *watchState
	artwork  *artworkQueue
	imports  *importer
	overview overviewCache
	// activeCheck is the update check in progress, if any.
	activeCheck activeCheck
	// now is the clock of the update schedule. Tests replace it.
	now func() time.Time
}

// Open creates or opens the library database and prepares the background jobs.
func Open(ctx context.Context, opts Options) (*Library, error) {
	if opts.DBPath == "" {
		return nil, errors.New("library: DBPath is required")
	}
	db, err := openDB(ctx, opts.DBPath)
	if err != nil {
		return nil, err
	}
	sources := opts.Sources
	if sources == nil {
		sources = noSources{}
	}
	return &Library{
		Store:    NewStore(db),
		db:       db,
		opts:     opts,
		sources:  sources,
		media:    newMediaCache(opts.MediaDir),
		launches: newLaunchSet(),
		watch:    newWatchState(),
		artwork:  newArtworkQueue(),
		imports:  newImporter(),
		now:      time.Now,
	}, nil
}

// Close stops imports and artwork downloads, then closes the database. Running launches keep running, but their playtime is lost if the app exits first.
func (l *Library) Close() error {
	l.imports.stop()
	l.artwork.stop()
	return l.db.Close()
}

func (l *Library) notify(channel string, data any) {
	if l.opts.Notify != nil {
		l.opts.Notify(channel, data)
	}
}

// Get returns one game with its running state and local cover path set.
func (l *Library) Get(ctx context.Context, id int64) (Game, error) {
	game, err := l.Store.Get(ctx, id)
	if err != nil {
		return Game{}, err
	}
	l.decorate(&game)
	return game, nil
}

// Patch is Store.Patch, returning the decorated row.
func (l *Library) Patch(ctx context.Context, id int64, p Patch) (Game, error) {
	game, err := l.Store.Patch(ctx, id, p)
	if err != nil {
		return Game{}, err
	}
	l.decorate(&game)
	return game, nil
}

// List is Store.List, with every returned row decorated.
func (l *Library) List(ctx context.Context, f Filter) (Page, error) {
	page, err := l.Store.List(ctx, f)
	if err != nil {
		return Page{}, err
	}
	for i := range page.Items {
		l.decorate(&page.Items[i])
	}
	return page, nil
}

// decorate fills the fields that depend on the running process, not only on the stored row.
func (l *Library) decorate(game *Game) {
	game.Running = l.launches.isRunning(game.ID)
	game.CoverLocal = l.media.localPath(game.CoverPath)
	game.ScreenshotsLocal = l.media.localPaths(game.Screenshots, "screenshots")
	game.DescriptionText = plainText(game.Description)
}

func (l *Library) watchFolder() string {
	if l.opts.WatchFolder == nil {
		return ""
	}
	return l.opts.WatchFolder()
}
