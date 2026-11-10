package plugins

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
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
	destination, err := env.grantedPath(req.Destination)
	if err != nil {
		return nil, err
	}
	if err := extractZip(archive, destination); err != nil {
		return nil, internalError(err.Error())
	}
	return map[string]string{"destination": destination}, nil
}

// grantedPath returns a clean absolute path that lies inside one of the plugin's granted folders.
func (env *pluginEnv) grantedPath(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", invalidParams("path must be absolute")
	}
	clean := filepath.Clean(path)
	for _, folder := range env.manifest.Permissions.FS {
		root, ok := env.deps.folders[folder]
		if ok && isWithin(root, clean) {
			return clean, nil
		}
	}
	return "", &RPCError{Code: CodePermissionDenied, Message: "path is outside the folders granted in plugin.json"}
}

// isWithin reports whether path is root or lies below it.
func isWithin(root, path string) bool {
	rel, err := filepath.Rel(filepath.Clean(root), path)
	if err != nil || filepath.IsAbs(rel) {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
