// Package contract holds the types shared by the plugin host, library and downloader.
package contract

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
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

// DecryptAlgorithmAES128CTR is the only decryption the host applies to a prepared download.
const DecryptAlgorithmAES128CTR = "aes-128-ctr"

// DecryptSpec tells the downloader how to decrypt a finished file. Key and Nonce are base64:
// the key is 16 bytes and the nonce is 8 bytes. AES-CTR counts from zero after the nonce.
// The key belongs to one download and must never be written to a log.
type DecryptSpec struct {
	Algorithm string `json:"algorithm"`
	Key       string `json:"key"`
	Nonce     string `json:"nonce"`
}

// ErrBadDecrypt means a decrypt spec is malformed or names an algorithm the host does not support.
var ErrBadDecrypt = errors.New("bad decrypt spec")

// String redacts the key and nonce, so printing a spec by accident cannot leak the download key.
func (d DecryptSpec) String() string {
	return fmt.Sprintf("decrypt{algorithm:%s key:redacted nonce:redacted}", d.Algorithm)
}

// Decode checks the spec and returns the raw key and nonce. Its errors never contain either value.
func (d *DecryptSpec) Decode() (key, nonce []byte, err error) {
	if d == nil {
		return nil, nil, fmt.Errorf("%w: decrypt spec is missing", ErrBadDecrypt)
	}
	if d.Algorithm != DecryptAlgorithmAES128CTR {
		return nil, nil, fmt.Errorf("%w: algorithm %q is not supported", ErrBadDecrypt, d.Algorithm)
	}
	key, err = base64.StdEncoding.DecodeString(d.Key)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: key is not standard base64", ErrBadDecrypt)
	}
	if len(key) != 16 {
		return nil, nil, fmt.Errorf("%w: key has %d bytes, want 16", ErrBadDecrypt, len(key))
	}
	nonce, err = base64.StdEncoding.DecodeString(d.Nonce)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: nonce is not standard base64", ErrBadDecrypt)
	}
	if len(nonce) != 8 {
		return nil, nil, fmt.Errorf("%w: nonce has %d bytes, want 8", ErrBadDecrypt, len(nonce))
	}
	return key, nonce, nil
}

// Validate reports whether the host can apply the decrypt spec. A nil spec is valid and means
// the finished file is not decrypted.
func (d *DecryptSpec) Validate() error {
	if d == nil {
		return nil
	}
	_, _, err := d.Decode()
	return err
}

// DownloadRequest asks the host downloader to fetch one file for a library game.
type DownloadRequest struct {
	GameID   int64             `json:"gameId"`
	URL      string            `json:"url"`
	Headers  map[string]string `json:"headers,omitempty"`
	Filename string            `json:"filename,omitempty"`
	Install  bool              `json:"install"`
	// Decrypt, when set, describes how the host decrypts the finished file before an install.
	Decrypt *DecryptSpec `json:"decrypt,omitempty"`

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

// ErrSecretsNotListable is returned by a Secrets store that cannot enumerate a plugin's keys.
// Callers fall back to deleting the keys they have recorded.
var ErrSecretsNotListable = errors.New("secret store cannot list keys")

// Secrets is implemented by the host secret store (OS keyring with an encrypted-file fallback).
type Secrets interface {
	Get(ctx context.Context, plugin, key string) (value string, found bool, err error)
	Set(ctx context.Context, plugin, key, value string) error
	Delete(ctx context.Context, plugin, key string) error
	// DeleteAll removes every secret stored for plugin, including keys the caller never named.
	// A store that cannot list a plugin's keys returns ErrSecretsNotListable and changes nothing.
	DeleteAll(ctx context.Context, plugin string) error
}

// SourceInfo describes one game source contributed by a plugin.
type SourceInfo struct {
	Plugin       string   `json:"plugin"`
	SourceID     string   `json:"sourceId"`
	Label        string   `json:"label"`
	Capabilities []string `json:"capabilities"`
	// Tile is the card shape the source declares: poster, wide or square. Kept for older consumers.
	Tile string `json:"tile,omitempty"`
	// Cover is the card box as "W:H", from the manifest cover or from Tile.
	Cover string `json:"cover"`
	// Actions are the buttons the source adds to the Library, with the default scope filled in.
	Actions []SourceAction `json:"actions,omitempty"`
	// Card names the game fields the source's cards show. Nil uses the defaults.
	Card *SourceCard `json:"card,omitempty"`
	// Sidebar is the filter groups the source declares. Empty when it declares none.
	Sidebar []SidebarGroup `json:"sidebar,omitempty"`
	// GamesFolder is the resolved folder the plugin may scan or install into. Empty when the plugin
	// does not hold the $GAMES grant.
	GamesFolder string `json:"gamesFolder,omitempty"`
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
	// Browse runs source.browse and ListFilters runs source.listFilters, each with its own time limit.
	Browse(ctx context.Context, plugin string, request BrowseRequest) (BrowsePage, error)
	ListFilters(ctx context.Context, plugin string, request FiltersRequest) (FiltersPage, error)
	// CallAction runs an action method the plugin declares in its manifest. The name need not start with "source.".
	CallAction(ctx context.Context, plugin, method string, params any, out any) error
}
