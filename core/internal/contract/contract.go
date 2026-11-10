// Package contract holds the types shared by the plugin host, library and downloader.
package contract

import "context"

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

// DownloadRequest asks the host downloader to fetch one file for a library game.
type DownloadRequest struct {
	GameID   int64             `json:"gameId"`
	URL      string            `json:"url"`
	Headers  map[string]string `json:"headers,omitempty"`
	Filename string            `json:"filename,omitempty"`
	Install  bool              `json:"install"`
}

// Downloads is implemented by the host download manager.
type Downloads interface {
	Enqueue(ctx context.Context, request DownloadRequest) (downloadID int64, err error)
}

// Library is implemented by the host game library.
type Library interface {
	Upsert(ctx context.Context, game GameDTO) (gameID int64, err error)
}

// Secrets is implemented by the host secret store (OS keyring with an encrypted-file fallback).
type Secrets interface {
	Get(ctx context.Context, plugin, key string) (value string, found bool, err error)
	Set(ctx context.Context, plugin, key, value string) error
	Delete(ctx context.Context, plugin, key string) error
}
