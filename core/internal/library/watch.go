package library

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"light-launcher/core/internal/logger"
)

// Watch folder polling. Polling works the same on Windows and Linux, so no inotify is needed.
// A file is imported once its size is the same on two polls in a row, which means the download finished.

var partialSuffixes = []string{".part", ".crdownload", ".tmp"}

// watchedFile is one candidate file in the watch folder.
type watchedFile struct {
	Name    string
	Size    int64
	ModTime time.Time
}

// watchState is owned by the watcher goroutine. It is not shared with request handlers.
type watchState struct {
	lastSeen        map[string]watchedFile
	unmatched       map[string]bool
	lastCount       int
	warnedDownloads bool
}

func newWatchState() *watchState {
	return &watchState{lastSeen: map[string]watchedFile{}, unmatched: map[string]bool{}, lastCount: -1}
}

// RunWatcher polls the watch folder every interval until ctx is done.
func (l *Library) RunWatcher(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := l.ScanWatchFolder(ctx); err != nil {
				logger.Warn(logCategory, "watch scan failed: %v", err)
			}
		}
	}
}

// ScanWatchFolder runs one poll. It imports files that have one clear match and returns how many it imported.
func (l *Library) ScanWatchFolder(ctx context.Context) (int, error) {
	folder := l.watchFolder()
	if folder == "" {
		return 0, nil
	}
	if sameDir(folder, l.opts.DownloadsDir) {
		l.warnDownloadsOnce()
		return 0, nil
	}
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return 0, fmt.Errorf("create watch folder: %w", err)
	}
	files, err := listWatchable(folder)
	if err != nil {
		return 0, err
	}
	ignored, err := l.ignoredNames(ctx)
	if err != nil {
		return 0, err
	}
	games, err := l.All(ctx)
	if err != nil {
		return 0, err
	}
	titles := gameTitles(games)

	next := map[string]watchedFile{}
	imported := 0
	for _, file := range files {
		prev, seenBefore := l.watch.lastSeen[file.Name]
		stable := seenBefore && file.Size > 0 && prev.Size == file.Size
		if !stable || ignored[file.Name] {
			next[file.Name] = file
			continue
		}
		match, ok := bestConfidentMatch(file.Name, titles)
		if !ok {
			next[file.Name] = file
			l.reportUnmatched(file.Name)
			continue
		}
		if err := l.importWatched(ctx, folder, file.Name, games[match.Index].ID); err != nil {
			logger.Warn(logCategory, "could not import %s: %v", file.Name, err)
			next[file.Name] = file
			continue
		}
		delete(l.watch.unmatched, file.Name)
		imported++
	}
	l.watch.lastSeen = next
	if imported > 0 {
		logger.Info(logCategory, "imported %d watched file(s)", imported)
	}
	return imported, l.broadcastIncomingCount(ctx)
}

// broadcastIncomingCount tells the UI when the number of files that need a person changes.
func (l *Library) broadcastIncomingCount(ctx context.Context) error {
	report, err := l.Incoming(ctx)
	if err != nil {
		return err
	}
	if report.Count != l.watch.lastCount {
		l.watch.lastCount = report.Count
		l.notify("library.incoming", map[string]any{"count": report.Count})
	}
	return nil
}

func (l *Library) reportUnmatched(name string) {
	if l.watch.unmatched[name] {
		return
	}
	l.watch.unmatched[name] = true
	logger.Info(logCategory, "%s needs a game; open Incoming to link it", name)
}

func (l *Library) warnDownloadsOnce() {
	if l.watch.warnedDownloads {
		return
	}
	l.watch.warnedDownloads = true
	logger.Warn(logCategory, "watch folder is the downloads folder; pick a different folder")
}

// importWatched moves a watched file into the downloads folder and hands it to the sink for a game.
func (l *Library) importWatched(ctx context.Context, folder, name string, gameID int64) error {
	if l.opts.Sink == nil || l.opts.DownloadsDir == "" {
		return errors.New("watch import needs a download folder and an archive sink")
	}
	dest := uniquePath(filepath.Join(l.opts.DownloadsDir, name))
	if err := moveFile(filepath.Join(folder, name), dest); err != nil {
		return err
	}
	if err := l.opts.Sink.Accept(ctx, gameID, dest); err != nil {
		return err
	}
	return l.restoreIncoming(ctx, name)
}

// listWatchable returns the regular, non-partial files in folder, newest first.
func listWatchable(folder string) ([]watchedFile, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, fmt.Errorf("read watch folder: %w", err)
	}
	var files []watchedFile
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !isCandidateName(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, watchedFile{Name: entry.Name(), Size: info.Size(), ModTime: info.ModTime()})
	}
	return files, nil
}

// isCandidateName skips hidden files and downloads that are still being written.
func isCandidateName(name string) bool {
	if strings.HasPrefix(name, ".") {
		return false
	}
	lower := strings.ToLower(name)
	for _, suffix := range partialSuffixes {
		if strings.HasSuffix(lower, suffix) {
			return false
		}
	}
	return true
}

func gameTitles(games []Game) []string {
	titles := make([]string, len(games))
	for i, game := range games {
		titles[i] = game.Title
	}
	return titles
}

// sameDir compares two folder paths after cleaning them. An empty path never matches.
func sameDir(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	absA, errA := filepath.Abs(a)
	absB, errB := filepath.Abs(b)
	if errA != nil || errB != nil {
		return false
	}
	return filepath.Clean(absA) == filepath.Clean(absB)
}

// uniquePath adds " (2)", " (3)" and so on before the extension until the name is free.
func uniquePath(path string) string {
	if !exists(path) {
		return path
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if !exists(candidate) {
			return candidate
		}
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// moveFile renames a file, or copies and removes it when the rename crosses a device boundary.
func moveFile(from, to string) error {
	if err := os.Rename(from, to); err == nil {
		return nil
	}
	if err := copyFile(from, to); err != nil {
		return err
	}
	return os.Remove(from)
}

func copyFile(from, to string) error {
	src, err := os.Open(from)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(to, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return err
	}
	return dst.Close()
}
