package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
)

// shellEvent is one line of the NDJSON output of the assist mode.
type shellEvent struct {
	Event   string `json:"event"`
	Message string `json:"message,omitempty"`
}

// downloadEvent is the download line. Every field is always present.
type downloadEvent struct {
	Event    string            `json:"event"`
	URL      string            `json:"url"`
	Filename string            `json:"filename"`
	Size     int64             `json:"size"`
	Headers  map[string]string `json:"headers"`
}

func readyEvent() shellEvent {
	return shellEvent{Event: "ready"}
}

func closedEvent() shellEvent {
	return shellEvent{Event: "closed"}
}

func errorEvent(message string) shellEvent {
	return shellEvent{Event: "error", Message: message}
}

// newDownloadEvent describes a download the page started. The size is -1 when unknown.
func newDownloadEvent(rawURL, filename string, size int64, headers map[string]string) downloadEvent {
	if headers == nil {
		headers = map[string]string{}
	}
	return downloadEvent{Event: "download", URL: rawURL, Filename: filename, Size: size, Headers: headers}
}

// eventEmitter writes one JSON object per line and flushes every line.
type eventEmitter struct {
	mu     sync.Mutex
	writer *bufio.Writer
	err    error
}

func newEventEmitter(out io.Writer) *eventEmitter {
	return &eventEmitter{writer: bufio.NewWriter(out)}
}

func (e *eventEmitter) emit(event any) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.err != nil {
		return e.err
	}
	data, err := json.Marshal(event)
	if err != nil {
		e.err = fmt.Errorf("encode event: %w", err)
		return e.err
	}
	data = append(data, '\n')
	if _, err := e.writer.Write(data); err != nil {
		e.err = fmt.Errorf("write event: %w", err)
		return e.err
	}
	if err := e.writer.Flush(); err != nil {
		e.err = fmt.Errorf("flush event: %w", err)
		return e.err
	}
	return nil
}

func (e *eventEmitter) Err() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.err
}

type cookiePair struct {
	Name  string
	Value string
}

// cookieHeader builds the Cookie header from cookies that match the download URL.
func cookieHeader(cookies []cookiePair) string {
	parts := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie.Name == "" {
			continue
		}
		parts = append(parts, cookie.Name+"="+cookie.Value)
	}
	return strings.Join(parts, "; ")
}

// downloadHeaders keeps only the headers the shell knows.
func downloadHeaders(cookies []cookiePair, referer, userAgent string) map[string]string {
	headers := map[string]string{}
	if value := cookieHeader(cookies); value != "" {
		headers["Cookie"] = value
	}
	if referer != "" {
		headers["Referer"] = referer
	}
	if userAgent != "" {
		headers["User-Agent"] = userAgent
	}
	return headers
}

// resultFileName returns the file name of a download result path, on either separator.
func resultFileName(path string) string {
	if index := strings.LastIndexAny(path, `\/`); index >= 0 {
		return path[index+1:]
	}
	return path
}

// bannerScript injects a small non-interactive banner above the page. It never blocks
// clicks or changes the page itself, and it is not added to child frames.
func bannerScript(message string) string {
	if message == "" {
		return ""
	}
	text, err := json.Marshal(message)
	if err != nil {
		return ""
	}
	return `(function () {
	if (window.top !== window) return;
	window.addEventListener("DOMContentLoaded", function () {
		if (!document.body) return;
		var banner = document.createElement("div");
		banner.textContent = ` + string(text) + `;
		banner.style.cssText = "position:fixed;top:0;left:0;right:0;z-index:2147483647;" +
			"padding:6px 12px;background:#1f2937;color:#fff;font:13px system-ui,sans-serif;" +
			"pointer-events:none;text-align:center;";
		document.body.appendChild(banner);
	});
})();`
}
