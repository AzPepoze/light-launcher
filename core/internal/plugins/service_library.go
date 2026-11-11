package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/downloads"
)

type libraryParams struct {
	Games []contract.GameDTO `json:"games"`
}

// serviceLibraryUpsert saves games. Each game must come from a source this plugin contributes.
// It returns the library id of each game, in the same order, so a plugin can pass gameId to downloads.enqueue.
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
	ids, err := env.deps.library.UpsertMany(ctx, req.Games)
	if err != nil {
		return nil, internalError(err.Error())
	}
	return map[string]any{"upserted": len(req.Games), "ids": ids}, nil
}

func checkGameSources(games []contract.GameDTO, sources map[string]bool) error {
	for _, game := range games {
		if !sources[game.Source] {
			return invalidParams(fmt.Sprintf("game source %q is not a source of this plugin", game.Source))
		}
	}
	return nil
}

// isPlainFileName rejects anything that is not one safe path segment on every OS.
func isPlainFileName(name string) bool {
	return filepath.Base(name) == name && downloads.ValidateName(name) == nil
}
