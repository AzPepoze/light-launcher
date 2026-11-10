package plugins

import (
	"context"
	"encoding/json"
	"fmt"
)

// serviceFunc runs one plugin request after its permission check has passed.
type serviceFunc func(ctx context.Context, env *pluginEnv, params json.RawMessage) (any, error)

// service is a host service a plugin can call.
// needs reports whether the manifest grants it. A nil needs means no manifest grant is required.
type service struct {
	needs     func(Permissions) bool
	grantName string
	run       serviceFunc
}

// services lists every method a plugin may call on the host, with its grant.
var services = map[string]service{
	"http.request":      {run: serviceHTTPRequest},
	"secrets.get":       {needs: hasSecrets, grantName: "secrets", run: serviceSecretsGet},
	"secrets.set":       {needs: hasSecrets, grantName: "secrets", run: serviceSecretsSet},
	"secrets.delete":    {needs: hasSecrets, grantName: "secrets", run: serviceSecretsDelete},
	"storage.get":       {run: serviceStorageGet},
	"storage.set":       {run: serviceStorageSet},
	"settings.get":      {run: serviceSettingsGet},
	"downloads.enqueue": {run: serviceDownloadsEnqueue},
	"library.upsert":    {run: serviceLibraryUpsert},
	"extract":           {needs: hasFolders, grantName: "fs", run: serviceExtract},
	"browser.assist":    {needs: hasBrowser, grantName: "browser", run: serviceBrowserAssist},
}

func hasSecrets(p Permissions) bool { return p.Secrets }
func hasBrowser(p Permissions) bool { return p.Browser }
func hasFolders(p Permissions) bool { return len(p.FS) > 0 }

// decodeParams reads a JSON params object. Missing params decode to the zero value.
func decodeParams[T any](raw json.RawMessage) (T, error) {
	var value T
	if len(raw) == 0 {
		return value, nil
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return value, invalidParams(err.Error())
	}
	return value, nil
}

func invalidParams(detail string) error {
	return &RPCError{Code: CodeInvalidParams, Message: "invalid params: " + detail}
}

func permissionDenied(method, grant string) error {
	return &RPCError{
		Code:    CodePermissionDenied,
		Message: fmt.Sprintf("%s needs the %s permission in plugin.json", method, grant),
	}
}

func internalError(detail string) error {
	return &RPCError{Code: CodeInternal, Message: detail}
}
