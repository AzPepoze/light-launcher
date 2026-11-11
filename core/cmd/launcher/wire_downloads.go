package main

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite" // pure Go SQLite driver, no cgo

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/downloads"
	"light-launcher/core/internal/library"
	"light-launcher/core/internal/logger"
)

// openDownloads starts the download manager. It keeps its rows in downloads.db, a file of its own,
// because the library opens its database itself.
func (s *services) openDownloads(ctx context.Context, dirs folders, hub *api.Hub) error {
	db, err := openSQLite(ctx, filepath.Join(dirs.Config, "downloads.db"))
	if err != nil {
		return err
	}
	s.downloadsDB = db
	store, err := downloads.NewSQLiteStore(ctx, db)
	if err != nil {
		return err
	}
	manager, err := downloads.NewManager(downloads.Options{
		Store:     store,
		Events:    hub,
		Dir:       dirs.Downloads,
		GamesDir:  dirs.Games,
		Installer: s.installer,
	})
	if err != nil {
		return err
	}
	s.downloads = manager
	return manager.Restore(ctx)
}

func (s *services) closeDownloads() {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := s.downloads.Shutdown(ctx); err != nil {
		logger.Warn("main", "stop downloads: %v", err)
	}
	if err := s.downloadsDB.Close(); err != nil {
		logger.Warn("main", "close downloads db: %v", err)
	}
}

// openSQLite uses the same settings as the library database: WAL journal and a busy timeout.
func openSQLite(ctx context.Context, path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)",
		filepath.ToSlash(path))
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", filepath.Base(path), err)
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("open %s: %w", filepath.Base(path), err)
	}
	return db, nil
}

// archiveSink installs an archive that the watch folder matched to a game.
// The file is already in the downloads folder and has no download row, so it gets download id 0.
type archiveSink struct {
	installer *downloads.FolderInstaller
}

func (a archiveSink) Accept(ctx context.Context, gameID int64, path string) error {
	return a.installer.Install(ctx, downloads.InstallJob{GameID: gameID, Archive: path})
}

// updateInstaller downloads a newer version of a game and installs it.
type updateInstaller struct {
	library   *library.Library // set after the library is open
	downloads *downloads.Manager
}

func (u *updateInstaller) InstallUpdate(ctx context.Context, game library.Game, version string) error {
	files, err := u.library.ListFiles(ctx, game.ID)
	if err != nil {
		return err
	}
	file, ok := pickUpdateFile(files, version)
	if !ok {
		return fmt.Errorf("no download file for %s version %s", game.Title, version)
	}
	_, err = u.downloads.Enqueue(ctx, contract.DownloadRequest{
		GameID:   game.ID,
		URL:      file.URL,
		Filename: valueOf(file.Filename),
		Install:  true,
	})
	return err
}

// pickUpdateFile returns the file whose label or name names the version. Otherwise it returns the only
// file that needs no sign-in. Files that need a sign-in are never picked.
func pickUpdateFile(files []library.GameFile, version string) (library.GameFile, bool) {
	var open []library.GameFile
	for _, file := range files {
		if file.RequiresAuth == 0 {
			open = append(open, file)
		}
	}
	for _, file := range open {
		if version != "" && strings.Contains(file.Label+" "+valueOf(file.Filename), version) {
			return file, true
		}
	}
	if len(open) == 1 {
		return open[0], true
	}
	return library.GameFile{}, false
}

func valueOf(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
