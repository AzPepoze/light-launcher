package library

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"light-launcher/core/internal/contract"
)

// updateSources is a fake plugin that reports a newer release for "f95" games.
func updateSources(reply func(params any) (any, error)) *fakeSources {
	return &fakeSources{
		infos: []contract.SourceInfo{{Plugin: "f95plugin", SourceID: "f95", Capabilities: []string{CapabilityUpdateCheck}}},
		replies: map[string]func(any) (any, error){
			methodCheckUpdate: reply,
		},
	}
}

func f95Game(t *testing.T, lib *Library, version string) Game {
	t.Helper()
	id, err := lib.Upsert(context.Background(), contract.GameDTO{Source: "f95", SourceID: "42", URL: "https://f95/42", Title: "Harbor", Version: version})
	if err != nil {
		t.Fatal(err)
	}
	game, err := lib.Get(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return game
}

func TestCheckGameRecordsNewerVersion(t *testing.T) {
	sources := updateSources(func(any) (any, error) {
		return map[string]string{"version": "0.8", "changelog": "more things"}, nil
	})
	lib := openTestLibrary(t, Options{Sources: sources})
	game := f95Game(t, lib, "0.7")

	result, err := lib.CheckGame(context.Background(), game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !result.HasUpdate || derefString(result.To) != "0.8" {
		t.Errorf("result = %+v, want update to 0.8", result)
	}
	stored, _ := lib.Get(context.Background(), game.ID)
	if derefString(stored.Version) != "0.8" || stored.LastCheckedAt == nil {
		t.Errorf("stored game = version %v checked %v", stored.Version, stored.LastCheckedAt)
	}
	checks, _ := lib.ListChecks(context.Background(), game.ID, 10)
	if len(checks) != 1 || derefString(checks[0].NewVersion) != "0.8" || derefString(checks[0].Changelog) != "more things" {
		t.Errorf("checks = %+v", checks)
	}
}

func TestCheckGameNotModifiedIsNotAnError(t *testing.T) {
	sources := updateSources(func(any) (any, error) { return nil, errors.New("not-modified") })
	lib := openTestLibrary(t, Options{Sources: sources})
	game := f95Game(t, lib, "1.0")

	result, err := lib.CheckGame(context.Background(), game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if result.HasUpdate || result.Error != "" {
		t.Errorf("result = %+v, want no update and no error", result)
	}
}

func TestCheckGameReportsPluginError(t *testing.T) {
	sources := updateSources(func(any) (any, error) { return nil, errors.New("login expired") })
	lib := openTestLibrary(t, Options{Sources: sources})
	game := f95Game(t, lib, "1.0")

	result, err := lib.CheckGame(context.Background(), game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Error != "login expired" {
		t.Errorf("error = %q, want login expired", result.Error)
	}
}

func TestCheckGameSkipsSourcesWithoutUpdateCapability(t *testing.T) {
	sources := &fakeSources{infos: []contract.SourceInfo{{Plugin: "p", SourceID: "f95"}}}
	lib := openTestLibrary(t, Options{Sources: sources})
	game := f95Game(t, lib, "1.0")

	result, err := lib.CheckGame(context.Background(), game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if result.HasUpdate || result.Error != "" {
		t.Errorf("result = %+v, want a quiet no-update", result)
	}
}

type recordingInstaller struct {
	calls atomic.Int32
}

func (r *recordingInstaller) InstallUpdate(context.Context, Game, string) error {
	r.calls.Add(1)
	return nil
}

func TestAutoUpdateCallsInstallerOnlyWhenFlagged(t *testing.T) {
	sources := updateSources(func(any) (any, error) { return map[string]string{"version": "2.0"}, nil })
	installer := &recordingInstaller{}
	lib := openTestLibrary(t, Options{
		Sources:    sources,
		Installer:  installer,
		AutoUpdate: func(Game) bool { return true },
	})
	f95Game(t, lib, "1.0")

	if _, err := lib.CheckAll(context.Background(), nil, false); err != nil {
		t.Fatal(err)
	}
	if got := installer.calls.Load(); got != 1 {
		t.Errorf("installer calls = %d, want 1", got)
	}
}

func TestSkipRecentLeavesRecentlyCheckedGames(t *testing.T) {
	calls := atomic.Int32{}
	sources := updateSources(func(any) (any, error) {
		calls.Add(1)
		return nil, nil
	})
	lib := openTestLibrary(t, Options{Sources: sources, UpdateRecheck: time.Hour})
	game := f95Game(t, lib, "1.0")
	stamp := nowStamp()
	if _, err := lib.Patch(context.Background(), game.ID, Patch{LastCheckedAt: &stamp}); err != nil {
		t.Fatal(err)
	}

	if _, err := lib.CheckAll(context.Background(), nil, true); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 0 {
		t.Errorf("source was called %d times, want 0 for a recently checked game", calls.Load())
	}
	if _, err := lib.CheckAll(context.Background(), nil, false); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Errorf("manual check should call the source once, got %d", calls.Load())
	}
}
