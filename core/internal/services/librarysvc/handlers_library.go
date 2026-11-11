package librarysvc

import (
	"context"
	"errors"
	"strings"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/library"
)

type updatePayload struct {
	ID         int64     `json:"id"`
	Title      *string   `json:"title"`
	Notes      *string   `json:"notes"`
	Tags       *[]string `json:"tags"`
	Hidden     *bool     `json:"hidden"`
	Favorite   *bool     `json:"favorite"`
	ExePath    *string   `json:"exePath"`
	ExeArgs    *string   `json:"exeArgs"`
	InstallDir *string   `json:"installDir"`
}

// removePayload deletes a game. DeleteFiles also deletes its install folder, when it is inside the games folder.
type removePayload struct {
	ID          int64 `json:"id"`
	DeleteFiles bool  `json:"deleteFiles"`
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

// downloadFilePayload names one saved file of a game. The file id comes from the game's file list.
type downloadFilePayload struct {
	GameID int64 `json:"gameId"`
	FileID int64 `json:"fileId"`
}

// downloadFileResult is the id of the queued download.
type downloadFileResult struct {
	ID int64 `json:"id"`
}

// updateHistoryPayload pages the check history. Before is the id of the last entry already shown.
type updateHistoryPayload struct {
	GameID int64 `json:"gameId"`
	Limit  int   `json:"limit"`
	Before int64 `json:"before"`
}

// list returns one page of games. The sources, tags and stats come from overview.
func (s *service) list(ctx context.Context, filter library.Filter) (library.Page, error) {
	return s.lib.List(ctx, filter)
}

func (s *service) overview(ctx context.Context) (library.Overview, error) {
	return s.lib.Overview(ctx)
}

func (s *service) get(ctx context.Context, in idPayload) (library.GameDetail, error) {
	return s.lib.Detail(ctx, in.ID)
}

func (s *service) update(ctx context.Context, in updatePayload) (library.Game, error) {
	if in.Title != nil && strings.TrimSpace(*in.Title) == "" {
		return library.Game{}, errors.New("title cannot be empty")
	}
	if in.InstallDir != nil && *in.InstallDir != "" {
		if err := s.lib.CheckInstallDir(ctx, in.ID, *in.InstallDir); err != nil {
			return library.Game{}, err
		}
	}
	patch := library.Patch{
		Title:      trimmed(in.Title),
		Notes:      in.Notes,
		Tags:       in.Tags,
		Hidden:     in.Hidden,
		Favorite:   in.Favorite,
		ExePath:    in.ExePath,
		ExeArgs:    in.ExeArgs,
		InstallDir: in.InstallDir,
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

// removeResult says whether the install folder was deleted along with the game.
type removeResult struct {
	OK           bool `json:"ok"`
	DeletedFiles bool `json:"deletedFiles"`
}

// remove deletes a game and keeps a history entry. Install files are deleted only with deleteFiles,
// and only inside the games folder.
func (s *service) remove(ctx context.Context, in removePayload) (removeResult, error) {
	deleted, err := s.lib.RemoveGame(ctx, in.ID, in.DeleteFiles)
	if err != nil {
		return removeResult{}, err
	}
	s.events.Emit("library.removed", map[string]any{"id": in.ID})
	return removeResult{OK: true, DeletedFiles: deleted}, nil
}

func (s *service) search(ctx context.Context, in searchPayload) ([]library.SearchResult, error) {
	return library.SearchSources(ctx, s.sources, in.Query), nil
}

// browse returns one filtered page of a source's games. Nothing is saved.
func (s *service) browse(ctx context.Context, in contract.BrowseRequest) (contract.BrowsePage, error) {
	return s.lib.Browse(ctx, in)
}

// filters returns the values of one dynamic sidebar group, such as a tag list.
func (s *service) filters(ctx context.Context, in contract.FiltersRequest) (contract.FiltersPage, error) {
	return s.lib.ListFilters(ctx, in)
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

// downloadFile queues a saved file of a game through its source plugin when it has one.
func (s *service) downloadFile(ctx context.Context, in downloadFilePayload) (downloadFileResult, error) {
	id, err := s.lib.DownloadFile(ctx, in.GameID, in.FileID)
	if err != nil {
		return downloadFileResult{}, err
	}
	return downloadFileResult{ID: id}, nil
}

func (s *service) checkUpdates(ctx context.Context, in checkPayload) ([]library.CheckResult, error) {
	return s.lib.CheckAll(ctx, in.IDs, !in.Force)
}

func (s *service) updatesSummary(ctx context.Context) (library.UpdatesSummary, error) {
	return s.lib.UpdatesSummary(ctx)
}

func (s *service) updateHistory(ctx context.Context, in updateHistoryPayload) ([]library.UpdateHistoryEntry, error) {
	return s.lib.UpdateHistory(ctx, in.GameID, in.Limit, in.Before)
}

func (s *service) setFavorite(ctx context.Context, in favoritePayload) (library.Game, error) {
	game, err := s.lib.Patch(ctx, in.ID, library.Patch{Favorite: &in.Favorite})
	if err != nil {
		return library.Game{}, err
	}
	s.changed(in.ID)
	return game, nil
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
