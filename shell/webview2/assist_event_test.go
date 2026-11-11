package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestEventEmitterWritesOneJSONObjectPerLine(t *testing.T) {
	var output bytes.Buffer
	emitter := newEventEmitter(&output)

	if err := emitter.emit(readyEvent()); err != nil {
		t.Fatalf("emit ready: %v", err)
	}
	headers := map[string]string{"Cookie": "session=abc", "Referer": "https://example.com", "User-Agent": "Example/1"}
	if err := emitter.emit(newDownloadEvent("https://cdn.example.com/file.zip", "file.zip", -1, headers)); err != nil {
		t.Fatalf("emit download: %v", err)
	}
	if err := emitter.emit(closedEvent()); err != nil {
		t.Fatalf("emit closed: %v", err)
	}

	lines := strings.Split(strings.TrimSuffix(output.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3: %q", len(lines), output.String())
	}
	for _, line := range lines {
		if strings.Contains(line, "\n") {
			t.Fatalf("event contains a raw newline: %q", line)
		}
	}

	assertEventField(t, lines[0], "event", "ready")
	assertEventField(t, lines[1], "event", "download")
	assertEventField(t, lines[1], "url", "https://cdn.example.com/file.zip")
	assertEventField(t, lines[1], "filename", "file.zip")
	assertEventField(t, lines[1], "size", float64(-1))
	assertEventField(t, lines[2], "event", "closed")

	var download map[string]any
	if err := json.Unmarshal([]byte(lines[1]), &download); err != nil {
		t.Fatalf("decode download: %v", err)
	}
	headersValue, ok := download["headers"].(map[string]any)
	if !ok {
		t.Fatalf("download headers = %#v, want an object", download["headers"])
	}
	if headersValue["Cookie"] != "session=abc" {
		t.Errorf("Cookie = %v, want session=abc", headersValue["Cookie"])
	}
	if _, ok := headersValue["X-Unknown"]; ok {
		t.Error("unexpected header key was written")
	}

	var ready map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &ready); err != nil {
		t.Fatalf("decode ready: %v", err)
	}
	if len(ready) != 1 {
		t.Errorf("ready event has extra fields: %#v", ready)
	}
}

func TestDownloadEventIncludesZeroSize(t *testing.T) {
	event := newDownloadEvent("https://example.com/empty", "empty", 0, nil)
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(data), `"size":0`) {
		t.Fatalf("download event without size: %s", data)
	}
	if !strings.Contains(string(data), `"headers":{}`) {
		t.Fatalf("download event without headers: %s", data)
	}
}

func TestErrorEventCarriesMessage(t *testing.T) {
	data, err := json.Marshal(errorEvent("cannot start"))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var event map[string]any
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if event["event"] != "error" || event["message"] != "cannot start" {
		t.Fatalf("error event = %#v", event)
	}
	if len(event) != 2 {
		t.Fatalf("error event has extra fields: %#v", event)
	}
}

func TestCookieHeader(t *testing.T) {
	tests := []struct {
		name    string
		cookies []cookiePair
		want    string
	}{
		{name: "none"},
		{name: "one", cookies: []cookiePair{{Name: "session", Value: "abc"}}, want: "session=abc"},
		{
			name:    "many keep order",
			cookies: []cookiePair{{Name: "a", Value: "1"}, {Name: "b", Value: "2"}},
			want:    "a=1; b=2",
		},
		{
			name:    "empty name skipped",
			cookies: []cookiePair{{Name: "", Value: "x"}, {Name: "b", Value: "2"}},
			want:    "b=2",
		},
		{name: "empty value kept", cookies: []cookiePair{{Name: "a", Value: ""}}, want: "a="},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := cookieHeader(test.cookies); got != test.want {
				t.Fatalf("cookieHeader(%q) = %q, want %q", test.cookies, got, test.want)
			}
		})
	}
}

func TestDownloadHeaders(t *testing.T) {
	headers := downloadHeaders(
		[]cookiePair{{Name: "session", Value: "abc"}},
		"https://example.com/page",
		"Example/1",
	)
	want := map[string]string{
		"Cookie":     "session=abc",
		"Referer":    "https://example.com/page",
		"User-Agent": "Example/1",
	}
	if len(headers) != len(want) {
		t.Fatalf("headers = %#v, want %#v", headers, want)
	}
	for key, value := range want {
		if headers[key] != value {
			t.Errorf("headers[%q] = %q, want %q", key, headers[key], value)
		}
	}

	empty := downloadHeaders(nil, "", "")
	if len(empty) != 0 {
		t.Fatalf("empty headers = %#v, want an empty map", empty)
	}
}

func TestResultFileName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{path: `C:\Users\example\Downloads\file.zip`, want: "file.zip"},
		{path: "/tmp/downloads/file.zip", want: "file.zip"},
		{path: "file.zip", want: "file.zip"},
		{path: "", want: ""},
	}
	for _, test := range tests {
		if got := resultFileName(test.path); got != test.want {
			t.Errorf("resultFileName(%q) = %q, want %q", test.path, got, test.want)
		}
	}
}

func TestBannerScript(t *testing.T) {
	if got := bannerScript(""); got != "" {
		t.Fatalf("bannerScript(\"\") = %q, want an empty string", got)
	}

	message := "Finish the page, then press \"Download\""
	script := bannerScript(message)
	if !strings.Contains(script, "window.top !== window") {
		t.Error("banner script runs in child frames")
	}
	if !strings.Contains(script, "pointer-events:none") {
		t.Error("banner script can block clicks")
	}
	encoded, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}
	if !strings.Contains(script, string(encoded)) {
		t.Fatalf("banner script does not contain the escaped message: %s", script)
	}
}

func TestEventEmitterKeepsFirstError(t *testing.T) {
	emitter := newEventEmitter(failingWriter{})
	if err := emitter.emit(readyEvent()); err == nil {
		t.Fatal("first emit succeeded, want a write error")
	}
	if err := emitter.emit(closedEvent()); err == nil {
		t.Fatal("second emit succeeded, want the first error")
	}
	if emitter.Err() == nil {
		t.Fatal("Err() = nil, want the first error")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func assertEventField(t *testing.T, line, key string, want any) {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal([]byte(line), &value); err != nil {
		t.Fatalf("decode event %q: %v", line, err)
	}
	if got := value[key]; got != want {
		t.Fatalf("%s = %#v, want %#v in %q", key, got, want, line)
	}
}
