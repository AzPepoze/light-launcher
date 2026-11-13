package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/types"
)

const (
	sessionPollInterval = 2 * time.Second
	sessionStartWait    = 30 * time.Second // how long to wait for the game to show up as running
)

// errUnexpectedSessions means the runtime answered GetRunningSessions with an unexpected type.
var errUnexpectedSessions = errors.New("unexpected running sessions reply")

// gameLauncher starts library games through the runtime's RunGame method and waits until they exit.
// The library adds the playtime when Launch returns.
type gameLauncher struct {
	router *api.Router
}

func (g gameLauncher) Launch(ctx context.Context, exePath, args, name string) error {
	payload, err := json.Marshal(map[string]any{
		"options": types.LaunchOptions{Name: name, GamePath: exePath, CustomArgs: args},
	})
	if err != nil {
		return err
	}
	if _, err := g.router.Dispatch(ctx, "RunGame", payload); err != nil {
		return err
	}
	return g.waitForExit(ctx, exePath)
}

// LaunchConfig starts a local game with its saved prefix, Proton and argument settings, then waits
// until it exits so the library can add the playtime.
func (g gameLauncher) LaunchConfig(ctx context.Context, options types.LaunchOptions) error {
	payload, err := json.Marshal(map[string]any{"options": options})
	if err != nil {
		return err
	}
	if _, err := g.router.Dispatch(ctx, "RunGame", payload); err != nil {
		return err
	}
	path := options.GamePath
	if !options.UseGamePath && options.LauncherPath != "" {
		path = options.LauncherPath
	}
	return g.waitForExit(ctx, path)
}

// waitForExit returns once the game was seen running and is no longer running. RunGame returns as soon
// as the instance is spawned, so the first checks may come before the game shows up.
func (g gameLauncher) waitForExit(ctx context.Context, exePath string) error {
	deadline := time.Now().Add(sessionStartWait)
	seen := false
	for {
		running, err := g.isRunning(ctx, exePath)
		if err != nil {
			return err
		}
		if running {
			seen = true
		} else if seen || time.Now().After(deadline) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sessionPollInterval):
		}
	}
}

func (g gameLauncher) isRunning(ctx context.Context, exePath string) (bool, error) {
	reply, err := g.router.Dispatch(ctx, "GetRunningSessions", nil)
	if err != nil {
		return false, err
	}
	sessions, ok := reply.([]types.RunningSession)
	if !ok {
		return false, errUnexpectedSessions
	}
	for _, session := range sessions {
		if session.GamePath == exePath {
			return true, nil
		}
	}
	return false, nil
}
