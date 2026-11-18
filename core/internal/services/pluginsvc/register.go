// Package pluginsvc exposes the plugin host to the UI as router methods.
package pluginsvc

import (
	"context"
	"encoding/json"
	"strings"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/config"
	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/plugins"
	"light-launcher/core/internal/types"
	"light-launcher/core/internal/uploads"
)

// updateInput updates a plugin. AcceptedPermissions is needed when the new version asks for more permissions.
// Source and SHA256 are set for a GitHub release update.
type updateInput struct {
	ID                  string   `json:"id"`
	Source              string   `json:"source"`
	SHA256              string   `json:"sha256"`
	AllowUnsigned       bool     `json:"allowUnsigned"`
	AcceptedPermissions []string `json:"acceptedPermissions"`
}

type idInput struct {
	ID string `json:"id"`
}

type removeInput struct {
	ID         string `json:"id"`
	DeleteData bool   `json:"deleteData"`
}

// devLinkInput names the source folder to run a plugin from.
type devLinkInput struct {
	Path string `json:"path"`
}

type setEnabledInput struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

type sourceInput struct {
	Source string `json:"source"`
	SHA256 string `json:"sha256"`
}

type settingsSetInput struct {
	ID       string         `json:"id"`
	Settings map[string]any `json:"settings"`
}

// callInput names either a source method or a command of the plugin, not both.
type callInput struct {
	Plugin  string          `json:"plugin"`
	Method  string          `json:"method"`
	Command string          `json:"command"`
	Params  json.RawMessage `json:"params"`
}

type logsInput struct {
	ID    string `json:"id"`
	Lines int    `json:"lines"`
}

// trustedKeyInput names one base64 ed25519 public key to add to or remove from trusted_keys.
type trustedKeyInput struct {
	Key string `json:"key"`
}

// removeUpload deletes the upload that source is in. A failure only logs, since the next start clears leftovers.
func removeUpload(store *uploads.Store, source string) {
	if err := store.Remove(source); err != nil {
		logger.Warn("plugins", "remove upload: %v", err)
	}
}

// pluginCall runs a command when one is named, and a source method otherwise.
func pluginCall(ctx context.Context, host *plugins.Host, in callInput) (any, error) {
	if in.Command != "" {
		return host.CallCommand(ctx, in.Plugin, in.Command, in.Params)
	}
	return host.CallPlugin(ctx, in.Plugin, in.Method, in.Params)
}

// registrySettings is the settings view the community registry method reads.
type registrySettings interface {
	Load() types.AppSettings
}

// registryFetch loads a community index from its https URL. Tests replace it with a fixture.
type registryFetch func(ctx context.Context, url string) (plugins.RegistryIndex, error)

// Register adds the plugin methods to r. Host events such as "plugin-state" go to ev.
// Installs and updates from an upload remove that upload whether or not the call succeeded.
func Register(r *api.Router, ev api.Events, host *plugins.Host, uploads *uploads.Store) {
	register(r, ev, host, uploads, config.NewSettingsStore(config.SettingsFilePath()), plugins.FetchRegistryIndex)
}

