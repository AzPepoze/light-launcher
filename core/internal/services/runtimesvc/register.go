// Package runtimesvc runs games through light-launcher-instance and manages Proton, prefixes, LSFG and sessions.
package runtimesvc

import (
	"context"
	"encoding/json"
	"fmt"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/platform"
)

// Register adds the runtime methods to the router. Progress events are sent through ev.
// Prefix, Proton and LSFG methods return an error (or an empty list) where the platform lacks them.
func Register(r *api.Router, ev api.Events) {
	features := platform.Current()
	svc := newDefaultService()
	r.Handle("RunGame", typed(func(in runGameInput) (any, error) {
		return nil, svc.runGame(in.Options, in.ShowLogs)
	}))
	r.Handle("GetRunningSessions", plain(func() (any, error) {
		return svc.getRunningSessions()
	}))
	r.Handle("KillSession", typed(func(in pidInput) (any, error) {
		return nil, killSession(in.Pid)
	}))

	registerPrefix(r, features.Prefix, svc)
	registerProton(r, features.Proton, ev)
	registerLsfg(r, features.Lsfg, ev, svc)
}

// typed decodes the payload into In before calling fn.
func typed[In any](fn func(In) (any, error)) api.Handler {
	return api.Typed(func(_ context.Context, in In) (any, error) {
		return fn(in)
	})
}

// plain wraps a method that takes no payload.
func plain(fn func() (any, error)) api.Handler {
	return api.NoInput(func(context.Context) (any, error) {
		return fn()
	})
}

// gated returns a handler that fails with "<feature> not supported on <os>" when the feature is off.
func gated(enabled bool, feature string, handler api.Handler) api.Handler {
	if enabled {
		return handler
	}
	return func(context.Context, json.RawMessage) (any, error) {
		return nil, unsupportedError(feature)
	}
}

// listOrEmpty returns an empty list when the feature is off, so the UI can show nothing.
func listOrEmpty(enabled bool, handler api.Handler) api.Handler {
	if enabled {
		return handler
	}
	return func(context.Context, json.RawMessage) (any, error) {
		return []any{}, nil
	}
}

func unsupportedError(feature string) error {
	return fmt.Errorf("%s not supported on %s", feature, platform.OS())
}

// progressEvent is the payload of the install progress events.
type progressEvent struct {
	Percent int    `json:"percent"`
	Message string `json:"message"`
}

func progressEmitter(ev api.Events, channel string) func(percent int, message string) {
	return func(percent int, message string) {
		ev.Emit(channel, progressEvent{Percent: percent, Message: message})
	}
}
