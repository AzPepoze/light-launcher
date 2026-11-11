package library

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// blockingLauncher starts when Launch is called and returns when release is closed.
type blockingLauncher struct {
	started chan string
	release chan struct{}
}

func (b *blockingLauncher) Launch(_ context.Context, exePath, args, name string) error {
	b.started <- exePath + "|" + args + "|" + name
	<-b.release
	return nil
}

func installedGame(t *testing.T, lib *Library) (Game, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, launchFixtureName), []byte("#!/bin/sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	game := upsertGame(t, lib, "launch", "Launch Me")
	args := "--fullscreen"
	if _, err := lib.Patch(context.Background(), game.ID, Patch{InstallDir: &dir, ExeArgs: &args}); err != nil {
		t.Fatal(err)
	}
	game, _ = lib.Get(context.Background(), game.ID)
	return game, dir
}

// launchFixtureName is a file the executable detector accepts on the current OS.
var launchFixtureName = fixtureNameFor(runtime.GOOS)

func fixtureNameFor(goos string) string {
	if goos == "windows" {
		return "run.exe"
	}
	return "run.sh"
}

func TestLaunchStartsGameAndBlocksSecondLaunch(t *testing.T) {
	launcher := &blockingLauncher{started: make(chan string, 1), release: make(chan struct{})}
	lib := openTestLibrary(t, Options{Launcher: launcher})
	game, dir := installedGame(t, lib)
	ctx := context.Background()

	result, err := lib.Launch(ctx, game.ID)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	if !result.Started || result.ExePath != filepath.Join(dir, launchFixtureName) {
		t.Errorf("result = %+v", result)
	}
	if got := <-launcher.started; got != filepath.Join(dir, launchFixtureName)+"|--fullscreen|Launch Me" {
		t.Errorf("launcher got %q", got)
	}

	if _, err := lib.Launch(ctx, game.ID); !errors.Is(err, ErrAlreadyRunning) {
		t.Errorf("second launch error = %v, want ErrAlreadyRunning", err)
	}

	close(launcher.release)
	waitUntil(t, func() bool { return !lib.launches.isRunning(game.ID) })
	stored, _ := lib.Get(ctx, game.ID)
	if stored.LastPlayedAt == nil || derefString(stored.ExePath) == "" {
		t.Errorf("last played or exe not recorded: %+v", stored)
	}
}

func TestLaunchNeedsInstallFolder(t *testing.T) {
	lib := openTestLibrary(t, Options{Launcher: &blockingLauncher{}})
	game := upsertGame(t, lib, "not-installed", "Not Installed")
	if _, err := lib.Launch(context.Background(), game.ID); err == nil {
		t.Error("launching a game without an install folder must fail")
	}
}

func TestAddPlaytimeAddsSeconds(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	game := upsertGame(t, lib, "play", "Playtime")
	if err := lib.addPlaytime(context.Background(), game.ID, 90); err != nil {
		t.Fatal(err)
	}
	if err := lib.addPlaytime(context.Background(), game.ID, 30); err != nil {
		t.Fatal(err)
	}
	stored, _ := lib.Get(context.Background(), game.ID)
	if stored.PlaytimeS != 120 {
		t.Errorf("playtime = %d, want 120", stored.PlaytimeS)
	}
}

// waitUntil polls until the condition holds or the test times out.
func waitUntil(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}
