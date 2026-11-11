// Package contract holds the types shared by the plugin host, library and downloader.
package contract

import (
	"context"
	"errors"
	"time"
)

type Platform string

const (
	PlatformWindows Platform = "windows"
	PlatformLinux   Platform = "linux"
	PlatformMac     Platform = "mac"
	PlatformUnknown Platform = "unknown"
)

// GameDTO is a game as a plugin reports it. It matches the SDK GameDTO.
type GameDTO struct {
	Source          string         `json:"source"`
	SourceID        string         `json:"sourceId"`
	URL             string         `json:"url"`
	Title           string         `json:"title"`
	Version         string         `json:"version,omitempty"`
	Developer       string         `json:"developer,omitempty"`
	Engine          string         `json:"engine,omitempty"`
	Tags            []string       `json:"tags,omitempty"`
	CoverURL        string         `json:"coverUrl,omitempty"`
	Screenshots     []string       `json:"screenshots,omitempty"`
	DescriptionHTML string         `json:"descriptionHtml,omitempty"`
	Changelog       string         `json:"changelog,omitempty"`
	Rating          *float64       `json:"rating,omitempty"`
	Status          string         `json:"status,omitempty"`
	UpdatedAt       string         `json:"updatedAt,omitempty"`
	Extra           map[string]any `json:"extra,omitempty"`
}

// FileCandidate is one downloadable file a plugin found for a game.
type FileCandidate struct {
	ID           string            `json:"id"`
	Label        string            `json:"label"`
	URL          string            `json:"url"`
	Filename     string            `json:"filename,omitempty"`
	Size         int64             `json:"size,omitempty"`
	Host         string            `json:"host,omitempty"`
	Kind         string            `json:"kind"`
	Platform     Platform          `json:"platform,omitempty"`
	RequiresAuth bool              `json:"requiresAuth,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
}

// Download roots. A request names one of them; the host turns the name into a folder.
const (
	RootDownloads = "downloads"
	RootGames     = "games"
)

// ErrBadFolder means a requested folder is not allowed: a bad path, an unknown root or a link.
var ErrBadFolder = errors.New("bad folder")

// ErrBadRequest means a download request has options that do not fit together.
var ErrBadRequest = errors.New("bad request")

// DownloadRequest asks the host downloader to fetch one file for a library game.
type DownloadRequest struct {
	GameID   int64             `json:"gameId"`
	URL      string            `json:"url"`
	Headers  map[string]string `json:"headers,omitempty"`
	Filename string            `json:"filename,omitempty"`
	Install  bool              `json:"install"`

	// Root is RootDownloads (the default) or RootGames. Subfolder is a relative folder under Root.
	Root      string `json:"root,omitempty"`
	Subfolder string `json:"subfolder,omitempty"`
	// InstallSubfolder is a relative folder under the games folder for the installed game. It needs Install.
	InstallSubfolder string `json:"installSubfolder,omitempty"`
	// DeleteArchive, when set, decides whether a successful install deletes the archive. Nil uses the settings.
	DeleteArchive *bool `json:"deleteArchive,omitempty"`
}

// Downloads is implemented by the host download manager.
type Downloads interface {
	Enqueue(ctx context.Context, request DownloadRequest) (downloadID int64, err error)
}

// Library is implemented by the host game library.
type Library interface {
	Upsert(ctx context.Context, game GameDTO) (gameID int64, err error)
	// UpsertMany saves games in one transaction and returns the game ids in the same order.
	UpsertMany(ctx context.Context, games []GameDTO) (gameIDs []int64, err error)
}

// Secrets is implemented by the host secret store (OS keyring with an encrypted-file fallback).
type Secrets interface {
	Get(ctx context.Context, plugin, key string) (value string, found bool, err error)
	Set(ctx context.Context, plugin, key, value string) error
	Delete(ctx context.Context, plugin, key string) error
}

// SourceInfo describes one game source contributed by a plugin.
type SourceInfo struct {
	Plugin       string   `json:"plugin"`
	SourceID     string   `json:"sourceId"`
	Label        string   `json:"label"`
	Capabilities []string `json:"capabilities"`
	// Tile is the card shape the source declares: poster, wide or square.
	Tile string `json:"tile,omitempty"`
}

// PrepareDownloadTimeout is the longest source.prepareDownload may take. A person can finish a browser
// step in that time; the host browser.assist caps at 10 minutes, so this leaves a minute of margin.
const PrepareDownloadTimeout = 11 * time.Minute

// Sources is implemented by the plugin host. Call runs a plugin method such as "source.fetchGame"
// and decodes the JSON result into out. StartDownload asks a plugin to prepare a file
// (source.prepareDownload) and queues the answer with the downloader.
type Sources interface {
	Sources() []SourceInfo
	Call(ctx context.Context, plugin, method string, params any, out any) error
	StartDownload(ctx context.Context, plugin string, params any, gameID int64) (downloadID int64, err error)
}
