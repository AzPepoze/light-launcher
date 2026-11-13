package library

import (
	"context"
	"errors"
	"fmt"

	"light-launcher/core/internal/contract"
)

// ErrFetchGame means the game's source could not fetch the game, so nothing was refreshed.
var ErrFetchGame = errors.New("the source could not fetch the game")

// ErrNoSourceGameID means a stored plugin game has no source id to ask its plugin about.
var ErrNoSourceGameID = errors.New("game has no source id")

// RefreshGameFiles asks the game's source for its current details and download files, saves both,
// and returns the updated game page. It repairs games saved before plugin file ids existed: the
// new source.resolveFiles answer carries the plugin file ids that source.prepareDownload needs.
//
// The stored identity (source and source id) wins over the reply, and the user's fields survive:
// install folder, exe and arguments, playtime, favorite, hidden, notes, pinned and tags stay as
// the row has them, and an installed row keeps its installed status. Local launch settings are
// not touched either. A source that cannot fetch a game changes nothing and returns ErrFetchGame.
// A source without resolveFiles refreshes only the metadata and keeps its stored file list.
func (l *Library) RefreshGameFiles(ctx context.Context, gameID int64) (GameDetail, error) {
	game, err := l.Get(ctx, gameID)
	if err != nil {
		return GameDetail{}, err
	}
	info, ok := l.pluginFor(game.Source)
	if !ok {
		return GameDetail{}, fmt.Errorf("refresh files for %q: %w: %q", game.Title, ErrUnknownSource, game.Source)
	}
	if game.SourceID == "" {
		return GameDetail{}, fmt.Errorf("refresh files for %q: %w", game.Title, ErrNoSourceGameID)
	}
	dto, err := l.fetchGame(ctx, info, ParsedRef{Kind: "game", ID: game.SourceID, URL: game.URL})
	if err != nil {
		return GameDetail{}, fmt.Errorf("refresh files for %q: %w: %w", game.Title, ErrFetchGame, err)
	}
	if err := l.saveRefreshedGame(ctx, gameID, dto); err != nil {
		return GameDetail{}, fmt.Errorf("refresh files for %q: %w", game.Title, err)
	}
	if err := l.replaceRefreshedFiles(ctx, gameID, info); err != nil {
		return GameDetail{}, fmt.Errorf("refresh files for %q: %w", game.Title, err)
	}
	return l.Detail(ctx, gameID)
}

// saveRefreshedGame writes the source-owned fields of a fresh fetch onto the stored row.
// It reloads the row so an edit made while the plugin worked is not overwritten.
func (l *Library) saveRefreshedGame(ctx context.Context, gameID int64, dto contract.GameDTO) error {
	stored, err := l.Get(ctx, gameID)
	if err != nil {
		return err
	}
	// The row keeps its own identity: a reply can refresh it, never move it to another game.
	dto.Source = stored.Source
	dto.SourceID = stored.SourceID
	if dto.URL == "" {
		dto.URL = stored.URL
	}
	// Tags are the user's: a refresh never replaces them with the source's list.
	dto.Tags = stored.Tags
	// An installed row keeps the status its install gave it, even when the source reports another.
	if stored.InstallDir != nil || stored.Status == "installed" {
		dto.Status = stored.Status
	}
	_, err = l.Upsert(ctx, dto)
	return err
}

// replaceRefreshedFiles stores the source's current file list with its plugin file ids.
// A source without resolveFiles keeps the stored list: it has no new ids to offer.
func (l *Library) replaceRefreshedFiles(ctx context.Context, gameID int64, info contract.SourceInfo) error {
	if !hasCapability(info, CapabilityResolveFiles) {
		return nil
	}
	game, err := l.Get(ctx, gameID)
	if err != nil {
		return err
	}
	files, err := l.resolveFiles(ctx, info, gameRefOf(game))
	if err != nil {
		return fmt.Errorf("resolve files: %w", err)
	}
	// Nil means the source sent no file list, so the stored list is still the best it has.
	if files == nil {
		return nil
	}
	return l.ReplaceFiles(ctx, gameID, files)
}
