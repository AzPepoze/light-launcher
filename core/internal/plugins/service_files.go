package plugins

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"light-launcher/core/internal/downloads"
)

type extractParams struct {
	Archive     string `json:"archive"`
	Destination string `json:"destination"`
}

// serviceExtract unpacks a zip into a folder. Both paths must sit inside a granted $FOLDER.
func serviceExtract(_ context.Context, env *pluginEnv, params json.RawMessage) (any, error) {
	req, err := decodeParams[extractParams](params)
	if err != nil {
		return nil, err
	}
	archive, err := env.grantedPath(req.Archive)
	if err != nil {
		return nil, err
	}
	destination, root, err := env.grantedPathUnderRoot(req.Destination)
	if err != nil {
		return nil, err
	}
	if err := checkNewPath(root, destination); err != nil {
		return nil, invalidParams(err.Error())
	}
	if err := extractZip(archive, destination); err != nil {
		return nil, internalError(err.Error())
	}
	return map[string]string{"destination": destination}, nil
}

// grantedPath returns a clean absolute path that lies inside one of the plugin's granted folders.
func (env *pluginEnv) grantedPath(path string) (string, error) {
	clean, _, err := env.grantedPathUnderRoot(path)
	return clean, err
}

// grantedPathUnderRoot also returns the granted folder that contains the path.
func (env *pluginEnv) grantedPathUnderRoot(path string) (string, string, error) {
	if !filepath.IsAbs(path) {
		return "", "", invalidParams("path must be absolute")
	}
	clean := filepath.Clean(path)
	for _, folder := range env.manifest.Permissions.FS {
		root, ok := env.deps.folders[folder]
		if ok && isWithin(root, clean) {
			return clean, root, nil
		}
	}
	return "", "", &RPCError{Code: CodePermissionDenied, Message: "path is outside the folders granted in plugin.json"}
}

// checkNewPath checks names the plugin adds below a granted folder.
func checkNewPath(root, path string) error {
	rel, err := filepath.Rel(filepath.Clean(root), path)
	if err != nil {
		return err
	}
	if rel == "." {
		return nil
	}
	for _, segment := range strings.Split(filepath.ToSlash(rel), "/") {
		if err := downloads.ValidateName(segment); err != nil {
			return err
		}
	}
	return nil
}

// isWithin reports whether path is root or lies below it.
func isWithin(root, path string) bool {
	rel, err := filepath.Rel(filepath.Clean(root), path)
	if err != nil || filepath.IsAbs(rel) {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
