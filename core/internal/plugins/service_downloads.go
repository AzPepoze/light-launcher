package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/downloads"
)

// downloadParams are the options of downloads.enqueue. Every field except url and filename is optional.
type downloadParams struct {
	URL              string            `json:"url"`
	Filename         string            `json:"filename"`
	Headers          map[string]string `json:"headers"`
	GameID           int64             `json:"gameId"`
	Root             string            `json:"root"`
	Subfolder        string            `json:"subfolder"`
	Install          bool              `json:"install"`
	InstallSubfolder string            `json:"installSubfolder"`
	DeleteArchive    *bool             `json:"deleteArchive"`
}

// serviceDownloadsEnqueue queues a file for the host downloader. The URL host must be in the network grant.
func serviceDownloadsEnqueue(ctx context.Context, env *pluginEnv, params json.RawMessage) (any, error) {
	req, err := decodeParams[downloadParams](params)
	if err != nil {
		return nil, err
	}
	target, err := parseHTTPTarget(req.URL)
	if err != nil {
		return nil, err
	}
	if !env.manifest.Permissions.allowsHost(target.Hostname()) {
		return nil, hostDenied(target.Hostname())
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

// downloadRequest checks the folder options and builds the host request. A plugin needs the fs
// grant for $GAMES to install, or to save into the games folder.
func (env *pluginEnv) downloadRequest(req downloadParams) (contract.DownloadRequest, error) {
	root := req.Root
	if root == "" {
		root = contract.RootDownloads
	}
	if root != contract.RootDownloads && root != contract.RootGames {
		return contract.DownloadRequest{}, invalidParams(`root must be "downloads" or "games"`)
	}
	if err := downloads.ValidateSubfolder(req.Subfolder); err != nil {
		return contract.DownloadRequest{}, invalidParams(err.Error())
	}
	if err := downloads.ValidateSubfolder(req.InstallSubfolder); err != nil {
		return contract.DownloadRequest{}, invalidParams(err.Error())
	}
	if req.GameID < 0 {
		return contract.DownloadRequest{}, invalidParams("gameId must not be negative")
	}
	if root == contract.RootGames || req.Install {
		if err := env.checkFolderGrant("$GAMES"); err != nil {
			return contract.DownloadRequest{}, err
		}
	}
	return contract.DownloadRequest{
		GameID:           req.GameID,
		URL:              req.URL,
		Headers:          req.Headers,
		Filename:         req.Filename,
		Install:          req.Install,
		Root:             root,
		Subfolder:        req.Subfolder,
		InstallSubfolder: req.InstallSubfolder,
		DeleteArchive:    req.DeleteArchive,
	}, nil
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
