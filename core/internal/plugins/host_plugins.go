package plugins

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"time"

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
	Idle           bool                  `json:"idle"`
	LastActiveAt   string                `json:"lastActiveAt,omitempty"`
	Error          string                `json:"error,omitempty"`
	DisabledReason string                `json:"disabledReason,omitempty"`
	Source         string                `json:"source,omitempty"`
	Unsigned       bool                  `json:"unsigned"`
	Permissions    Permissions           `json:"permissions"`
	Sources        []contract.SourceInfo `json:"sources"`
	Author         string                `json:"author"`
	Description    string                `json:"description"`
	Homepage       string                `json:"homepage"`
	Repository     string                `json:"repository,omitempty"`
	HasSettings    bool                  `json:"hasSettings"`
	// Commands are the palette actions the plugin declares. They call the plugin when picked.
	Commands []CommandContribution `json:"commands"`
}

// InstallRequest asks for one package, from a local path or an https URL.
// AcceptedPermissions must list the permissions the plugin asks for, as the preview's consent list.
type InstallRequest struct {
	Source              string   `json:"source"`
	AllowUnsigned       bool     `json:"allowUnsigned"`
	AcceptedPermissions []string `json:"acceptedPermissions"`
}

// UpdateRequest installs a newer version. An empty Source uses the recorded install source.
// A GitHub update passes the checked asset URL and, when the release lists it, its SHA256.
type UpdateRequest struct {
	ID                  string   `json:"id"`
	Source              string   `json:"source,omitempty"`
	SHA256              string   `json:"sha256,omitempty"`
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
	return markSourceCollisions(infos)
}

// markSourceCollisions keeps the first enabled plugin that claims a source id and marks later ones
// broken. Without this, two plugins could shadow each other and the UI would silently pick one.
func markSourceCollisions(infos []PluginInfo) []PluginInfo {
	owners := map[string]string{}
	for i := range infos {
		info := &infos[i]
		if !info.Enabled || info.Status == "broken" {
			continue
		}
		for _, source := range info.Sources {
			if owner, taken := owners[source.SourceID]; taken && owner != info.ID {
				info.Status = "broken"
				info.Error = fmt.Sprintf("source id %q is already provided by plugin %s", source.SourceID, owner)
				break
			}
			owners[source.SourceID] = info.ID
		}
	}
	return infos
}

// PermissionsPreview verifies a package and returns it for the consent screen. Nothing is installed.
// A non-empty expectedSHA must match the package file, as the release SHA256SUMS says.
func (h *Host) PermissionsPreview(ctx context.Context, source, expectedSHA string) (Preview, error) {
	preview, err := h.install.Preview(ctx, source, expectedSHA)
	if err != nil {
		return Preview{}, err
	}
	h.flagRepositoryChange(&preview)
	return preview, nil
}

// flagRepositoryChange warns when an update would point future updates at another repository.
func (h *Host) flagRepositoryChange(preview *Preview) {
	installed, err := h.loadInstalled(preview.Manifest.ID)
	if err != nil {
		return
	}
	if installed.manifest.Repository == preview.Manifest.Repository {
		return
	}
	preview.RepositoryChanged = true
	preview.InstalledRepository = installed.manifest.Repository
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

// Update installs a newer version. A version that asks for new permissions needs the request to list
// all of them. A version that fails to start is rolled back to the previous one.
func (h *Host) Update(ctx context.Context, req UpdateRequest) (PluginInfo, error) {
	if err := requireID(req.ID); err != nil {
		return PluginInfo{}, err
	}
	source, sha256, err := h.updateSource(req.ID, req.Source, req.SHA256)
	if err != nil {
		return PluginInfo{}, err
	}
	// Checks run first, so a refused update leaves the running plugin alone.
	prepared, source, err := h.install.prepareUpdate(ctx, req.ID, source, sha256, req.AllowUnsigned, req.AcceptedPermissions)
	if err != nil {
		return PluginInfo{}, err
	}
	defer prepared.discard()
	wasRunning := h.running(req.ID)
	h.stopPlugin(req.ID)
	if err := h.install.commitUpdate(prepared, source); err != nil {
		return PluginInfo{}, err
	}
	if err := h.startUpdateTarget(ctx, req.ID, wasRunning); err != nil {
		return PluginInfo{}, h.rollbackFailedUpdate(ctx, req.ID, wasRunning, err)
	}
	h.forgetUpdate(req.ID)
	h.emitState(req.ID)
	return h.info(req.ID), nil
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
	h.forgetUpdate(id)
	h.emitState(id)
	return h.info(id), nil
}

// Remove uninstalls a plugin and deletes its secrets. Its data folder is kept unless deleteData is set.
func (h *Host) Remove(id string, deleteData bool) error {
	if err := requireID(id); err != nil {
		return err
	}
	h.stopPlugin(id)
	if err := h.deleteSecrets(id); err != nil {
		return err
	}
	if err := h.install.Remove(id, deleteData); err != nil {
		return err
	}
	h.forgetUpdate(id)
	h.emitRemoved(id)
	return nil
}

// deleteSecrets removes every secret the plugin stored. It runs before the state is dropped.
func (h *Host) deleteSecrets(id string) error {
	ctx := context.Background()
	keys := slices.Clone(h.install.state.get(id).SecretKeys)
	for _, key := range keys {
		if err := h.deps.secrets.Delete(ctx, id, key); err != nil {
			return fmt.Errorf("delete secret %q: %w", key, err)
		}
	}
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
	info.Repository = inst.manifest.Repository
	info.HasSettings = inst.manifest.Contributes.Settings != nil
	info.Commands = inst.manifest.Contributes.Commands
	info.Sources = sourceInfos(id, inst.manifest)
	sup := h.supervisor(id)
	info.Running = sup != nil && sup.Running()
	if sup != nil {
		info.Idle = sup.Idle()
		if lastActive := sup.LastActive(); !lastActive.IsZero() {
			info.LastActiveAt = lastActive.UTC().Format(time.RFC3339Nano)
		}
	}
	switch {
	case !state.Enabled:
		info.Status = "disabled"
	case info.Running:
		info.Status = "running"
	}
	return info
}

// supervisor returns the supervisor of a plugin, or nil when it never started or was reset.
func (h *Host) supervisor(id string) *Supervisor {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.supervisors[id]
}

// running reports whether the plugin process is up.
func (h *Host) running(id string) bool {
	sup := h.supervisor(id)
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
			Tile:         tileOrDefault(source.Tile),
			Cover:        contract.NormalizeCover(coverRatio(source.Cover), source.Tile),
			Actions:      actionsWithScope(source.Actions),
			Card:         source.Card,
			Sidebar:      source.Sidebar,
		})
	}
	return sources
}

// tileOrDefault returns the card shape of a source, poster when the manifest names none.
func tileOrDefault(tile string) string {
	if tile == "" {
		return TilePoster
	}
	return tile
}

// coverRatio returns the declared cover ratio, or "" when the manifest has no cover.
func coverRatio(cover *CoverContribution) string {
	if cover == nil {
		return ""
	}
	return cover.Ratio
}

// actionsWithScope returns a copy of the actions with the library scope filled in where none is set.
func actionsWithScope(actions []contract.SourceAction) []contract.SourceAction {
	out := slices.Clone(actions)
	for i := range out {
		if out[i].Scope == "" {
			out[i].Scope = contract.ActionScopeLibrary
		}
	}
	return out
}
