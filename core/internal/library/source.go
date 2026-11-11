package library

import (
	"context"
	"errors"
	"fmt"

	"light-launcher/core/internal/contract"
)

// Capabilities a source plugin lists in contract.SourceInfo.Capabilities.
const (
	CapabilityBrowse       = "browse"
	CapabilityUpdateCheck  = "updateCheck"
	CapabilityResolveFiles = "resolveFiles"
	CapabilityListGames    = "listGames"
)

// Plugin methods the library calls. Params and results mirror the old provider types.
const (
	methodParseRef     = "source.parseRef"
	methodFetchGame    = "source.fetchGame"
	methodSearch       = "source.search"
	methodCheckUpdate  = "source.checkUpdate"
	methodResolveFiles = "source.resolveFiles"
	methodListGames    = "source.listGames"
)

// ErrUnknownURL means no plugin recognises a link.
var ErrUnknownURL = errors.New("no source recognises that URL")

// GameRef identifies a stored game in a plugin call. ID is omitted for games the launcher has not saved.
type GameRef struct {
	ID        int64          `json:"id,omitempty"`
	Source    string         `json:"source"`
	SourceID  string         `json:"sourceId"`
	URL       string         `json:"url"`
	Title     string         `json:"title"`
	Version   *string        `json:"version"`
	UpdatedAt *string        `json:"updatedAt"`
	Extra     map[string]any `json:"extra"`
}

// ParsedRef is what source.parseRef returns for a link.
type ParsedRef struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
	URL  string `json:"url"`
}

// Params and replies of the source.* methods, as the plugin SDK defines them.
// Every call names the source that owns the game, so the plugin knows which of its sources to use.

type fetchGameParams struct {
	Source string    `json:"source"`
	Ref    ParsedRef `json:"ref"`
}

type fetchGameReply struct {
	Game contract.GameDTO `json:"game"`
}

type searchParams struct {
	Source string `json:"source"`
	Query  string `json:"query"`
}

type searchReply struct {
	Games []contract.GameDTO `json:"games"`
}

type gameParams struct {
	Source string  `json:"source"`
	Game   GameRef `json:"game"`
}

type filesReply struct {
	Files []contract.FileCandidate `json:"files"`
}

type updateReply struct {
	Update *UpdateInfo `json:"update"`
}

// UpdateInfo is what source.checkUpdate returns for a game.
type UpdateInfo struct {
	Version   string `json:"version,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	Changelog string `json:"changelog,omitempty"`
}

// noSources is used when the host passes no plugin host.
type noSources struct{}

func (noSources) Sources() []contract.SourceInfo { return nil }

func (noSources) Call(context.Context, string, string, any, any) error {
	return errors.New("no plugin host configured")
}

func (noSources) StartDownload(context.Context, string, any, int64) (int64, error) {
	return 0, errors.New("no plugin host configured")
}

func (noSources) Browse(context.Context, string, contract.BrowseRequest) (contract.BrowsePage, error) {
	return contract.BrowsePage{}, errors.New("no plugin host configured")
}

func (noSources) ListFilters(context.Context, string, contract.FiltersRequest) (contract.FiltersPage, error) {
	return contract.FiltersPage{}, errors.New("no plugin host configured")
}

// pluginFor returns the plugin that serves a source id, such as "f95".
func (l *Library) pluginFor(sourceID string) (contract.SourceInfo, bool) {
	for _, info := range l.sources.Sources() {
		if info.SourceID == sourceID {
			return info, true
		}
	}
	return contract.SourceInfo{}, false
}

func hasCapability(info contract.SourceInfo, capability string) bool {
	for _, c := range info.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}

// resolveURL asks each plugin in turn whether it owns the link.
func (l *Library) resolveURL(ctx context.Context, url string) (contract.SourceInfo, ParsedRef, error) {
	asked := map[string]bool{}
	for _, info := range l.sources.Sources() {
		if asked[info.Plugin] {
			continue
		}
		asked[info.Plugin] = true
		var ref *ParsedRef
		if err := l.sources.Call(ctx, info.Plugin, methodParseRef, map[string]string{"url": url}, &ref); err != nil {
			continue
		}
		if ref != nil {
			return info, *ref, nil
		}
	}
	return contract.SourceInfo{}, ParsedRef{}, ErrUnknownURL
}

func (l *Library) fetchGame(ctx context.Context, info contract.SourceInfo, ref ParsedRef) (contract.GameDTO, error) {
	var reply fetchGameReply
	params := fetchGameParams{Source: info.SourceID, Ref: ref}
	if err := l.sources.Call(ctx, info.Plugin, methodFetchGame, params, &reply); err != nil {
		return contract.GameDTO{}, fmt.Errorf("fetch game: %w", err)
	}
	return reply.Game, nil
}

// resolveFiles asks the source for the download candidates of a game, when the source can.
// It returns nil, nil for sources without the resolveFiles capability.
func (l *Library) resolveFiles(ctx context.Context, info contract.SourceInfo, game GameRef) ([]contract.FileCandidate, error) {
	if !hasCapability(info, CapabilityResolveFiles) {
		return nil, nil
	}
	var reply filesReply
	if err := l.sources.Call(ctx, info.Plugin, methodResolveFiles, gameParams{Source: info.SourceID, Game: game}, &reply); err != nil {
		return nil, err
	}
	return reply.Files, nil
}

// refreshFiles stores the download candidates a plugin reports for a game, when it can.
func (l *Library) refreshFiles(ctx context.Context, game Game) error {
	info, ok := l.pluginFor(game.Source)
	if !ok {
		return nil
	}
	candidates, err := l.resolveFiles(ctx, info, gameRefOf(game))
	if err != nil || candidates == nil {
		return err
	}
	return l.ReplaceFiles(ctx, game.ID, candidates)
}

// gameRefOfDTO builds the reference for a source game that is not saved yet.
func gameRefOfDTO(dto contract.GameDTO) GameRef {
	return GameRef{
		Source:    dto.Source,
		SourceID:  dto.SourceID,
		URL:       dto.URL,
		Title:     dto.Title,
		Version:   stringOrNil(dto.Version),
		UpdatedAt: stringOrNil(dto.UpdatedAt),
		Extra:     dto.Extra,
	}
}

func gameRefOf(game Game) GameRef {
	return GameRef{
		ID:        game.ID,
		Source:    game.Source,
		SourceID:  game.SourceID,
		URL:       game.URL,
		Title:     game.Title,
		Version:   game.Version,
		UpdatedAt: game.SourceUpdatedAt,
		Extra:     game.Extra,
	}
}
