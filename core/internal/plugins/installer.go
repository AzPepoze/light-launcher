package plugins

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
// A non-empty expectedSHA must match the package file, as the release SHA256SUMS says.
func (in *installer) Preview(ctx context.Context, source, expectedSHA string) (Preview, error) {
	prepared, err := in.preparePackage(ctx, source, expectedSHA)
	if err != nil {
		return Preview{}, err
	}
	defer prepared.discard()
	return Preview{
		Manifest: prepared.Manifest,
		Unsigned: prepared.Unsigned,
		Consent:  PermissionSummary(prepared.Manifest.Permissions),
	}, nil
}

// Install adds a new plugin. It fails when the plugin id is already installed, or when accepted
// does not list the permissions the plugin asks for.
func (in *installer) Install(ctx context.Context, source string, allowUnsigned bool, accepted []string) (string, error) {
	prepared, err := in.preparePackage(ctx, source, "")
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
	summary := PermissionSummary(prepared.Manifest.Permissions)
	if err := checkConsent(summary, accepted); err != nil {
		return "", err
	}
	return id, in.commit(prepared, source, summary)
}

// Update installs the version found at the source the plugin was installed from.
// When the new version asks for permissions the user has not accepted, accepted must list all of them.
func (in *installer) Update(ctx context.Context, id string, allowUnsigned bool, accepted []string) error {
	prepared, source, err := in.prepareUpdate(ctx, id, "", "", allowUnsigned, accepted)
	if err != nil {
		return err
	}
	defer prepared.discard()
	return in.commitUpdate(prepared, source)
}

// prepareUpdate unpacks the newest version and runs every check before anything is changed.
// An empty source uses the recorded install source. The caller must discard the returned package.
func (in *installer) prepareUpdate(ctx context.Context, id, source, expectedSHA string, allowUnsigned bool, accepted []string) (preparedPackage, string, error) {
	if source == "" {
		source = in.state.get(id).Source
	}
	if source == "" {
		return preparedPackage{}, "", ErrNoSource
	}
	prepared, err := in.preparePackage(ctx, source, expectedSHA)
	if err != nil {
		return preparedPackage{}, "", err
	}
	if err := in.checkUpdate(id, prepared, allowUnsigned, accepted); err != nil {
		prepared.discard()
		return preparedPackage{}, "", err
	}
	return prepared, source, nil
}

// checkUpdate returns an error when the prepared package must not replace the installed version.
func (in *installer) checkUpdate(id string, prepared preparedPackage, allowUnsigned bool, accepted []string) error {
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
	summary := PermissionSummary(prepared.Manifest.Permissions)
	baseline, err := in.acceptedBaseline(id, current)
	if err != nil {
		return err
	}
	if added := newPermissions(baseline, summary); len(added) > 0 && checkConsent(summary, accepted) != nil {
		return fmt.Errorf("%w: %s", ErrNewPermissions, strings.Join(added, ", "))
	}
	return nil
}

// commitUpdate installs a package from prepareUpdate and makes it the current version.
func (in *installer) commitUpdate(prepared preparedPackage, source string) error {
	return in.commit(prepared, source, PermissionSummary(prepared.Manifest.Permissions))
}

// acceptedBaseline returns the permissions the user accepted for a plugin. Plugins installed before
// consent was recorded fall back to the permissions of their installed version.
func (in *installer) acceptedBaseline(id, current string) ([]string, error) {
	if recorded := in.state.get(id).AcceptedPermissions; recorded != nil {
		return recorded, nil
	}
	data, err := os.ReadFile(filepath.Join(in.layout.VersionDir(id, current), "plugin.json"))
	if err != nil {
		return nil, err
	}
	manifest, err := ParseManifest(data)
	if err != nil {
		return nil, err
	}
	return PermissionSummary(manifest.Permissions), nil
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

// commit moves the unpacked package into place, points "current" at it, and records the accepted permissions.
func (in *installer) commit(prepared preparedPackage, source string, consent []string) error {
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
		state.AcceptedPermissions = consent
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
