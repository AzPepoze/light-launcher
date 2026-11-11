package urlx

import (
	"net/url"
	"strings"
	"testing"
)

func TestShellURLPrefersTheEnvironment(t *testing.T) {
	got, err := ShellURL("http://127.0.0.1:1/#token", []string{"https://example.com"})
	if err != nil {
		t.Fatalf("ShellURL: %v", err)
	}
	if want := "http://127.0.0.1:1/#token"; got != want {
		t.Fatalf("ShellURL = %q, want %q", got, want)
	}
}

func TestShellURLUsesTheArgumentWhenTheEnvironmentIsEmpty(t *testing.T) {
	got, err := ShellURL("", []string{"https://example.com/page"})
	if err != nil {
		t.Fatalf("ShellURL: %v", err)
	}
	if want := "https://example.com/page"; got != want {
		t.Fatalf("ShellURL = %q, want %q", got, want)
	}
}

func TestShellURLRejectsMissingOrExtraArguments(t *testing.T) {
	for _, args := range [][]string{nil, {"https://example.com", "https://example.org"}} {
		if _, err := ShellURL("", args); err == nil {
			t.Errorf("ShellURL(%q) succeeded, want an error", args)
		}
	}
}

func TestParseTarget(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{raw: "https://example.com/page?q=1", want: "https://example.com/page?q=1"},
		{raw: "http://127.0.0.1:8080/file", want: "http://127.0.0.1:8080/file"},
	}
	for _, test := range tests {
		target, err := ParseTarget(test.raw)
		if err != nil {
			t.Fatalf("ParseTarget(%q): %v", test.raw, err)
		}
		if got := target.String(); got != test.want {
			t.Errorf("ParseTarget(%q) = %q, want %q", test.raw, got, test.want)
		}
	}
}

func TestParseTargetRejectsBadURL(t *testing.T) {
	for _, raw := range []string{"", "ftp://example.com", "https:///page", "not a url", "https://example.com/%zz"} {
		if _, err := ParseTarget(raw); err == nil {
			t.Errorf("ParseTarget(%q) succeeded, want an error", raw)
		}
	}
	const secret = "https://example.com/%zz?token=private"
	if _, err := ParseTarget(secret); err == nil || strings.Contains(err.Error(), "token=private") {
		t.Fatalf("error leaks the input: %v", err)
	}
}

func TestOrigin(t *testing.T) {
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
		if got := Origin(target); got != test.want {
			t.Errorf("Origin(%q) = %q, want %q", test.raw, got, test.want)
		}
	}
	if got := Origin(nil); got != "" {
		t.Errorf("Origin(nil) = %q, want empty", got)
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
		if got := IsExternal(target, launcher); got != test.want {
			t.Errorf("IsExternal(%q) = %v, want %v", test.raw, got, test.want)
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
		if got := IsWebURL(target); got != test.want {
			t.Errorf("IsWebURL(%q) = %v, want %v", test.raw, got, test.want)
		}
	}
}

func TestIsWebAddress(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{url: "https://example.com/file", want: true},
		{url: "http://example.com/file", want: true},
		{url: "blob:https://example.com/id", want: false},
		{url: "data:text/plain,hello", want: false},
		{url: "ftp://example.com/file", want: false},
		{url: "https:///file", want: false},
		{url: "not a url", want: false},
	}
	for _, test := range tests {
		if got := IsWebAddress(test.url); got != test.want {
			t.Errorf("IsWebAddress(%q) = %v, want %v", test.url, got, test.want)
		}
	}
}
