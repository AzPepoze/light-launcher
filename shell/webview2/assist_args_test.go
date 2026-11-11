package main

import (
	"testing"

	"light-launcher/shell/common/assist"
)

func TestAssistRequested(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "assist first", args: []string{"--assist", "https://example.com"}, want: true},
		{name: "normal mode", args: []string{"--transparent", "https://example.com"}, want: false},
		{name: "assist later", args: []string{"https://example.com", "--assist"}, want: false},
		{name: "no arguments", args: nil, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := assistRequested(test.args); got != test.want {
				t.Fatalf("assistRequested(%q) = %v, want %v", test.args, got, test.want)
			}
		})
	}
}

func TestParseAssistOptions(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantURL     string
		wantProfile string
		wantTitle   string
		wantMessage string
	}{
		{
			name:        "URL before flags",
			args:        []string{"--assist", "https://example.com/page", "--profile", `C:\profiles\example`},
			wantURL:     "https://example.com/page",
			wantProfile: `C:\profiles\example`,
			wantTitle:   assist.Title,
		},
		{
			name:        "flags before URL",
			args:        []string{"--assist", "--profile", `C:\profiles\example`, "http://example.com/download"},
			wantURL:     "http://example.com/download",
			wantProfile: `C:\profiles\example`,
			wantTitle:   assist.Title,
		},
		{
			name:        "all options",
			args:        []string{"--assist", "--profile", "/tmp/profile", "--title", "Example", "--message", "Finish the page", "https://example.com"},
			wantURL:     "https://example.com",
			wantProfile: "/tmp/profile",
			wantTitle:   "Example",
			wantMessage: "Finish the page",
		},
		{
			name:        "empty title falls back",
			args:        []string{"--assist", "https://example.com", "--profile", "/tmp/profile", "--title", ""},
			wantURL:     "https://example.com",
			wantProfile: "/tmp/profile",
			wantTitle:   assist.Title,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseAssistOptions(test.args)
			if err != nil {
				t.Fatalf("parseAssistOptions: %v", err)
			}
			if got.url.String() != test.wantURL {
				t.Errorf("url = %q, want %q", got.url, test.wantURL)
			}
			if got.profile != test.wantProfile {
				t.Errorf("profile = %q, want %q", got.profile, test.wantProfile)
			}
			if got.title != test.wantTitle {
				t.Errorf("title = %q, want %q", got.title, test.wantTitle)
			}
			if got.message != test.wantMessage {
				t.Errorf("message = %q, want %q", got.message, test.wantMessage)
			}
		})
	}
}

func TestParseAssistOptionsRejectsBadArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "no arguments"},
		{name: "mode without URL", args: []string{"--assist"}},
		{name: "mode not first", args: []string{"https://example.com", "--assist", "--profile", "/tmp/p"}},
		{name: "missing profile", args: []string{"--assist", "https://example.com"}},
		{name: "missing profile value", args: []string{"--assist", "https://example.com", "--profile"}},
		{name: "unknown option", args: []string{"--assist", "https://example.com", "--profile", "/tmp/p", "--stealth"}},
		{name: "two URLs", args: []string{"--assist", "https://example.com", "https://example.org", "--profile", "/tmp/p"}},
		{name: "no host", args: []string{"--assist", "https:///page", "--profile", "/tmp/p"}},
		{name: "wrong scheme", args: []string{"--assist", "ftp://example.com/file", "--profile", "/tmp/p"}},
		{name: "not a URL", args: []string{"--assist", "http://[::1", "--profile", "/tmp/p"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := parseAssistOptions(test.args); err == nil {
				t.Fatalf("parseAssistOptions(%q) succeeded, want an error", test.args)
			}
		})
	}
}
