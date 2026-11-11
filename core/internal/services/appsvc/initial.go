package appsvc

import (
	"os"
	"sync"

	"light-launcher/core/internal/config"
)

// initialArgs holds the paths and flag the app was started with.
var initialArgs struct {
	mu       sync.Mutex
	launcher string
	game     string
	editLsfg bool
}

// SetInitialArgs stores the launch arguments. Environment variables are the fallback.
func SetInitialArgs(launcher, game string, editLsfg bool) {
	initialArgs.mu.Lock()
	defer initialArgs.mu.Unlock()
	initialArgs.launcher = launcher
	initialArgs.game = game
	initialArgs.editLsfg = editLsfg
}

func initialLauncherPath() string {
	initialArgs.mu.Lock()
	defer initialArgs.mu.Unlock()
	return config.FirstNonEmpty(initialArgs.launcher, os.Getenv("LIGHT_LAUNCHER_LAUNCHER_PATH"))
}

func initialGamePath() string {
	initialArgs.mu.Lock()
	defer initialArgs.mu.Unlock()
	return config.FirstNonEmpty(initialArgs.game, os.Getenv("LIGHT_LAUNCHER_GAME_PATH"))
}

func shouldEditLsfg() bool {
	initialArgs.mu.Lock()
	defer initialArgs.mu.Unlock()
	return initialArgs.editLsfg || os.Getenv("LIGHT_LAUNCHER_EDIT_LSFG") == "1"
}