// register is Register with the settings and the index fetch passed in, so tests do not read the
// real settings file or touch the network.
func register(r *api.Router, ev api.Events, host *plugins.Host, uploads *uploads.Store, settings registrySettings, fetch registryFetch) {
	host.SetEvents(ev)

	r.Handle("PluginsList", api.NoInput(func(context.Context) (any, error) {
		return host.List(), nil
	}))
	// PluginRegistry returns the configured community index. A missing URL makes no request, and a
	// failed fetch comes back in the result's error field so the UI can retry without a failed call.
	r.Handle("PluginRegistry", api.NoInput(func(ctx context.Context) (plugins.RegistryIndexResult, error) {
		return communityRegistry(ctx, settings, fetch), nil
	}))
	// Trusted keys are the ed25519 public keys a SIGNATURE must match. Add and remove return the
	// refreshed list, so the UI never shows a stale key after a change.
	r.Handle("PluginTrustedKeys", api.NoInput(func(context.Context) (any, error) {
		return host.TrustedKeys()
	}))
	r.Handle("PluginTrustedKeyAdd", api.Typed(func(_ context.Context, in trustedKeyInput) (any, error) {
		return host.AddTrustedKey(in.Key)
	}))
	r.Handle("PluginTrustedKeyRemove", api.Typed(func(_ context.Context, in trustedKeyInput) (any, error) {
		return host.RemoveTrustedKey(in.Key)
	}))
	r.Handle("PluginInstall", api.Typed(func(ctx context.Context, in plugins.InstallRequest) (any, error) {
		info, err := host.Install(ctx, in)
		removeUpload(uploads, in.Source)
		return info, err
	}))
	// PluginInstallOrUpdate installs a new plugin, or updates the installed one with the same id.
	// The drop and pick flows use it so a newer package replaces the old version in place.
	r.Handle("PluginInstallOrUpdate", api.Typed(func(ctx context.Context, in plugins.InstallRequest) (any, error) {
		info, err := host.InstallOrUpdate(ctx, in)
		removeUpload(uploads, in.Source)
		return info, err
	}))
	r.Handle("PluginUpdate", api.Typed(func(ctx context.Context, in updateInput) (any, error) {
		return host.Update(ctx, plugins.UpdateRequest{
			ID:                  in.ID,
			Source:              in.Source,
			SHA256:              in.SHA256,
			AllowUnsigned:       in.AllowUnsigned,
			AcceptedPermissions: in.AcceptedPermissions,
		})
	}))
	r.Handle("PluginRemove", api.Typed(func(_ context.Context, in removeInput) (any, error) {
		return nil, host.Remove(in.ID, in.DeleteData)
	}))
	// PluginDevLink runs a plugin from a source folder, and PluginDevUnlink hands it back to the installed copy.
	r.Handle("PluginDevLink", api.Typed(func(_ context.Context, in devLinkInput) (any, error) {
		return host.DevLink(in.Path)
	}))
	r.Handle("PluginDevUnlink", api.Typed(func(_ context.Context, in idInput) (any, error) {
		return nil, host.DevUnlink(in.ID)
	}))
	r.Handle("PluginSetEnabled", api.Typed(func(_ context.Context, in setEnabledInput) (any, error) {
		return host.SetEnabled(in.ID, in.Enabled)
	}))
	r.Handle("PluginRollback", api.Typed(func(_ context.Context, in idInput) (any, error) {
		return host.Rollback(in.ID)
	}))
	r.Handle("PluginPermissionsPreview", api.Typed(func(ctx context.Context, in sourceInput) (any, error) {
		return host.PermissionsPreview(ctx, in.Source, in.SHA256)
	}))
	r.Handle("PluginSettingsGet", api.Typed(func(_ context.Context, in idInput) (any, error) {
		return host.SettingsGet(in.ID)
	}))
	r.Handle("PluginSettingsSet", api.Typed(func(_ context.Context, in settingsSetInput) (any, error) {
		if err := host.SettingsSet(in.ID, in.Settings); err != nil {
			return nil, err
		}
		ev.Emit("plugin-settings", map[string]any{"id": in.ID})
		return nil, nil
	}))
	r.Handle("PluginCall", api.Typed(func(ctx context.Context, in callInput) (any, error) {
		return pluginCall(ctx, host, in)
	}))
	r.Handle("PluginLogs", api.Typed(func(_ context.Context, in logsInput) (any, error) {
		return host.Logs(in.ID, in.Lines)
	}))
	r.Handle("PluginUpdatesCheck", api.Typed(func(ctx context.Context, in idInput) (any, error) {
		return host.CheckUpdates(ctx, in.ID)
	}))
	r.Handle("PluginUpdatesStatus", api.NoInput(func(context.Context) (any, error) {
		return host.UpdatesStatus(), nil
	}))
}

// communityRegistry returns the configured community plugin index. It reads the URL on every call,
// so a settings change applies without a restart, and an empty URL never starts a request.
func communityRegistry(ctx context.Context, settings registrySettings, fetch registryFetch) plugins.RegistryIndexResult {
	rawURL := strings.TrimSpace(settings.Load().PluginRegistryURL)
	result := plugins.RegistryIndexResult{Entries: []plugins.RegistryIndexEntry{}}
	if rawURL == "" {
		return result
	}
	result.Configured = true
	result.URL = rawURL

	index, err := fetch(ctx, rawURL)
	if err != nil {
		logger.Warn("plugins", "could not load the community plugin index: %v", err)
		result.Error = err.Error()
		return result
	}
	if index.Plugins != nil {
		result.Entries = index.Plugins
	}
	return result
}
