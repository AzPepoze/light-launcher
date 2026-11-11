// Package bridge holds the page-to-shell request protocol shared by the shells.
package bridge

import (
	"encoding/json"
	"fmt"
	"net/url"

	"light-launcher/shell/common/urlx"
)

// Request is one method call from the page.
type Request struct {
	ID     int    `json:"id"`
	Method string `json:"method"`
}

// ParseRequest reads one request body. A request needs a positive id.
func ParseRequest(body string) (Request, bool) {
	var request Request
	if err := json.Unmarshal([]byte(body), &request); err != nil || request.ID <= 0 {
		return Request{}, false
	}
	return request, true
}

// Allows reports whether a page may use the native bridge. Only the launcher
// origin may, and an empty launcher origin allows nothing.
func Allows(pageURI, launcherOrigin string) bool {
	if launcherOrigin == "" {
		return false
	}
	target, err := url.Parse(pageURI)
	if err != nil {
		return false
	}
	return urlx.Origin(target) == launcherOrigin
}

// ResolveScript builds the JavaScript that settles one pending promise. A cancel
// resolves to null.
func ResolveScript(id int, path string) (string, error) {
	var value any
	if path != "" {
		value = path
	}
	literal, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("window.llShell._resolve(%d, %s)", id, literal), nil
}
