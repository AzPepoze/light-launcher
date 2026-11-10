package main

import (
	"path/filepath"
	"testing"
)

func TestParseOptionsDefaults(t *testing.T) {
	opts, err := parseOptions(nil)
	if err != nil {
		t.Fatal(err)
	}
	if opts.shell != "auto" || opts.noShell || opts.uiDir != "" || opts.gamePath != "" {
		t.Fatalf("defaults = %+v", opts)
	}
}

func TestParseOptionsRejectsUnknownShell(t *testing.T) {
	if _, err := parseOptions([]string{"--shell=kde"}); err == nil {
		t.Fatal("unknown shell was accepted")
	}
}

func TestParseOptionsMakesGamePathAbsolute(t *testing.T) {
	opts, err := parseOptions([]string{"--no-shell", "game.exe"})
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.Abs("game.exe")
	if !opts.noShell || opts.gamePath != want {
		t.Fatalf("options = %+v, want game path %q", opts, want)
	}
}

func TestParseOptionsAcceptsIgnoredFlags(t *testing.T) {
	opts, err := parseOptions([]string{"--wayland", "--x11", "--ozone-platform=wayland"})
	if err != nil {
		t.Fatal(err)
	}
	if len(opts.ignored) != 3 {
		t.Fatalf("ignored = %v", opts.ignored)
	}
}

func TestParseOptionsRejectsPortOutOfRange(t *testing.T) {
	if _, err := parseOptions([]string{"--port=70000"}); err == nil {
		t.Fatal("port 70000 was accepted")
	}
}

func TestAPIEndpointUsesFixedValuesOnlyWithDevToken(t *testing.T) {
	port, token := apiEndpoint(options{port: 5000})
	if port != 5000 || token != "" {
		t.Fatalf("no dev token: port %d, token %q", port, token)
	}
	port, token = apiEndpoint(options{devToken: "dev"})
	if port != defaultDevPort || token != "dev" {
		t.Fatalf("dev token without port: port %d, token %q", port, token)
	}
	port, _ = apiEndpoint(options{devToken: "dev", port: 7000})
	if port != 7000 {
		t.Fatalf("explicit port ignored: %d", port)
	}
}
