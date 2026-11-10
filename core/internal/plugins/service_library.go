package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"light-launcher/core/internal/contract"
)

type downloadParams struct {
	URL      string            `json:"url"`
	Filename string            `json:"filename"`
	Headers  map[string]string `json:"headers"`
}

type libraryParams struct {
	Games []contract.GameDTO `json:"games"`
}

// serviceDownloadsEnqueue queues a file for the host downloader. The URL host must be in the network grant.
func serviceDownloadsEnqueue(ctx context.Context, env *pluginEnv, params json.RawMessage) (any, error) {
	req, err := decodeParams[downloadParams](params)
	if err != nil {
		return nil, err
	}
	target, err := parseHTTPTarget(req.URL)
	if err != nil {
		return nil, err
	}
	if !env.manifest.Permissions.allowsHost(target.Hostname()) {
		return nil, hostDenied(target.Hostname())
	}
	if !isPlainFileName(req.Filename) {
		return nil, invalidParams("filename must be a plain file name")
	}
	if env.deps.downloads == nil {
		return nil, internalError("downloads are not available")
	}
	id, err := env.deps.downloads.Enqueue(ctx, contract.DownloadRequest{
		URL:      req.URL,
		Headers:  req.Headers,
		Filename: req.Filename,
	})
	if err != nil {
		return nil, internalError(err.Error())
	}
	return map[string]int64{"id": id}, nil
}

// serviceLibraryUpsert saves games. Each game must come from a source this plugin contributes.
func serviceLibraryUpsert(ctx context.Context, env *pluginEnv, params json.RawMessage) (any, error) {
	req, err := decodeParams[libraryParams](params)
	if err != nil {
		return nil, err
	}
	if err := checkGameSources(req.Games, env.sourceIDs()); err != nil {
		return nil, err
	}
	if env.deps.library == nil {
		return nil, internalError("library is not available")
	}
	for _, game := range req.Games {
		if _, err := env.deps.library.Upsert(ctx, game); err != nil {
			return nil, internalError(err.Error())
		}
	}
	return map[string]int{"upserted": len(req.Games)}, nil
}

func checkGameSources(games []contract.GameDTO, sources map[string]bool) error {
	for _, game := range games {
		if !sources[game.Source] {
			return invalidParams(fmt.Sprintf("game source %q is not a source of this plugin", game.Source))
		}
	}
	return nil
}

// isPlainFileName rejects empty names, dot names and anything with a folder part.
func isPlainFileName(name string) bool {
	return name != "" && name != "." && name != ".." && filepath.Base(name) == name
}
