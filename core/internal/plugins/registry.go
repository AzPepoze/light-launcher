package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/logger"
)

const (
	methodPrepareDownload = "source.prepareDownload"
	methodBrowse          = "source.browse"
	methodListFilters     = "source.listFilters"
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

// Browse runs source.browse for a filtered, sorted page of games. It never saves them.
func (r *Registry) Browse(ctx context.Context, plugin string, request contract.BrowseRequest) (contract.BrowsePage, error) {
	var page contract.BrowsePage
	err := r.callWithTimeout(ctx, plugin, methodBrowse, contract.BrowseTimeout, request, &page)
	if isInvalidCursor(err) {
		return contract.BrowsePage{}, contract.ErrInvalidCursor
	}
	return page, err
}

// invalidCursorCode is the error data code a plugin sends when it does not know a browse cursor.
const invalidCursorCode = "invalid_cursor"

// isInvalidCursor reports whether a plugin error says the browse cursor is unknown.
func isInvalidCursor(err error) bool {
	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) {
		return false
	}
	data, ok := rpcErr.Data.(map[string]any)
	return ok && data["code"] == invalidCursorCode
}

// ListFilters runs source.listFilters for the values of one dynamic sidebar group.
func (r *Registry) ListFilters(ctx context.Context, plugin string, request contract.FiltersRequest) (contract.FiltersPage, error) {
	var page contract.FiltersPage
	err := r.callWithTimeout(ctx, plugin, methodListFilters, contract.ListFiltersTimeout, request, &page)
	return page, err
}

// callWithTimeout runs a source method and stops waiting after timeout.
func (r *Registry) callWithTimeout(ctx context.Context, plugin, method string, timeout time.Duration, params, out any) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return r.Call(ctx, plugin, method, params, out)
}

// PrepareDownload runs source.prepareDownload and builds the host request from its answer.
// The plugin needs the $GAMES fs grant for games-folder downloads and installs, as with downloads.enqueue.
func (r *Registry) PrepareDownload(ctx context.Context, plugin string, params any, gameID int64) (contract.DownloadRequest, error) {
	sup, err := r.host.supervisorFor(plugin)
	if err != nil {
		return contract.DownloadRequest{}, fmt.Errorf("prepare download with %s: %w", plugin, err)
	}
	ctx, cancel := prepareContext(ctx)
	defer cancel()
	var raw json.RawMessage
	if err := sup.Call(ctx, methodPrepareDownload, params, &raw); err != nil {
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

// prepareContext bounds one source.prepareDownload call. A person may finish a browser step inside it.
func prepareContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, contract.PrepareDownloadTimeout)
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
