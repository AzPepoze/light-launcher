package library

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/types"
)

// ErrAlreadyRunning means the game has a launch in progress.
var ErrAlreadyRunning = errors.New("game is already running")

// LaunchResult says which executable was started.
type LaunchResult struct {
	ExePath string `json:"exePath"`
	Started bool   `json:"started"`
}

// launchSet records which games are running, so one game never starts twice.
type launchSet struct {
	mu      sync.Mutex
	running map[int64]bool
}

func newLaunchSet() *launchSet {
	return &launchSet{running: map[int64]bool{}}
}

func (s *launchSet) begin(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running[id] {
		return false
	}
	s.running[id] = true
	return true
}

func (s *launchSet) end(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.running, id)
}

func (s *launchSet) isRunning(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running[id]
}

// Launch starts an installed game. It returns once the game is started; the
// launcher keeps running in the background, and playtime is added when it exits.
func (l *Library) Launch(ctx context.Context, id int64) (LaunchResult, error) {
	game, err := l.Get(ctx, id)
	if err != nil {
		return LaunchResult{}, err
	}
	if game.LocalConfig != nil {
		return l.launchLocal(ctx, game)
	}
	if l.opts.Launcher == nil {
		return LaunchResult{}, errors.New("no launcher is configured")
	}
	if game.InstallDir == nil {
		return LaunchResult{}, errors.New("game is not installed")
	}
	exe, err := l.resolveExecutable(ctx, game)
	if err != nil {
		return LaunchResult{}, err
	}
	if !l.launches.begin(id) {
		return LaunchResult{}, ErrAlreadyRunning
	}
	started := time.Now()
	playedAt := stamp(started)
	status := "installed"
	if _, err := l.Patch(ctx, id, Patch{LastPlayedAt: &playedAt, Status: &status}); err != nil {
		l.launches.end(id)
		return LaunchResult{}, err
	}
	go l.runGame(id, game, exe, started)
	l.notify("library.game", map[string]any{"id": id})
	return LaunchResult{ExePath: exe, Started: true}, nil
}

// launchLocal starts a game that has its own launch settings. The launcher keeps the prefix,
// Proton and argument handling; the library only records the session and its playtime.
func (l *Library) launchLocal(ctx context.Context, game Game) (LaunchResult, error) {
	if l.opts.ConfigLauncher == nil {
		return LaunchResult{}, errors.New("no local launcher is configured")
	}
	config := *game.LocalConfig
	exe := config.LauncherPath
	if !config.UseGamePath && exe == "" {
		exe = config.GamePath
	}
	if !l.launches.begin(game.ID) {
		return LaunchResult{}, ErrAlreadyRunning
	}
	started := time.Now()
	playedAt := stamp(started)
	status := "completed"
	if game.InstallDir != nil {
		status = "installed"
	}
	if _, err := l.Patch(ctx, game.ID, Patch{LastPlayedAt: &playedAt, Status: &status}); err != nil {
		l.launches.end(game.ID)
		return LaunchResult{}, err
	}
	go l.runLocal(game, config, started)
	l.notify("library.game", map[string]any{"id": game.ID})
	return LaunchResult{ExePath: exe, Started: true}, nil
}

// runLocal waits for a local game to exit and adds the playtime.
func (l *Library) runLocal(game Game, config types.LaunchOptions, started time.Time) {
	err := l.opts.ConfigLauncher.LaunchConfig(context.Background(), config)
	l.launches.end(game.ID)
	seconds := int64(math.Round(time.Since(started).Seconds()))
	if err := l.addPlaytime(context.Background(), game.ID, seconds); err != nil {
		logger.Warn(logCategory, "could not save playtime for %s: %v", game.Title, err)
	}
	if err != nil {
		logger.Warn(logCategory, "%s exited with error: %v", game.Title, err)
	} else {
		logger.Info(logCategory, "%s exited after %ds", game.Title, seconds)
	}
	l.notify("library.game", map[string]any{"id": game.ID})
}

// runGame waits for the launcher to return and adds the elapsed time to the game's playtime.
// It uses a background context so closing the request does not stop the game.
func (l *Library) runGame(id int64, game Game, exe string, started time.Time) {
	err := l.opts.Launcher.Launch(context.Background(), exe, derefString(game.ExeArgs), game.Title)
	l.launches.end(id)
	seconds := int64(math.Round(time.Since(started).Seconds()))
	if err := l.addPlaytime(context.Background(), id, seconds); err != nil {
		logger.Warn(logCategory, "could not save playtime for %s: %v", game.Title, err)
	}
	if err != nil {
		logger.Warn(logCategory, "%s exited with error: %v", game.Title, err)
	} else {
		logger.Info(logCategory, "%s exited after %ds", game.Title, seconds)
	}
	l.notify("library.game", map[string]any{"id": id})
}

func (l *Library) addPlaytime(ctx context.Context, id, seconds int64) error {
	current, err := l.Get(ctx, id)
	if err != nil {
		return err
	}
	total := current.PlaytimeS + seconds
	_, err = l.Patch(ctx, id, Patch{PlaytimeS: &total})
	return err
}

// resolveExecutable returns the stored launch file, or finds one in the install folder and stores it.
func (l *Library) resolveExecutable(ctx context.Context, game Game) (string, error) {
	if game.ExePath != nil {
		if _, err := os.Stat(*game.ExePath); err == nil {
			return *game.ExePath, nil
		}
	}
	found := DetectExecutable(*game.InstallDir)
	if found == "" {
		return "", errors.New("no executable found in the install folder; set the launch executable in the game details")
	}
	if _, err := l.Patch(ctx, game.ID, Patch{ExePath: &found}); err != nil {
		return "", fmt.Errorf("save executable: %w", err)
	}
	return found, nil
}
