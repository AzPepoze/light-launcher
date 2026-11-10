package plugins

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

var (
	ErrUnsigned         = errors.New("package is not signed by a trusted key, pass allowUnsigned to install it anyway")
	ErrAlreadyInstalled = errors.New("plugin is already installed, use update")
	ErrNotInstalled     = errors.New("plugin is not installed")
	ErrNoUpdate         = errors.New("plugin is already at this version")
	ErrNoPrevious       = errors.New("no previous version to roll back to")
	ErrNoSource         = errors.New("plugin has no recorded install source, install it again")
)

// installer installs, updates, rolls back and removes packages under a Layout.
type installer struct {
	layout      Layout
	state       *stateStore
	client      *http.Client
	hostVersion string
}

func newInstaller(layout Layout, hostVersion string) (*installer, error) {
	state, err := loadState(layout.StatePath())
	if err != nil {
		return nil, err
	}
	return &installer{layout: layout, state: state, client: newDownloadClient(), hostVersion: hostVersion}, nil
}

// Preview unpacks and verifies a package and returns its manifest. Nothing is installed.
func (in *installer) Preview(ctx context.Context, source string) (Preview, error) {
	prepared, err := in.preparePackage(ctx, source)
	if err != nil {
		return Preview{}, err
	}
	defer prepared.discard()
	return Preview{Manifest: prepared.Manifest, Unsigned: prepared.Unsigned}, nil
}

// Install adds a new plugin. It fails when the plugin id is already installed.
func (in *installer) Install(ctx context.Context, source string, allowUnsigned bool) (string, error) {
	prepared, err := in.preparePackage(ctx, source)
	if err != nil {
		return "", err
	}
	defer prepared.discard()
	if prepared.Unsigned && !allowUnsigned {
		return "", ErrUnsigned
	}
	id := prepared.Manifest.ID
	if exists(in.layout.PluginDir(id)) {
		return "", ErrAlreadyInstalled
	}
	return id, in.commit(prepared, source)
}

// Update installs the version found at the source the plugin was installed from.
func (in *installer) Update(ctx context.Context, id string, allowUnsigned bool) error {
	source := in.state.get(id).Source
	if source == "" {
		return ErrNoSource
	}
	prepared, err := in.preparePackage(ctx, source)
	if err != nil {
		return err
	}
	defer prepared.discard()
	if prepared.Manifest.ID != id {
		return fmt.Errorf("package is for %q, not %q", prepared.Manifest.ID, id)
	}
	current, err := readPointer(in.layout.PluginDir(id), currentPointer)
	if err != nil {
		return err
	}
	if current == prepared.Manifest.Version {
		return ErrNoUpdate
	}
	if prepared.Unsigned && !allowUnsigned {
		return ErrUnsigned
	}
	return in.commit(prepared, source)
}

// Rollback makes the previous version current again. The version it replaces becomes the previous one.
func (in *installer) Rollback(id string) error {
	pluginDir := in.layout.PluginDir(id)
	previous, err := readPointer(pluginDir, previousPointer)
	if err != nil {
		return err
	}
	if previous == "" || !exists(in.layout.VersionDir(id, previous)) {
		return ErrNoPrevious
	}
	current, err := readPointer(pluginDir, currentPointer)
	if err != nil {
		return err
	}
	if err := writePointer(pluginDir, currentPointer, previous); err != nil {
		return err
	}
	if current == "" {
		return nil
	}
	return writePointer(pluginDir, previousPointer, current)
}

// Remove deletes the plugin files and its state. Plugin data is deleted only when deleteData is set.
func (in *installer) Remove(id string, deleteData bool) error {
	if !exists(in.layout.PluginDir(id)) {
		return ErrNotInstalled
	}
	if err := os.RemoveAll(in.layout.PluginDir(id)); err != nil {
		return err
	}
	if err := in.state.remove(id); err != nil {
		return err
	}
	if deleteData {
		return os.RemoveAll(in.layout.DataDir(id))
	}
	return nil
}

// commit moves the unpacked package into place and points "current" at it.
func (in *installer) commit(prepared preparedPackage, source string) error {
	manifest := prepared.Manifest
	pluginDir := in.layout.PluginDir(manifest.ID)
	versionDir := in.layout.VersionDir(manifest.ID, manifest.Version)
	if exists(versionDir) {
		return fmt.Errorf("%s %s is already installed", manifest.ID, manifest.Version)
	}
	if err := markEntriesExecutable(prepared.Dir, manifest); err != nil {
		return err
	}
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		return err
	}
	if err := os.Rename(prepared.Dir, versionDir); err != nil {
		return err
	}
	current, err := readPointer(pluginDir, currentPointer)
	if err != nil {
		return err
	}
	if current != "" {
		if err := writePointer(pluginDir, previousPointer, current); err != nil {
			return err
		}
	}
	if err := writePointer(pluginDir, currentPointer, manifest.Version); err != nil {
		return err
	}
	return in.state.update(manifest.ID, func(state *PluginState) {
		state.Source = source
		state.Unsigned = prepared.Unsigned
	})
}

// markEntriesExecutable sets the execute bit on the plugin programs. It does nothing on Windows.
func markEntriesExecutable(dir string, manifest Manifest) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	for _, entry := range manifest.Entry {
		if err := os.Chmod(filepath.Join(dir, filepath.FromSlash(entry)), 0o755); err != nil {
			return err
		}
	}
	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
