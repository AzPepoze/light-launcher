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
	Assist(ctx context.Context, request BrowserAssistRequest) (BrowserCapture, error)
}

// BrowserAssistRequest asks the browser to open a page and capture one link.
type BrowserAssistRequest struct {
	Plugin  string
	URL     string
	Message string
	Token   string // the plugin's token, or the request id when it sent none
	Timeout time.Duration
	Notify  BrowserNotifier
}

// BrowserCapture is the download link the browser caught, with what the download needs.
type BrowserCapture struct {
	URL      string
	Filename string
	Size     int64
	Headers  map[string]string
}

// BrowserNotifier tells the plugin what the browser step is doing. Each call carries the request token.
type BrowserNotifier interface {
	// Waiting asks the user to act, such as press a download button.
	Waiting(message string)
	// Progress reports percent from 0 to 100.
	Progress(percent int, message string)
}

type browserParams struct {
	URL       string `json:"url"`
	Message   string `json:"message"`
	TimeoutMs int    `json:"timeoutMs"`
	Token     string `json:"token"`
}

// browserResult is the browser.assist answer. Unknown optional fields are left out.
type browserResult struct {
	CapturedURL *string           `json:"capturedUrl"`
	Filename    string            `json:"filename,omitempty"`
	Size        int64             `json:"size,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// serviceBrowserAssist runs the browser hook and returns what it captured.
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
	token := req.Token
	if token == "" {
		token = requestIDFrom(ctx)
	}
	capture, err := env.deps.browser.Assist(ctx, BrowserAssistRequest{
		Plugin:  env.id,
		URL:     req.URL,
		Message: req.Message,
		Token:   token,
		Timeout: browserTimeout(req.TimeoutMs),
		Notify:  pluginNotifier{env: env, token: token},
	})
	if err != nil {
		return nil, internalError(err.Error())
	}
	return browserResultOf(capture), nil
}

func browserResultOf(capture BrowserCapture) browserResult {
	result := browserResult{Filename: capture.Filename, Size: capture.Size, Headers: capture.Headers}
	if capture.URL != "" {
		result.CapturedURL = &capture.URL
	}
	return result
}

// pluginNotifier sends browser.waiting and browser.progress to the plugin that asked.
type pluginNotifier struct {
	env   *pluginEnv
	token string
}

func (n pluginNotifier) Waiting(message string) {
	n.env.notifyPlugin("browser.waiting", map[string]any{"token": n.token, "message": message})
}

func (n pluginNotifier) Progress(percent int, message string) {
	n.env.notifyPlugin("browser.progress", map[string]any{"token": n.token, "percent": percent, "message": message})
}

func browserTimeout(ms int) time.Duration {
	if ms <= 0 {
		return defaultBrowserTimeout
	}
	return min(time.Duration(ms)*time.Millisecond, maxBrowserTimeout)
}
