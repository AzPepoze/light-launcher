package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/downloads"
)

// downloadOptions are the folder and install fields shared by downloads.enqueue and source.prepareDownload.
type downloadOptions struct {
	Root             string `json:"root"`
	Subfolder        string `json:"subfolder"`
	Install          bool   `json:"install"`
	InstallSubfolder string `json:"installSubfolder"`
	DeleteArchive    *bool  `json:"deleteArchive"`
}

// downloadParams are the options of downloads.enqueue. Every field except url and filename is optional.
type downloadParams struct {
	URL      string            `json:"url"`
	Filename string            `json:"filename"`
	Headers  map[string]string `json:"headers"`
	GameID   int64             `json:"gameId"`
	downloadOptions
}

// serviceDownloadsEnqueue queues a file for the host downloader. The URL host must be in the network grant.
func serviceDownloadsEnqueue(ctx context.Context, env *pluginEnv, params json.RawMessage) (any, error) {
	req, err := decodeParams[downloadParams](params)
	if err != nil {
		return nil, err
	}
	if err := env.checkDownloadURL(req.URL); err != nil {
		return nil, err
	}
	if !isPlainFileName(req.Filename) {
		return nil, invalidParams("filename must be a plain file name")
	}
	hostReq, err := env.downloadRequest(req)
	if err != nil {
		return nil, err
	}
	if env.deps.downloads == nil {
		return nil, internalError("downloads are not available")
	}
	id, err := env.deps.downloads.Enqueue(ctx, hostReq)
	if err != nil {
		return nil, enqueueError(err)
	}
	return map[string]int64{"id": id}, nil
}

// checkDownloadURL checks that the URL is http(s) and its host is in the network grant.
func (env *pluginEnv) checkDownloadURL(rawURL string) error {
	target, err := parseHTTPTarget(rawURL)
	if err != nil {
		return err
	}
	if !env.manifest.Permissions.allowsHost(target.Hostname()) {
		return hostDenied(target.Hostname())
	}
	return nil
}

// downloadRequest checks the folder options and builds the host request. A plugin needs the fs
// grant for $GAMES to install, or to save into the games folder.
func (env *pluginEnv) downloadRequest(req downloadParams) (contract.DownloadRequest, error) {
	return env.requestFor(req.GameID, req.URL, req.Headers, req.Filename, req.downloadOptions)
}

// preparedDownloadRequest builds the host request from a source.prepareDownload answer.
func (env *pluginEnv) preparedDownloadRequest(gameID int64, result PrepareDownloadResult) (contract.DownloadRequest, error) {
	return env.requestFor(gameID, result.URL, result.Headers, result.Filename, result.downloadOptions())
}

func (env *pluginEnv) requestFor(gameID int64, rawURL string, headers map[string]string, filename string, opts downloadOptions) (contract.DownloadRequest, error) {
	root := opts.Root
	if root == "" {
		root = contract.RootDownloads
	}
	req := contract.DownloadRequest{
		GameID:           gameID,
		URL:              rawURL,
		Headers:          headers,
		Filename:         filename,
		Install:          opts.Install,
		Root:             root,
		Subfolder:        opts.Subfolder,
		InstallSubfolder: opts.InstallSubfolder,
		DeleteArchive:    opts.DeleteArchive,
	}
	if err := downloads.ValidateRequest(req); err != nil {
		return contract.DownloadRequest{}, invalidParams(err.Error())
	}
	if gameID < 0 {
		return contract.DownloadRequest{}, invalidParams("gameId must not be negative")
	}
	if root == contract.RootGames || opts.Install {
		if err := env.checkFolderGrant("$GAMES"); err != nil {
			return contract.DownloadRequest{}, err
		}
	}
	return req, nil
}

// checkFolderGrant reports permission denied unless plugin.json grants the folder name.
func (env *pluginEnv) checkFolderGrant(name string) error {
	for _, granted := range env.manifest.Permissions.FS {
		if granted == name {
			return nil
		}
	}
	return &RPCError{
		Code:    CodePermissionDenied,
		Message: fmt.Sprintf("this download needs %s in the fs permission of plugin.json", name),
	}
}

// enqueueError maps a rejected request to invalid params and anything else to an internal error.
func enqueueError(err error) error {
	if errors.Is(err, contract.ErrBadFolder) || errors.Is(err, contract.ErrBadRequest) {
		return invalidParams(err.Error())
	}
	return internalError(err.Error())
}
