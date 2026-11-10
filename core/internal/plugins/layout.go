package plugins

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"light-launcher/core/internal/config"
)

const (
	currentPointer  = "current"
	previousPointer = "previous"
	stateFile       = "state.json"
)

// Layout holds the folders the host keeps for plugins.
//
//	<config>/plugins/<id>/<version>/   installed package
//	<config>/plugins/<id>/current      version in use
//	<config>/plugins/<id>/previous     version to roll back to
//	<config>/plugins/state.json        enabled flags and install sources
//	<config>/plugins/trusted_keys      ed25519 public keys
//	<config>/plugin-data/<id>/         data kept by the plugin
type Layout struct {
	Root string
	Data string
}

// DefaultLayout uses the per-OS config folder of the launcher.
func DefaultLayout() Layout {
	base := config.GetBaseDirectory()
	return Layout{
		Root: filepath.Join(base, "plugins"),
		Data: filepath.Join(base, "plugin-data"),
	}
}

func (l Layout) PluginDir(id string) string {
	return filepath.Join(l.Root, id)
}

func (l Layout) VersionDir(id, version string) string {
	return filepath.Join(l.Root, id, version)
}

func (l Layout) DataDir(id string) string {
	return filepath.Join(l.Data, id)
}

func (l Layout) StatePath() string {
	return filepath.Join(l.Root, stateFile)
}

func (l Layout) TrustedKeysPath() string {
	return filepath.Join(l.Root, trustedKeysFile)
}

// readPointer returns the version a pointer file names, or "" when it does not exist.
func readPointer(pluginDir, name string) (string, error) {
	data, err := os.ReadFile(filepath.Join(pluginDir, name))
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// writePointer replaces a pointer file. It writes a temp file first, so a crash never leaves it half written.
func writePointer(pluginDir, name, version string) error {
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(pluginDir, name+".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(version + "\n"); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), filepath.Join(pluginDir, name))
}

// installedIDs lists the plugin folders under Root. Files and dot folders are skipped.
func (l Layout) installedIDs() ([]string, error) {
	entries, err := os.ReadDir(l.Root)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			ids = append(ids, entry.Name())
		}
	}
	return ids, nil
}
