package plugins

import (
	"context"
	"sort"

	"light-launcher/core/internal/contract"
)

// PluginInfo is what the UI shows for one plugin.
type PluginInfo struct {
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	Version        string                `json:"version"`
	Previous       string                `json:"previous,omitempty"`
	Status         string                `json:"status"`
	Enabled        bool                  `json:"enabled"`
	Running        bool                  `json:"running"`
	Error          string                `json:"error,omitempty"`
	DisabledReason string                `json:"disabledReason,omitempty"`
	Source         string                `json:"source,omitempty"`
	Unsigned       bool                  `json:"unsigned"`
	Permissions    Permissions           `json:"permissions"`
	Sources        []contract.SourceInfo `json:"sources"`
	Author         string                `json:"author"`
	Description    string                `json:"description"`
	Homepage       string                `json:"homepage"`
	HasSettings    bool                  `json:"hasSettings"`
}

// InstallRequest asks for one package, from a local path or an https URL.
// AcceptedPermissions must list the permissions the plugin asks for, as the preview's consent list.
type InstallRequest struct {
	Source              string   `json:"source"`
	AllowUnsigned       bool     `json:"allowUnsigned"`
	AcceptedPermissions []string `json:"acceptedPermissions"`
}

// List returns every installed plugin, sorted by id.
func (h *Host) List() []PluginInfo {
	ids, err := h.layout.installedIDs()
	if err != nil {
		return nil
	}
	sort.Strings(ids)
	infos := make([]PluginInfo, 0, len(ids))
	for _, id := range ids {
		infos = append(infos, h.info(id))
	}
	return infos
}

// PermissionsPreview verifies a package and returns it for the consent screen. Nothing is installed.
func (h *Host) PermissionsPreview(ctx context.Context, source string) (Preview, error) {
	return h.install.Preview(ctx, source)
}

// Install adds a plugin from a package. Unsigned packages need AllowUnsigned.
func (h *Host) Install(ctx context.Context, req InstallRequest) (PluginInfo, error) {
	id, err := h.install.Install(ctx, req.Source, req.AllowUnsigned, req.AcceptedPermissions)
	if err != nil {
		return PluginInfo{}, err
	}
	h.emitState(id)
	return h.info(id), nil
}

// Update installs the newest version from the source the plugin came from. A version that asks for new
// permissions needs accepted to list all of them.
func (h *Host) Update(ctx context.Context, id string, allowUnsigned bool, accepted []string) (PluginInfo, error) {
	if err := requireID(id); err != nil {
		return PluginInfo{}, err
	}
	h.stopPlugin(id)
	if err := h.install.Update(ctx, id, allowUnsigned, accepted); err != nil {
		return PluginInfo{}, err
	}
	h.emitState(id)
	return h.info(id), nil
}

// Rollback makes the previous version current again.
func (h *Host) Rollback(id string) (PluginInfo, error) {
	if err := requireID(id); err != nil {
		return PluginInfo{}, err
	}
	h.stopPlugin(id)
	if err := h.install.Rollback(id); err != nil {
		return PluginInfo{}, err
	}
	h.emitState(id)
	return h.info(id), nil
}

// Remove uninstalls a plugin. Its data folder is kept unless deleteData is set.
func (h *Host) Remove(id string, deleteData bool) error {
	if err := requireID(id); err != nil {
		return err
	}
	h.stopPlugin(id)
	if err := h.install.Remove(id, deleteData); err != nil {
		return err
	}
	h.emitRemoved(id)
	return nil
}

// SetEnabled turns a plugin on or off. Turning it on also clears an auto-disable.
func (h *Host) SetEnabled(id string, enabled bool) (PluginInfo, error) {
	if err := requireID(id); err != nil {
		return PluginInfo{}, err
	}
	if _, err := h.loadInstalled(id); err != nil {
		return PluginInfo{}, err
	}
	err := h.install.state.update(id, func(state *PluginState) {
		state.Enabled = enabled
		state.DisabledReason = ""
	})
	if err != nil {
		return PluginInfo{}, err
	}
	h.stopPlugin(id)
	h.emitState(id)
	return h.info(id), nil
}

// Logs returns the last lines a plugin wrote. A count of 0 returns all kept lines.
func (h *Host) Logs(id string, lines int) ([]LogLine, error) {
	if err := requireID(id); err != nil {
		return nil, err
	}
	return h.linesFor(id).tail(lines), nil
}

// info builds the UI view of one plugin. A plugin whose files cannot be read has status "broken".
func (h *Host) info(id string) PluginInfo {
	state := h.install.state.get(id)
	info := PluginInfo{
		ID:             id,
		Enabled:        state.Enabled,
		Source:         state.Source,
		Unsigned:       state.Unsigned,
		DisabledReason: state.DisabledReason,
		Status:         "stopped",
	}
	inst, err := h.loadInstalled(id)
	if err != nil {
		info.Status = "broken"
		info.Error = err.Error()
		return info
	}
	info.Name = inst.manifest.Name
	info.Version = inst.version
	info.Previous = inst.previous
	info.Permissions = inst.manifest.Permissions
	info.Author = inst.manifest.Author
	info.Description = inst.manifest.Description
	info.Homepage = inst.manifest.Homepage
	info.HasSettings = inst.manifest.Contributes.Settings != nil
	info.Sources = sourceInfos(id, inst.manifest)
	info.Running = h.running(id)
	switch {
	case !state.Enabled:
		info.Status = "disabled"
	case info.Running:
		info.Status = "running"
	}
	return info
}

func (h *Host) running(id string) bool {
	h.mu.Lock()
	sup := h.supervisors[id]
	h.mu.Unlock()
	return sup != nil && sup.Running()
}

func (h *Host) emitRemoved(id string) {
	if h.deps.events != nil {
		h.deps.events.Emit("plugin-state", map[string]any{"id": id, "removed": true})
	}
}

func sourceInfos(id string, manifest Manifest) []contract.SourceInfo {
	var sources []contract.SourceInfo
	for _, source := range manifest.Contributes.Sources {
		sources = append(sources, contract.SourceInfo{
			Plugin:       id,
			SourceID:     source.ID,
			Label:        source.Label,
			Capabilities: source.Capabilities,
		})
	}
	return sources
}
