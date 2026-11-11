package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/logger"
)

var errNotSourceMethod = errors.New("only source methods can be called on a plugin source")

// PrepareDownloadResult is the answer to source.prepareDownload. Filename and Size are set only when the plugin knows them.
type PrepareDownloadResult struct {
	URL      string            `json:"url"`
	Headers  map[string]string `json:"headers"`
	Filename string            `json:"filename,omitempty"`
	Size     int64             `json:"size,omitempty"`

	Root             string `json:"root,omitempty"`
	Subfolder        string `json:"subfolder,omitempty"`
	Install          bool   `json:"install,omitempty"`
	InstallSubfolder string `json:"installSubfolder,omitempty"`
	DeleteArchive    *bool  `json:"deleteArchive,omitempty"`
}

func (r PrepareDownloadResult) downloadOptions() downloadOptions {
	return downloadOptions{
		Root:             r.Root,
		Subfolder:        r.Subfolder,
		Install:          r.Install,
		InstallSubfolder: r.InstallSubfolder,
		DeleteArchive:    r.DeleteArchive,
	}
}

// Registry implements contract.Sources over the enabled plugins.
type Registry struct {
	host *Host
}

var _ contract.Sources = (*Registry)(nil)

// NewRegistry returns the source view of a host.
func NewRegistry(host *Host) *Registry {
	return &Registry{host: host}
}

// Sources lists the sources of every enabled plugin that installed cleanly.
func (r *Registry) Sources() []contract.SourceInfo {
	var sources []contract.SourceInfo
	for _, info := range r.host.List() {
		if info.Enabled && info.Status != "broken" {
			sources = append(sources, info.Sources...)
		}
	}
	return sources
}

// Call runs a "source." method on a plugin and decodes the answer into out.
func (r *Registry) Call(ctx context.Context, plugin, method string, params any, out any) error {
	if !strings.HasPrefix(method, "source.") {
		return errNotSourceMethod
	}
	sup, err := r.host.supervisorFor(plugin)
	if err != nil {
		return err
	}
	return sup.Call(ctx, method, params, out)
}

// PrepareDownload runs source.prepareDownload and builds the host request from its answer.
// The plugin needs the $GAMES fs grant for games-folder downloads and installs, as with downloads.enqueue.
func (r *Registry) PrepareDownload(ctx context.Context, plugin string, params any, gameID int64) (contract.DownloadRequest, error) {
	sup, err := r.host.supervisorFor(plugin)
	if err != nil {
		return contract.DownloadRequest{}, fmt.Errorf("prepare download with %s: %w", plugin, err)
	}
	var raw json.RawMessage
	if err := sup.Call(ctx, "source.prepareDownload", params, &raw); err != nil {
		return contract.DownloadRequest{}, fmt.Errorf("prepare download with %s: %w", plugin, err)
	}
	result, err := decodePrepareDownload(raw, sup.cfg.env.category())
	if err != nil {
		return contract.DownloadRequest{}, fmt.Errorf("prepare download with %s: %w", plugin, err)
	}
	if err := sup.cfg.env.checkDownloadURL(result.URL); err != nil {
		return contract.DownloadRequest{}, fmt.Errorf("prepare download with %s: %w", plugin, err)
	}
	if result.Filename != "" && !isPlainFileName(result.Filename) {
		return contract.DownloadRequest{}, invalidParams("filename must be a plain file name")
	}
	req, err := sup.cfg.env.preparedDownloadRequest(gameID, result)
	if err != nil {
		return contract.DownloadRequest{}, fmt.Errorf("prepare download with %s: %w", plugin, err)
	}
	return req, nil
}

// StartDownload runs source.prepareDownload and queues the file with the host downloader.
func (r *Registry) StartDownload(ctx context.Context, plugin string, params any, gameID int64) (int64, error) {
	req, err := r.PrepareDownload(ctx, plugin, params, gameID)
	if err != nil {
		return 0, err
	}
	if r.host.deps.downloads == nil {
		return 0, errors.New("downloads are not available")
	}
	id, err := r.host.deps.downloads.Enqueue(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("queue prepared download: %w", err)
	}
	return id, nil
}

// decodePrepareDownload reads the answer and logs optional fields this host does not know.
func decodePrepareDownload(raw json.RawMessage, category string) (PrepareDownloadResult, error) {
	var result PrepareDownloadResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, invalidParams("source.prepareDownload result: " + err.Error())
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return result, invalidParams("source.prepareDownload result: " + err.Error())
	}
	for name := range fields {
		if !knownPrepareDownloadField(name) {
			logger.Debug(category, "source.prepareDownload ignored unknown field %q", name)
		}
	}
	return result, nil
}

func knownPrepareDownloadField(name string) bool {
	switch name {
	case "url", "headers", "filename", "size", "root", "subfolder", "install", "installSubfolder", "deleteArchive":
		return true
	}
	return false
}
