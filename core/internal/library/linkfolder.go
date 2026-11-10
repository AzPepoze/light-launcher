package library

import (
	"context"
	"errors"
	"fmt"
	"os"
)

// LinkFolder points a game at an existing install folder, sets its size and finds its launcher.
// A folder that another game already uses is refused.
func (l *Library) LinkFolder(ctx context.Context, id int64, dir string) (Game, error) {
	if _, err := l.Get(ctx, id); err != nil {
		return Game{}, err
	}
	info, err := os.Stat(dir)
	if err != nil {
		return Game{}, fmt.Errorf("folder: %w", err)
	}
	if !info.IsDir() {
		return Game{}, errors.New("folder: not a folder")
	}
	owner, err := l.ByInstallDir(ctx, dir)
	if err == nil && owner.ID != id {
		return Game{}, fmt.Errorf("that folder already belongs to %q", owner.Title)
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return Game{}, err
	}

	size := dirSize(dir)
	status := "installed"
	patch := Patch{InstallDir: &dir, InstallSize: &size, Status: &status}
	if exe := DetectExecutable(dir); exe != "" {
		patch.ExePath = &exe
	}
	return l.Patch(ctx, id, patch)
}
