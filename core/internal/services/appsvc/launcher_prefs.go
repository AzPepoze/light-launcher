package appsvc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/store"
)

// launcherPreferencesVersion is the layout version of launcher-preferences.json.
const launcherPreferencesVersion = 1

// preferenceStore remembers the launcher the user chose for each game folder.
type preferenceStore interface {
	// Preferred maps each cleaned game folder to its chosen launcher path.
	Preferred() map[string]string
	Set(folder, launcherPath string) error
	Clear(folder string) error
}

// launcherPreferences is launcher-preferences.json. Every read-modify-write holds mu,
// because store.File only locks its own save.
type launcherPreferences struct {
	mu   sync.Mutex
	path string
	file *store.File
}

// launcherPreferenceFile is the document on disk.
type launcherPreferenceFile struct {
	Folders map[string]string `json:"folders"`
}

func newLauncherPreferences(path string) *launcherPreferences {
	return &launcherPreferences{path: path, file: store.NewFile(path, launcherPreferencesVersion)}
}

// launcherPreferencesPath is the file next to settings.json.
func launcherPreferencesPath() string {
	return filepath.Join(config.GetBaseDirectory(), "launcher-preferences.json")
}

// preferenceKey is the key of one game folder in the preferences file.
func preferenceKey(folder string) string {
	return filepath.Clean(strings.TrimSpace(folder))
}

// Preferred returns the saved launchers. A file that cannot be read counts as empty.
func (p *launcherPreferences) Preferred() map[string]string {
	p.mu.Lock()
	defer p.mu.Unlock()
	doc, err := p.load()
	if err != nil {
		logger.Warn("games", "launcher preferences ignored: %v", err)
	}
	return doc.Folders
}

// Set remembers launcherPath as the launcher of folder.
func (p *launcherPreferences) Set(folder, launcherPath string) error {
	key := preferenceKey(folder)
	return p.update(func(folders map[string]string) {
		folders[key] = launcherPath
	})
}

// Clear forgets the launcher of folder.
func (p *launcherPreferences) Clear(folder string) error {
	key := preferenceKey(folder)
	return p.update(func(folders map[string]string) {
		delete(folders, key)
	})
}

// update applies change and saves. A corrupt or other-version file is copied to a .bak file first.
func (p *launcherPreferences) update(change func(folders map[string]string)) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	doc, err := p.load()
	if err != nil {
		logger.Warn("games", "launcher preferences reset after a bad file: %v", err)
		if backupErr := p.backupUnreadable(); backupErr != nil {
			return backupErr
		}
	}
	change(doc.Folders)
	return p.file.Save(doc)
}

// load reads the file. On error it returns an empty document with a usable map.
func (p *launcherPreferences) load() (launcherPreferenceFile, error) {
	doc := emptyLauncherPreferences()
	if _, err := p.file.Load(&doc); err != nil {
		return emptyLauncherPreferences(), err
	}
	if doc.Folders == nil {
		doc.Folders = map[string]string{}
	}
	return doc, nil
}

func emptyLauncherPreferences() launcherPreferenceFile {
	return launcherPreferenceFile{Folders: map[string]string{}}
}

func (p *launcherPreferences) backupUnreadable() error {
	data, err := os.ReadFile(p.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s for backup: %w", p.path, err)
	}
	backupPath := p.path + ".bak"
	if err := os.WriteFile(backupPath, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", backupPath, err)
	}
	return nil
}
