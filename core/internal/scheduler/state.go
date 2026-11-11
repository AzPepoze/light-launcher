package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const (
	stateKey     = "scheduler.lastRuns"
	stateVersion = 1
)

// MetaStore keeps the scheduler state next to the library. *library.Store implements it.
type MetaStore interface {
	GetMeta(ctx context.Context, key string) (string, bool, error)
	SetMeta(ctx context.Context, key, value string) error
}

type persistedState struct {
	Version  int               `json:"version"`
	LastRuns map[string]string `json:"lastRuns"`
}

func loadLastRuns(ctx context.Context, store MetaStore) (map[string]time.Time, error) {
	if store == nil {
		return map[string]time.Time{}, nil
	}
	raw, found, err := store.GetMeta(ctx, stateKey)
	if err != nil {
		return nil, fmt.Errorf("read scheduler state: %w", err)
	}
	if !found {
		return map[string]time.Time{}, nil
	}
	return decodeState(raw)
}

func decodeState(raw string) (map[string]time.Time, error) {
	var state persistedState
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		return nil, fmt.Errorf("scheduler state: %w", err)
	}
	if state.Version > stateVersion {
		return nil, fmt.Errorf("scheduler state version %d is newer than %d", state.Version, stateVersion)
	}
	// Version 0 is a value written before the version field existed.
	runs := make(map[string]time.Time, len(state.LastRuns))
	for name, text := range state.LastRuns {
		at, err := time.Parse(time.RFC3339Nano, text)
		if err != nil {
			continue
		}
		runs[name] = at
	}
	return runs, nil
}

func encodeState(times map[string]time.Time) (string, error) {
	runs := make(map[string]string, len(times))
	for name, at := range times {
		if at.IsZero() {
			continue
		}
		runs[name] = at.UTC().Format(time.RFC3339Nano)
	}
	raw, err := json.Marshal(persistedState{Version: stateVersion, LastRuns: runs})
	if err != nil {
		return "", fmt.Errorf("encode scheduler state: %w", err)
	}
	return string(raw), nil
}
