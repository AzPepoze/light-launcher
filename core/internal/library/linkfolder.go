package library

import (
	"context"
	"errors"
	"fmt"
	"os"
)

// CheckInstallDir returns an error unless dir is an existing folder that no other game uses.
func (l *Library) CheckInstallDir(ctx context.Context, id int64, dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("folder: %w", err)
	}
	if !info.IsDir() {
		return errors.New("folder: not a folder")
	}
	owner, err := l.ByInstallDir(ctx, dir)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if owner.ID != id {
		return fmt.Errorf("that folder already belongs to %q", owner.Title)
	}
	return nil
}

// LinkFolder points a game at an existing install folder, sets its size and finds its launcher.
// A folder that another game already uses is refused.
func (l *Library) LinkFolder(ctx context.Context, id int64, dir string) (Game, error) {
	if _, err := l.Get(ctx, id); err != nil {
		return Game{}, err
	}
	if err := l.CheckInstallDir(ctx, id, dir); err != nil {
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
