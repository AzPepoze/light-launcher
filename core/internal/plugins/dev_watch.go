package plugins

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"light-launcher/core/internal/logger"
)

// devDebounce waits for a burst of saves to end before the plugin restarts.
const devDebounce = 300 * time.Millisecond

// ignoredDevDirs are folders the watcher skips. Version control and installed dependencies churn too much.
var ignoredDevDirs = map[string]bool{".git": true, "node_modules": true}

// devWatch reports changes under one linked plugin folder.
type devWatch struct {
	root    string
	watcher *fsnotify.Watcher
	done    chan struct{}
	once    sync.Once
}

// startDevWatch watches root and every folder below it that is not ignored. onChange runs once per burst of changes.
func startDevWatch(root string, onChange func()) (*devWatch, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &devWatch{root: root, watcher: watcher, done: make(chan struct{})}
	if err := w.addTree(root); err != nil {
		_ = watcher.Close()
		return nil, err
	}
	go w.run(onChange)
	return w, nil
}

// stop ends the watch. It does not wait for a restart that is already running.
func (w *devWatch) stop() {
	w.once.Do(func() {
		close(w.done)
		_ = w.watcher.Close()
	})
}

// addTree adds dir and every folder below it, except ignored ones. An unreadable folder below the root is skipped.
func (w *devWatch) addTree(dir string) error {
	return filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if path == dir {
				return err
			}
			return nil
		}
		if !entry.IsDir() {
			return nil
		}
		if path != dir && ignoredDevDirs[entry.Name()] {
			return filepath.SkipDir
		}
		return w.watcher.Add(path)
	})
}

func (w *devWatch) run(onChange func()) {
	var restart <-chan time.Time
	for {
		select {
		case <-w.done:
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if w.ignores(event.Name) || event.Op == fsnotify.Chmod {
				continue
			}
			w.watchNewFolder(event)
			restart = time.After(devDebounce)
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			logger.Warn("plugin-dev", "watching %s: %v", w.root, err)
		case <-restart:
			restart = nil
			onChange()
		}
	}
}

// watchNewFolder watches a folder created after the watch started, so its files are seen too.
func (w *devWatch) watchNewFolder(event fsnotify.Event) {
	if !event.Has(fsnotify.Create) {
		return
	}
	info, err := os.Stat(event.Name)
	if err != nil || !info.IsDir() {
		return
	}
	if err := w.addTree(event.Name); err != nil {
		logger.Warn("plugin-dev", "watching %s: %v", event.Name, err)
	}
}

// ignores reports whether path lies inside an ignored folder of the watched root.
func (w *devWatch) ignores(path string) bool {
	rel, err := filepath.Rel(w.root, path)
	if err != nil {
		return false
	}
	for _, part := range strings.Split(filepath.ToSlash(rel), "/") {
		if ignoredDevDirs[part] {
			return true
		}
	}
	return false
}

// ensureDevWatchLocked starts watching a linked plugin folder once. The caller holds h.mu.
func (h *Host) ensureDevWatchLocked(id, dir string) {
	if _, ok := h.devWatches[id]; ok {
		return
	}
	w, err := startDevWatch(dir, func() { h.restartDev(id) })
	if err != nil {
		logger.Warn("plugin:"+id, "cannot watch %s for changes: %v", dir, err)
		return
	}
	h.devWatches[id] = w
}

// stopDevWatch stops watching a plugin folder. The next start of a linked plugin watches again.
func (h *Host) stopDevWatch(id string) {
	h.mu.Lock()
	w := h.devWatches[id]
	delete(h.devWatches, id)
	h.mu.Unlock()
	if w != nil {
		w.stop()
	}
}

// takeDevWatchesLocked empties the watch map and returns the watches, which the caller stops. The caller holds h.mu.
func (h *Host) takeDevWatchesLocked() []*devWatch {
	watches := make([]*devWatch, 0, len(h.devWatches))
	for id, w := range h.devWatches {
		watches = append(watches, w)
		delete(h.devWatches, id)
	}
	return watches
}

// restartDev stops a linked plugin after its files changed, so the next call starts it from the folder again.
func (h *Host) restartDev(id string) {
	logger.Info("plugin:"+id, "files changed, restarting from the source folder")
	h.stopPlugin(id)
	h.emitState(id)
}
