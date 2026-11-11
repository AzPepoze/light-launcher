package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
)

// bridgeScript defines window.llShell. It is injected into every page the shell loads.
//
//go:embed bridge.js
var bridgeScript string

type bridgeRequest struct {
	ID     int    `json:"id"`
	Method string `json:"method"`
}

func parseBridgeRequest(body string) (bridgeRequest, bool) {
	var req bridgeRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil || req.ID <= 0 {
		return bridgeRequest{}, false
	}
	return req, true
}

// allowsBridge reports whether a page may use window.llShell. Only the launcher origin may.
func allowsBridge(pageURI, launcherOrigin string) bool {
	if launcherOrigin == "" {
		return false
	}
	target, err := url.Parse(pageURI)
	if err != nil {
		return false
	}
	return urlOrigin(target) == launcherOrigin
}

// resolveScript builds the JavaScript that settles one pending promise. A cancel resolves to null.
func resolveScript(id int, path string) (string, error) {
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
