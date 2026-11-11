// Package downloadsvc exposes the download manager to the UI through the api router.
package downloadsvc

import (
	"context"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/downloads"
)

// idRequest names one download.
type idRequest struct {
	ID int64 `json:"id"`
}

// removeRequest deletes a download, and optionally its files.
type removeRequest struct {
	ID         int64 `json:"id"`
	DeleteFile bool  `json:"deleteFile"`
}

// speedLimitRequest sets the total speed limit in KiB/s; zero means unlimited.
type speedLimitRequest struct {
	SpeedLimitKBps int `json:"speedLimitKBps"`
}

// okResponse is returned by actions that have no other result.
type okResponse struct {
	OK bool `json:"ok"`
}

// Register adds the download methods to the router. The manager sends its own events, so the
// events argument is not used here. It keeps the signature the same as the other services.
func Register(r *api.Router, _ api.Events, m *downloads.Manager) {
	r.Handle("DownloadsList", api.NoInput(m.List))
	r.Handle("DownloadEnqueue", api.Typed(enqueue(m)))
	r.Handle("DownloadPause", idAction(m.Pause))
	r.Handle("DownloadResume", idAction(m.Resume))
	r.Handle("DownloadCancel", idAction(m.Cancel))
	r.Handle("DownloadRetry", idAction(m.Retry))
	r.Handle("DownloadRemove", api.Typed(remove(m)))
	r.Handle("DownloadSetSpeedLimit", api.Typed(setSpeedLimit(m)))
	r.Handle("DownloadsStatus", api.NoInput(status(m)))
	r.Handle("DownloadsGetSettings", api.NoInput(getSettings(m)))
	r.Handle("DownloadsSetSettings", api.Typed(setSettings(m)))
}

func enqueue(m *downloads.Manager) func(context.Context, contract.DownloadRequest) (downloads.View, error) {
	return func(ctx context.Context, req contract.DownloadRequest) (downloads.View, error) {
		id, err := m.Enqueue(ctx, req)
		if err != nil {
			return downloads.View{}, err
		}
		return m.Get(ctx, id)
	}
}

func remove(m *downloads.Manager) func(context.Context, removeRequest) (okResponse, error) {
	return func(ctx context.Context, in removeRequest) (okResponse, error) {
		if err := m.Remove(ctx, in.ID, in.DeleteFile); err != nil {
			return okResponse{}, err
		}
		return okResponse{OK: true}, nil
	}
}

func setSpeedLimit(m *downloads.Manager) func(context.Context, speedLimitRequest) (downloads.Settings, error) {
	return func(_ context.Context, in speedLimitRequest) (downloads.Settings, error) {
		return m.SetSpeedLimit(in.SpeedLimitKBps), nil
	}
}

func status(m *downloads.Manager) func(context.Context) (downloads.Status, error) {
	return func(context.Context) (downloads.Status, error) {
		return m.Status(), nil
	}
}

func getSettings(m *downloads.Manager) func(context.Context) (downloads.Settings, error) {
	return func(context.Context) (downloads.Settings, error) {
		return m.Settings(), nil
	}
}

func setSettings(m *downloads.Manager) func(context.Context, downloads.Settings) (downloads.Settings, error) {
	return func(_ context.Context, in downloads.Settings) (downloads.Settings, error) {
		return m.SetSettings(in), nil
	}
}

// idAction wraps a manager action that takes only a download ID.
func idAction(fn func(context.Context, int64) error) api.Handler {
	return api.Typed(func(ctx context.Context, in idRequest) (okResponse, error) {
		if err := fn(ctx, in.ID); err != nil {
			return okResponse{}, err
		}
		return okResponse{OK: true}, nil
	})
}
