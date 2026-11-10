package library

import (
	"context"
	"errors"
	"fmt"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/logger"
)

// ExistingGame tells a preview which library row the link already belongs to.
type ExistingGame struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Installed bool   `json:"installed"`
}

// Preview is a source game fetched for display. Nothing is saved.
type Preview struct {
	Game     contract.GameDTO `json:"game"`
	Existing *ExistingGame    `json:"existing"`
}

// Preview fetches the game behind a link and reports whether it is already in the library.
func (l *Library) Preview(ctx context.Context, url string) (Preview, error) {
	info, ref, err := l.resolveURL(ctx, url)
	if err != nil {
		return Preview{}, err
	}
	dto, err := l.fetchGame(ctx, info, ref)
	if err != nil {
		return Preview{}, err
	}
	preview := Preview{Game: dto}
	existing, err := l.BySource(ctx, dto.Source, dto.SourceID)
	if err == nil {
		preview.Existing = &ExistingGame{ID: existing.ID, Title: existing.Title, Installed: existing.InstallDir != nil}
	} else if !errors.Is(err, ErrNotFound) {
		return Preview{}, err
	}
	return preview, nil
}

// AddFromURL fetches a game from a link and saves it, with its download candidates and cover.
func (l *Library) AddFromURL(ctx context.Context, url string) (Game, error) {
	info, ref, err := l.resolveURL(ctx, url)
	if err != nil {
		return Game{}, err
	}
	dto, err := l.fetchGame(ctx, info, ref)
	if err != nil {
		return Game{}, err
	}
	id, err := l.Upsert(ctx, dto)
	if err != nil {
		return Game{}, err
	}
	game, err := l.Get(ctx, id)
	if err != nil {
		return Game{}, err
	}
	if err := l.refreshFiles(ctx, game); err != nil {
		logger.Warn(logCategory, "could not resolve files for %s: %v", game.Title, err)
	}
	go l.cacheArtwork(game)
	return game, nil
}

// LinkLocal attaches a local game to a source thread. If the thread is already in the library,
// the two rows merge and keep their install data.
func (l *Library) LinkLocal(ctx context.Context, id int64, url string) (Game, error) {
	local, err := l.Get(ctx, id)
	if err != nil {
		return Game{}, err
	}
	info, ref, err := l.resolveURL(ctx, url)
	if err != nil {
		return Game{}, err
	}
	if ref.Kind != "game" {
		return Game{}, fmt.Errorf("that is not a %s game page", info.Plugin)
	}
	dto, err := l.fetchGame(ctx, info, ref)
	if err != nil {
		return Game{}, err
	}
	target, err := l.attachLocal(ctx, local, dto)
	if err != nil {
		return Game{}, err
	}
	if err := l.refreshFiles(ctx, target); err != nil {
		logger.Warn(logCategory, "file refresh failed after link: %v", err)
	}
	go l.cacheArtwork(target)
	return target, nil
}

// attachLocal relinks the local row to the source game, or merges it into the row that already exists.
func (l *Library) attachLocal(ctx context.Context, local Game, dto contract.GameDTO) (Game, error) {
	existing, err := l.BySource(ctx, dto.Source, dto.SourceID)
	if errors.Is(err, ErrNotFound) {
		return l.relinkLocal(ctx, local, dto)
	}
	if err != nil {
		return Game{}, err
	}
	if existing.ID == local.ID {
		return existing, nil
	}
	if _, err := l.Patch(ctx, existing.ID, mergePatch(local, existing)); err != nil {
		return Game{}, err
	}
	if err := l.Remove(ctx, local.ID); err != nil {
		return Game{}, err
	}
	logger.Info(logCategory, "merged local game %d into %s:%s", local.ID, dto.Source, dto.SourceID)
	return l.Get(ctx, existing.ID)
}

func (l *Library) relinkLocal(ctx context.Context, local Game, dto contract.GameDTO) (Game, error) {
	if err := l.Relink(ctx, local.ID, dto); err != nil {
		return Game{}, err
	}
	status := statusOr(dto.Status, "completed")
	installedVersion := derefString(local.InstalledVersion)
	if installedVersion == "" {
		installedVersion = dto.Version
	}
	if local.InstallDir != nil {
		status = "installed"
	}
	return l.Patch(ctx, local.ID, Patch{InstalledVersion: stringOrNil(installedVersion), Status: &status})
}

// mergePatch carries the install data of a local row onto the row it merges into.
func mergePatch(local, target Game) Patch {
	favorite := local.Favorite == 1 || target.Favorite == 1
	playtime := max(local.PlaytimeS, target.PlaytimeS)
	patch := Patch{
		InstallDir:       local.InstallDir,
		InstallSize:      local.InstallSize,
		ExePath:          local.ExePath,
		InstalledVersion: local.InstalledVersion,
		Notes:            local.Notes,
		LastPlayedAt:     local.LastPlayedAt,
		PlaytimeS:        &playtime,
		Favorite:         &favorite,
	}
	if local.InstallDir != nil {
		installed := "installed"
		patch.Status = &installed
	}
	return patch
}

// cacheArtwork downloads the cover and screenshots in the background and tells the UI when done.
func (l *Library) cacheArtwork(game Game) {
	ctx := context.Background()
	if _, err := l.AttachCover(ctx, game); err != nil {
		logger.Debug(logCategory, "cover for %s: %v", game.Title, err)
	}
	l.CacheScreenshots(ctx, game)
	l.notify("library.game", map[string]any{"id": game.ID})
}
