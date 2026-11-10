package plugins

import (
	"context"
	"encoding/json"
	"time"
)

const (
	defaultBrowserTimeout = 2 * time.Minute
	maxBrowserTimeout     = 10 * time.Minute
)

// BrowserAssist opens a browser for the user and returns the link it captures.
// The launcher supplies the implementation. This package does not drive a browser.
type BrowserAssist interface {
	Assist(ctx context.Context, request BrowserAssistRequest) (capturedURL string, err error)
}

// BrowserAssistRequest asks the browser to open a page and capture one link.
type BrowserAssistRequest struct {
	Plugin  string
	URL     string
	Message string
	Timeout time.Duration
}

type browserParams struct {
	URL       string `json:"url"`
	Message   string `json:"message"`
	TimeoutMs int    `json:"timeoutMs"`
}

// serviceBrowserAssist runs the browser hook. The result is {capturedUrl: string or null}.
func serviceBrowserAssist(ctx context.Context, env *pluginEnv, params json.RawMessage) (any, error) {
	req, err := decodeParams[browserParams](params)
	if err != nil {
		return nil, err
	}
	if _, err := parseHTTPTarget(req.URL); err != nil {
		return nil, err
	}
	if env.deps.browser == nil {
		return nil, internalError("browser assist is not set up")
	}
	captured, err := env.deps.browser.Assist(ctx, BrowserAssistRequest{
		Plugin:  env.id,
		URL:     req.URL,
		Message: req.Message,
		Timeout: browserTimeout(req.TimeoutMs),
	})
	if err != nil {
		return nil, internalError(err.Error())
	}
	var capturedURL any
	if captured != "" {
		capturedURL = captured
	}
	return map[string]any{"capturedUrl": capturedURL}, nil
}

func browserTimeout(ms int) time.Duration {
	if ms <= 0 {
		return defaultBrowserTimeout
	}
	return min(time.Duration(ms)*time.Millisecond, maxBrowserTimeout)
}
