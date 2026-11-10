package librarysvc

import (
	"context"
	"errors"
	"strings"

	"light-launcher/core/internal/library"
)

type updatePayload struct {
	ID       int64     `json:"id"`
	Title    *string   `json:"title"`
	Notes    *string   `json:"notes"`
	Tags     *[]string `json:"tags"`
	Hidden   *bool     `json:"hidden"`
	Favorite *bool     `json:"favorite"`
	ExePath  *string   `json:"exePath"`
	ExeArgs  *string   `json:"exeArgs"`
}

type favoritePayload struct {
	ID       int64 `json:"id"`
	Favorite bool  `json:"favorite"`
}

type searchPayload struct {
	Query string `json:"query"`
}

type urlPayload struct {
	URL string `json:"url"`
}

type linkPayload struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`
}

// checkPayload picks games to check. Without force, games checked within the recheck interval are skipped.
type checkPayload struct {
	IDs   []int64 `json:"ids"`
	Force bool    `json:"force"`
}

type pathPayload struct {
	Path string `json:"path"`
}

func (s *service) list(ctx context.Context, filter library.Filter) (library.Listing, error) {
	return s.lib.Listing(ctx, filter)
}

func (s *service) get(ctx context.Context, in idPayload) (library.GameDetail, error) {
	return s.lib.Detail(ctx, in.ID)
}

func (s *service) update(ctx context.Context, in updatePayload) (library.Game, error) {
	if in.Title != nil && strings.TrimSpace(*in.Title) == "" {
		return library.Game{}, errors.New("title cannot be empty")
	}
	patch := library.Patch{
		Title:    trimmed(in.Title),
		Notes:    in.Notes,
		Tags:     in.Tags,
		Hidden:   in.Hidden,
		Favorite: in.Favorite,
		ExePath:  in.ExePath,
		ExeArgs:  in.ExeArgs,
	}
	game, err := s.lib.Patch(ctx, in.ID, patch)
	if err != nil {
		return library.Game{}, err
	}
	s.changed(in.ID)
	return game, nil
}

func trimmed(value *string) *string {
	if value == nil {
		return nil
	}
	text := strings.TrimSpace(*value)
	return &text
}

// remove deletes a game and keeps a history entry. Files on disk are not touched.
func (s *service) remove(ctx context.Context, in idPayload) (okResult, error) {
	game, err := s.lib.Get(ctx, in.ID)
	if err != nil {
		return okResult{}, err
	}
	if _, err := s.lib.AddDeleted(ctx, deletedInputOf(game)); err != nil {
		return okResult{}, err
	}
	if err := s.lib.Remove(ctx, in.ID); err != nil {
		return okResult{}, err
	}
	s.events.Emit("library.removed", map[string]any{"id": in.ID})
	return okResult{OK: true}, nil
}

func deletedInputOf(game library.Game) library.DeletedInput {
	return library.DeletedInput{
		GameID:     game.ID,
		Source:     game.Source,
		SourceID:   game.SourceID,
		URL:        game.URL,
		Title:      game.Title,
		Version:    derefString(game.Version),
		Developer:  derefString(game.Developer),
		Tags:       game.Tags,
		InstallDir: derefString(game.InstallDir),
	}
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (s *service) search(ctx context.Context, in searchPayload) ([]library.SearchResult, error) {
	return library.SearchSources(ctx, s.sources, in.Query), nil
}

func (s *service) preview(ctx context.Context, in urlPayload) (library.Preview, error) {
	return s.lib.Preview(ctx, in.URL)
}

func (s *service) addFromURL(ctx context.Context, in urlPayload) (library.Game, error) {
	game, err := s.lib.AddFromURL(ctx, in.URL)
	if err != nil {
		return library.Game{}, err
	}
	s.changed(game.ID)
	return game, nil
}

func (s *service) linkLocal(ctx context.Context, in linkPayload) (library.Game, error) {
	game, err := s.lib.LinkLocal(ctx, in.ID, in.URL)
	if err != nil {
		return library.Game{}, err
	}
	s.changed(game.ID)
	return game, nil
}

func (s *service) checkUpdates(ctx context.Context, in checkPayload) ([]library.CheckResult, error) {
	return s.lib.CheckAll(ctx, in.IDs, !in.Force)
}

func (s *service) updatesSummary(ctx context.Context) (library.UpdatesSummary, error) {
	return s.lib.UpdatesSummary(ctx)
}

func (s *service) setFavorite(ctx context.Context, in favoritePayload) (library.Game, error) {
	game, err := s.lib.Patch(ctx, in.ID, library.Patch{Favorite: &in.Favorite})
	if err != nil {
		return library.Game{}, err
	}
	s.changed(in.ID)
	return game, nil
}

func (s *service) importLegacy(ctx context.Context, in pathPayload) (library.ImportReport, error) {
	report, err := s.lib.ImportLegacy(ctx, in.Path)
	if err != nil {
		return library.ImportReport{}, err
	}
	s.events.Emit("library.changed", map[string]any{})
	return report, nil
}

func (s *service) scanFolder(ctx context.Context, in pathPayload) (library.ScanReport, error) {
	return s.lib.ScanFolder(ctx, in.Path)
}

func (s *service) launch(ctx context.Context, in idPayload) (library.LaunchResult, error) {
	return s.lib.Launch(ctx, in.ID)
}

type linkFolderPayload struct {
	ID   int64  `json:"id"`
	Path string `json:"path"`
}

func (s *service) linkFolder(ctx context.Context, in linkFolderPayload) (library.Game, error) {
	game, err := s.lib.LinkFolder(ctx, in.ID, in.Path)
	if err != nil {
		return library.Game{}, err
	}
	s.changed(game.ID)
	return game, nil
}
