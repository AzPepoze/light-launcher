// Package pluginsvc exposes the plugin host to the UI as router methods.
package pluginsvc

import (
	"context"
	"encoding/json"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/plugins"
	"light-launcher/core/internal/uploads"
)

// updateInput updates a plugin. AcceptedPermissions is needed when the new version asks for more permissions.
type updateInput struct {
	ID                  string   `json:"id"`
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

type setEnabledInput struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

type sourceInput struct {
	Source string `json:"source"`
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

// removeUpload deletes the upload that source is in. A failure only leaves a folder to clear on the next start.
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

// Register adds the plugin methods to r. Host events such as "plugin-state" go to ev.
// Installs from an upload remove that upload once it is installed.
func Register(r *api.Router, ev api.Events, host *plugins.Host, uploads *uploads.Store) {
	host.SetEvents(ev)

	r.Handle("PluginsList", api.NoInput(func(context.Context) (any, error) {
		return host.List(), nil
	}))
	r.Handle("PluginInstall", api.Typed(func(ctx context.Context, in plugins.InstallRequest) (any, error) {
		info, err := host.Install(ctx, in)
		if err == nil {
			removeUpload(uploads, in.Source)
		}
		return info, err
	}))
	r.Handle("PluginUpdate", api.Typed(func(ctx context.Context, in updateInput) (any, error) {
		return host.Update(ctx, in.ID, in.AllowUnsigned, in.AcceptedPermissions)
	}))
	r.Handle("PluginRemove", api.Typed(func(_ context.Context, in removeInput) (any, error) {
		return nil, host.Remove(in.ID, in.DeleteData)
	}))
	r.Handle("PluginSetEnabled", api.Typed(func(_ context.Context, in setEnabledInput) (any, error) {
		return host.SetEnabled(in.ID, in.Enabled)
	}))
	r.Handle("PluginRollback", api.Typed(func(_ context.Context, in idInput) (any, error) {
		return host.Rollback(in.ID)
	}))
	r.Handle("PluginPermissionsPreview", api.Typed(func(ctx context.Context, in sourceInput) (any, error) {
		return host.PermissionsPreview(ctx, in.Source)
	}))
	r.Handle("PluginSettingsGet", api.Typed(func(_ context.Context, in idInput) (any, error) {
		return host.SettingsGet(in.ID)
	}))
	r.Handle("PluginSettingsSet", api.Typed(func(_ context.Context, in settingsSetInput) (any, error) {
		return nil, host.SettingsSet(in.ID, in.Settings)
	}))
	r.Handle("PluginCall", api.Typed(func(ctx context.Context, in callInput) (any, error) {
		return pluginCall(ctx, host, in)
	}))
	r.Handle("PluginLogs", api.Typed(func(_ context.Context, in logsInput) (any, error) {
		return host.Logs(in.ID, in.Lines)
	}))
}
