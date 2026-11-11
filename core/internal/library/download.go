package library

import (
	"context"
	"errors"
	"fmt"

	"light-launcher/core/internal/contract"
)

// ErrNoDownloads means the library was opened without a download queue.
var ErrNoDownloads = errors.New("downloads are not available")

// prepareDownloadParams is the source.prepareDownload request.
type prepareDownloadParams struct {
	Source string                 `json:"source"`
	File   contract.FileCandidate `json:"file"`
	Game   GameRef                `json:"game"`
}

// DownloadFile queues one saved file of a game. The file is named by its stored id, so the link
// always comes from the library, never from the caller. A source that can resolve files prepares
// the download through its plugin, which may wait for the person to finish a browser step. Other
// games queue the saved link as it is.
func (l *Library) DownloadFile(ctx context.Context, gameID, fileID int64) (int64, error) {
	game, err := l.Store.Get(ctx, gameID)
	if err != nil {
		return 0, err
	}
	file, err := l.Store.candidate(ctx, gameID, fileID)
	if err != nil {
		return 0, err
	}
	info, ok := l.pluginFor(game.Source)
	if !ok || !hasCapability(info, CapabilityResolveFiles) {
		return l.queueLink(ctx, gameID, file)
	}
	return l.queuePrepared(ctx, info, game, file)
}

// queuePrepared lets the source plugin prepare the file, then queues the answer.
func (l *Library) queuePrepared(ctx context.Context, info contract.SourceInfo, game Game, file contract.FileCandidate) (int64, error) {
	params := prepareDownloadParams{Source: info.SourceID, File: file, Game: gameRefOf(game)}
	id, err := l.sources.StartDownload(ctx, info.Plugin, params, game.ID)
	if err != nil {
		return 0, fmt.Errorf("start download: %w", err)
	}
	return id, nil
}

// queueLink queues a saved link without a plugin. Like a direct link, the archive is installed after it downloads.
func (l *Library) queueLink(ctx context.Context, gameID int64, file contract.FileCandidate) (int64, error) {
	if l.opts.Downloads == nil {
		return 0, ErrNoDownloads
	}
	return l.opts.Downloads.Enqueue(ctx, contract.DownloadRequest{
		GameID:  gameID,
		URL:     file.URL,
		Headers: file.Headers,
		Install: true,
	})
}
