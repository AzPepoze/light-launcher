package main

import (
	"encoding/json"
	"net/url"
	"testing"
)

func TestURLOrigin(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{raw: "http://127.0.0.1:4000/#token", want: "http://127.0.0.1:4000"},
		{raw: "https://Example.com/page?q=1", want: "https://example.com"},
		{raw: "https://user:pass@example.com:8443/page", want: "https://example.com:8443"},
	}
	for _, test := range tests {
		target, err := url.Parse(test.raw)
		if err != nil {
			t.Fatalf("parse %q: %v", test.raw, err)
		}
		if got := urlOrigin(target); got != test.want {
			t.Errorf("urlOrigin(%q) = %q, want %q", test.raw, got, test.want)
		}
	}
}

func TestAllowsBridgeOnlyForTheLauncherOrigin(t *testing.T) {
	const launcher = "http://127.0.0.1:4000"
	tests := []struct {
		page string
		want bool
	}{
		{page: "http://127.0.0.1:4000/", want: true},
		{page: "http://127.0.0.1:4000/game/1?tab=files#top", want: true},
		{page: "http://127.0.0.1:4001/", want: false},
		{page: "https://127.0.0.1:4000/", want: false},
		{page: "https://example.com/", want: false},
		{page: "about:blank", want: false},
		{page: "not a url", want: false},
	}
	for _, test := range tests {
		if got := allowsBridge(test.page, launcher); got != test.want {
			t.Errorf("allowsBridge(%q) = %v, want %v", test.page, got, test.want)
		}
	}
	if allowsBridge("http://127.0.0.1:4000/", "") {
		t.Error("an empty launcher origin allowed the bridge")
	}
	if !allowsBridge("http://EXAMPLE.com/", "http://example.com") {
		t.Error("a host with different case did not match the launcher origin")
	}
}

func TestIsExternal(t *testing.T) {
	const launcher = "https://example.com"
	tests := []struct {
		raw  string
		want bool
	}{
		{raw: "https://example.com/page", want: false},
		{raw: "https://sub.example.com/page", want: true},
		{raw: "https://example.com:8443/page", want: true},
		{raw: "http://example.com/page", want: true},
		{raw: "https://example.org/page", want: true},
		{raw: "mailto:person@example.com", want: true},
	}
	for _, test := range tests {
		target, err := url.Parse(test.raw)
		if err != nil {
			t.Fatalf("parse %q: %v", test.raw, err)
		}
		if got := isExternal(target, launcher); got != test.want {
			t.Errorf("isExternal(%q) = %v, want %v", test.raw, got, test.want)
		}
	}
}

func TestIsWebURL(t *testing.T) {
	tests := []struct {
		raw  string
		want bool
	}{
		{raw: "https://example.com/file", want: true},
		{raw: "http://127.0.0.1:4000/", want: true},
		{raw: "blob:https://example.com/id", want: false},
		{raw: "data:text/plain,hello", want: false},
		{raw: "ftp://example.com/file", want: false},
		{raw: "https:///file", want: false},
		{raw: "", want: false},
	}
	for _, test := range tests {
		target, err := url.Parse(test.raw)
		if err != nil {
			t.Fatalf("parse %q: %v", test.raw, err)
		}
		if got := isWebURL(target); got != test.want {
			t.Errorf("isWebURL(%q) = %v, want %v", test.raw, got, test.want)
		}
	}
}

func TestParseBridgeRequest(t *testing.T) {
	req, ok := parseBridgeRequest(`{"id":7,"method":"pickFolder"}`)
	if !ok {
		t.Fatal("valid request was rejected")
	}
	if req.ID != 7 || req.Method != "pickFolder" {
		t.Fatalf("request = %+v", req)
	}
	for _, body := range []string{"", "not json", `{"id":0,"method":"pickFile"}`, `{"method":"pickFile"}`} {
		if _, ok := parseBridgeRequest(body); ok {
			t.Errorf("parseBridgeRequest(%q) succeeded, want an error", body)
		}
	}
}

func TestResolveScript(t *testing.T) {
	script, err := resolveScript(3, "")
	if err != nil {
		t.Fatalf("resolveScript: %v", err)
	}
	if want := "window.llShell._resolve(3, null)"; script != want {
		t.Fatalf("resolveScript(cancel) = %q, want %q", script, want)
	}

	path := `C:\Games\a"b`
	literal, err := json.Marshal(path)
	if err != nil {
		t.Fatalf("marshal path: %v", err)
	}
	script, err = resolveScript(4, path)
	if err != nil {
		t.Fatalf("resolveScript: %v", err)
	}
	want := "window.llShell._resolve(4, " + string(literal) + ")"
	if script != want {
		t.Fatalf("resolveScript(path) = %q, want %q", script, want)
	}
}

func TestDownloadFilename(t *testing.T) {
	tests := []struct {
		suggested string
		rawURL    string
		want      string
	}{
		{suggested: "game.zip", rawURL: "https://example.com/other.zip", want: "game.zip"},
		{suggested: "", rawURL: "https://example.com/files/game.zip", want: "game.zip"},
		{suggested: "", rawURL: "https://example.com/", want: "download"},
		{suggested: "", rawURL: "", want: "download"},
	}
	for _, test := range tests {
		if got := downloadFilename(test.suggested, test.rawURL); got != test.want {
			t.Errorf("downloadFilename(%q, %q) = %q, want %q",
				test.suggested, test.rawURL, got, test.want)
		}
	}
}
