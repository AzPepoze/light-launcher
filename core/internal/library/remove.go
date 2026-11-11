package library

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"light-launcher/core/internal/logger"
)

// RemoveGame removes a game from the library and keeps a history entry.
// With deleteFiles, its install folder is deleted too, but only when that folder lies inside the games folder.
// It returns whether the folder was deleted.
func (l *Library) RemoveGame(ctx context.Context, id int64, deleteFiles bool) (bool, error) {
	game, err := l.Get(ctx, id)
	if err != nil {
		return false, err
	}
	deleted := false
	if deleteFiles && game.InstallDir != nil {
		deleted, err = l.deleteInstallDir(game)
		if err != nil {
			return false, err
		}
	}
	input := deletedInputOf(game)
	input.DeletedFiles = deleted
	if _, err := l.AddDeleted(ctx, input); err != nil {
		return false, err
	}
	if err := l.Remove(ctx, id); err != nil {
		return false, err
	}
	return deleted, nil
}

// deleteInstallDir deletes the game's install folder when it is strictly inside the games folder.
// Any other folder is kept and logged.
func (l *Library) deleteInstallDir(game Game) (bool, error) {
	dir := *game.InstallDir
	gamesDir := ""
	if l.opts.GamesDir != nil {
		gamesDir = l.opts.GamesDir()
	}
	if !isInsideFolder(gamesDir, dir) {
		logger.Warn(logCategory, "kept the files of %s: %s is not inside the games folder", game.Title, dir)
		return false, nil
	}
	if err := os.RemoveAll(dir); err != nil {
		return false, err
	}
	logger.Info(logCategory, "deleted the files of %s from %s", game.Title, dir)
	return true, nil
}

// isInsideFolder reports whether target is a folder strictly inside root. Symbolic links are resolved first.
func isInsideFolder(root, target string) bool {
	if root == "" || target == "" {
		return false
	}
	rootPath := resolvePath(root)
	targetPath := resolvePath(target)
	rel, err := filepath.Rel(rootPath, targetPath)
	if err != nil || rel == "." || rel == ".." || filepath.IsAbs(rel) {
		return false
	}
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// resolvePath returns the absolute, cleaned path, with symbolic links resolved when the path exists.
func resolvePath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved
	}
	return abs
}

// deletedInputOf describes a game for the removal history.
func deletedInputOf(game Game) DeletedInput {
	return DeletedInput{
		GameID:     game.ID,
		Source:     game.Source,
		SourceID:   game.SourceID,
		URL:        game.URL,
		Title:      game.Title,
		Version:    derefString(game.Version),
		Developer:  derefString(game.Developer),
		Tags:       game.Tags,
		InstallDir: derefString(game.InstallDir),
	}
}
