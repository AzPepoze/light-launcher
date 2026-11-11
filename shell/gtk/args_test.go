//go:build linux

package main

import (
	"strings"
	"testing"

	"light-launcher/shell/common/assist"
)

func TestParseOptionsKeepsNormalMode(t *testing.T) {
	t.Setenv("LL_URL", "")
	opts, err := parseOptions([]string{"--transparent", "https://example.com/ui"})
	if err != nil {
		t.Fatalf("parseOptions: %v", err)
	}
	if !opts.transparent {
		t.Error("transparent = false, want true")
	}
	if got, want := opts.url.String(), "https://example.com/ui"; got != want {
		t.Errorf("url = %q, want %q", got, want)
	}
}

func TestParseOptionsPrefersTheEnvironment(t *testing.T) {
	t.Setenv("LL_URL", "http://127.0.0.1:1/#token")
	opts, err := parseOptions([]string{"https://example.com"})
	if err != nil {
		t.Fatalf("parseOptions: %v", err)
	}
	if got, want := opts.url.String(), "http://127.0.0.1:1/#token"; got != want {
		t.Errorf("url = %q, want %q", got, want)
	}
}

func TestParseOptionsRejectsBadURLWithoutLeakingIt(t *testing.T) {
	t.Setenv("LL_URL", "")
	const secret = "https://example.com/%zz?token=private"
	_, err := parseOptions([]string{secret})
	if err == nil {
		t.Fatal("expected an error")
	}
	if strings.Contains(err.Error(), "token=private") {
		t.Fatalf("error leaks the input: %v", err)
	}
}

func TestWantsAssist(t *testing.T) {
	tests := []struct {
		args []string
		want bool
	}{
		{args: []string{"--assist", "https://example.com"}, want: true},
		{args: []string{"--assist=https://example.com"}, want: true},
		{args: []string{"-assist", "https://example.com"}, want: true},
		{args: []string{"--profile", "/tmp/profile", "--assist", "https://example.com"}, want: true},
		{args: []string{"--transparent", "https://example.com"}, want: false},
		{args: []string{"https://example.com"}, want: false},
		{args: []string{"--assistance", "https://example.com"}, want: false},
	}
	for _, test := range tests {
		if got := wantsAssist(test.args); got != test.want {
			t.Errorf("wantsAssist(%v) = %v, want %v", test.args, got, test.want)
		}
	}
}

func TestParseAssistOptions(t *testing.T) {
	opts, err := parseAssistOptions([]string{
		"--assist", "https://example.com/page",
		"--profile", "/tmp/profile",
		"--title", "Sign in",
		"--message", "Finish the page",
	})
	if err != nil {
		t.Fatalf("parseAssistOptions: %v", err)
	}
	if got, want := opts.target.String(), "https://example.com/page"; got != want {
		t.Errorf("target = %q, want %q", got, want)
	}
	if got, want := opts.profile, "/tmp/profile"; got != want {
		t.Errorf("profile = %q, want %q", got, want)
	}
	if got, want := opts.title, "Sign in"; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}
	if got, want := opts.message, "Finish the page"; got != want {
		t.Errorf("message = %q, want %q", got, want)
	}
}

func TestParseAssistOptionsDefaults(t *testing.T) {
	opts, err := parseAssistOptions([]string{"--assist=https://example.com", "--profile", "/tmp/profile"})
	if err != nil {
		t.Fatalf("parseAssistOptions: %v", err)
	}
	if got, want := opts.title, assist.Title; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}
	if opts.message != "" {
		t.Errorf("message = %q, want empty", opts.message)
	}
}

func TestParseAssistOptionsRejectsBadInput(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing URL", args: []string{"--profile", "/tmp/profile"}},
		{name: "missing profile", args: []string{"--assist", "https://example.com"}},
		{name: "extra argument", args: []string{"--assist", "https://example.com", "--profile", "/tmp/profile", "extra"}},
		{name: "wrong scheme", args: []string{"--assist", "ftp://example.com", "--profile", "/tmp/profile"}},
		{name: "no host", args: []string{"--assist", "https:///page", "--profile", "/tmp/profile"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := parseAssistOptions(test.args); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
