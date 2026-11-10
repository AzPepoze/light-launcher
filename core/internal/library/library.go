package library

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"light-launcher/core/internal/contract"
)

// Launcher starts a game and blocks until it exits. args is the raw argument line stored on the game.
type Launcher interface {
	Launch(ctx context.Context, exePath, args, name string) error
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
	Sources     contract.Sources
	Launcher    Launcher
	Installer   Installer
	Sink        ArchiveSink
	// AutoUpdate reports whether a game should be installed as soon as an update is found.
	AutoUpdate func(Game) bool
	// UpdateRecheck skips games checked more recently than this in scheduled runs. Zero checks every game.
	UpdateRecheck time.Duration
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
	}, nil
}

// Close closes the database. Running launches keep running, but their playtime is lost if the app exits first.
func (l *Library) Close() error {
	return l.db.Close()
}

func (l *Library) notify(channel string, data any) {
	if l.opts.Notify != nil {
		l.opts.Notify(channel, data)
	}
}

func (l *Library) watchFolder() string {
	if l.opts.WatchFolder == nil {
		return ""
	}
	return l.opts.WatchFolder()
}
