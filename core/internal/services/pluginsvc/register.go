// Package pluginsvc exposes the plugin host to the UI as router methods.
package pluginsvc

import (
	"context"
	"encoding/json"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/plugins"
)

type updateInput struct {
	ID            string `json:"id"`
	AllowUnsigned bool   `json:"allowUnsigned"`
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

type callInput struct {
	Plugin string          `json:"plugin"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type logsInput struct {
	ID    string `json:"id"`
	Lines int    `json:"lines"`
}

// Register adds the plugin methods to r. Host events such as "plugin-state" go to ev.
func Register(r *api.Router, ev api.Events, host *plugins.Host) {
	host.SetEvents(ev)

	r.Handle("PluginsList", api.NoInput(func(context.Context) (any, error) {
		return host.List(), nil
	}))
	r.Handle("PluginInstall", api.Typed(func(ctx context.Context, in plugins.InstallRequest) (any, error) {
		return host.Install(ctx, in)
	}))
	r.Handle("PluginUpdate", api.Typed(func(ctx context.Context, in updateInput) (any, error) {
		return host.Update(ctx, in.ID, in.AllowUnsigned)
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
		return host.CallPlugin(ctx, in.Plugin, in.Method, in.Params)
	}))
	r.Handle("PluginLogs", api.Typed(func(_ context.Context, in logsInput) (any, error) {
		return host.Logs(in.ID, in.Lines)
	}))
}
