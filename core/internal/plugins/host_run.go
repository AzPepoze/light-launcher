package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"light-launcher/core/internal/logger"
)

// installedPlugin is the current version of one installed plugin.
type installedPlugin struct {
	manifest Manifest
	version  string
	previous string
	dir      string
}

// loadInstalled reads the manifest of the current version.
func (h *Host) loadInstalled(id string) (installedPlugin, error) {
	if err := requireID(id); err != nil {
		return installedPlugin{}, err
	}
	pluginDir := h.layout.PluginDir(id)
	version, err := readPointer(pluginDir, currentPointer)
	if err != nil {
		return installedPlugin{}, err
	}
	if version == "" {
		return installedPlugin{}, ErrNotInstalled
	}
	dir := h.layout.VersionDir(id, version)
	data, err := os.ReadFile(filepath.Join(dir, "plugin.json"))
	if err != nil {
		return installedPlugin{}, fmt.Errorf("plugin %s %s: %w", id, version, err)
	}
	manifest, err := ParseManifest(data)
	if err != nil {
		return installedPlugin{}, err
	}
	if manifest.ID != id {
		return installedPlugin{}, fmt.Errorf("plugin folder %q holds plugin %q", id, manifest.ID)
	}
	previous, err := readPointer(pluginDir, previousPointer)
	if err != nil {
		return installedPlugin{}, err
	}
	return installedPlugin{manifest: manifest, version: version, previous: previous, dir: dir}, nil
}

// supervisorFor returns the running supervisor of an enabled plugin, creating it when needed.
func (h *Host) supervisorFor(id string) (*Supervisor, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if sup, ok := h.supervisors[id]; ok {
		return sup, nil
	}
	if !h.install.state.get(id).Enabled {
		return nil, ErrPluginDisabled
	}
	inst, err := h.loadInstalled(id)
	if err != nil {
		return nil, err
	}
	dataDir := h.layout.DataDir(id)
	env := newPluginEnv(inst.manifest, dataDir, h.deps, h.linesFor(id))
	cfg := h.tuning
	cfg.manifest = inst.manifest
	cfg.dir = inst.dir
	cfg.dataDir = dataDir
	cfg.hostVersion = h.hostVersion
	cfg.env = env
	cfg.onDisable = func(reason string) { h.disable(id, reason) }
	sup := newSupervisor(cfg)
	h.supervisors[id] = sup
	return sup, nil
}

// stopPlugin stops the running plugin and forgets its supervisor. The next call starts a fresh one.
func (h *Host) stopPlugin(id string) {
	h.mu.Lock()
	sup := h.supervisors[id]
	delete(h.supervisors, id)
	h.mu.Unlock()
	if sup != nil {
		sup.Shutdown()
	}
}

// disable turns a plugin off after it crashed too often.
func (h *Host) disable(id, reason string) {
	err := h.install.state.update(id, func(state *PluginState) {
		state.Enabled = false
		state.DisabledReason = reason
	})
	if err != nil {
		logger.Error("plugin:"+id, "could not save disabled state: %v", err)
	}
	h.stopPlugin(id)
	h.emitState(id)
}

// CallPlugin runs a method for the UI. Only source methods and declared commands are allowed.
func (h *Host) CallPlugin(ctx context.Context, id, method string, params json.RawMessage) (json.RawMessage, error) {
	inst, err := h.loadInstalled(id)
	if err != nil {
		return nil, err
	}
	if !allowedUIMethod(inst.manifest, method) {
		return nil, fmt.Errorf("method %q is not allowed from the UI", method)
	}
	sup, err := h.supervisorFor(id)
	if err != nil {
		return nil, err
	}
	var out json.RawMessage
	if err := sup.Call(ctx, method, params, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// allowedUIMethod allows "source.*" methods and "command.<id>" for commands the manifest declares.
func allowedUIMethod(manifest Manifest, method string) bool {
	if strings.HasPrefix(method, "source.") {
		return true
	}
	commandID, isCommand := strings.CutPrefix(method, "command.")
	if !isCommand {
		return false
	}
	for _, command := range manifest.Contributes.Commands {
		if command.ID == commandID {
			return true
		}
	}
	return false
}
