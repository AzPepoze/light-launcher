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
)

// Plugin methods the library calls. Params and results mirror the old provider types.
const (
	methodParseRef     = "source.parseRef"
	methodFetchGame    = "source.fetchGame"
	methodSearch       = "source.search"
	methodCheckUpdate  = "source.checkUpdate"
	methodResolveFiles = "source.resolveFiles"
)

// ErrUnknownURL means no plugin recognises a link.
var ErrUnknownURL = errors.New("no source recognises that URL")

// GameRef identifies a stored game in a plugin call.
type GameRef struct {
	ID        int64          `json:"id"`
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

func (l *Library) fetchGame(ctx context.Context, plugin string, ref ParsedRef) (contract.GameDTO, error) {
	var dto contract.GameDTO
	if err := l.sources.Call(ctx, plugin, methodFetchGame, ref, &dto); err != nil {
		return contract.GameDTO{}, fmt.Errorf("fetch game: %w", err)
	}
	return dto, nil
}

// refreshFiles stores the download candidates a plugin reports for a game, when it can.
func (l *Library) refreshFiles(ctx context.Context, game Game) error {
	info, ok := l.pluginFor(game.Source)
	if !ok || !hasCapability(info, CapabilityResolveFiles) {
		return nil
	}
	var candidates []contract.FileCandidate
	if err := l.sources.Call(ctx, info.Plugin, methodResolveFiles, gameRefOf(game), &candidates); err != nil {
		return err
	}
	return l.ReplaceFiles(ctx, game.ID, candidates)
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
