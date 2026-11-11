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
	schema, err := loadSettingsSchema(inst)
	if err != nil {
		return nil, err
	}
	dataDir := h.layout.DataDir(id)
	env := newPluginEnv(inst.manifest, dataDir, h.deps, h.linesFor(id))
	env.settingsSchema = schema
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

// CallPlugin runs a source method for the UI. Only "source.*" methods are allowed.
func (h *Host) CallPlugin(ctx context.Context, id, method string, params json.RawMessage) (json.RawMessage, error) {
	if !strings.HasPrefix(method, "source.") {
		return nil, fmt.Errorf("method %q is not allowed from the UI", method)
	}
	return h.callMethod(ctx, id, method, params)
}

// CallCommand runs a command from the plugin's manifest. The command must name a method.
func (h *Host) CallCommand(ctx context.Context, id, commandID string, params json.RawMessage) (json.RawMessage, error) {
	inst, err := h.loadInstalled(id)
	if err != nil {
		return nil, err
	}
	method, err := commandMethod(inst.manifest, commandID)
	if err != nil {
		return nil, err
	}
	return h.callMethod(ctx, id, method, params)
}

// commandMethod returns the plugin method a command runs. A command without a method cannot run.
func commandMethod(manifest Manifest, commandID string) (string, error) {
	for _, command := range manifest.Contributes.Commands {
		if command.ID != commandID {
			continue
		}
		if command.Method == "" {
			return "", fmt.Errorf("command %q has no method in plugin.json", commandID)
		}
		return command.Method, nil
	}
	return "", fmt.Errorf("plugin %s has no command %q", manifest.ID, commandID)
}

// callMethod starts the plugin when needed and runs one method on it.
func (h *Host) callMethod(ctx context.Context, id, method string, params json.RawMessage) (json.RawMessage, error) {
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
