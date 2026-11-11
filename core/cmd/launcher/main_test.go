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

func TestParseOptionsAcceptsGTKShell(t *testing.T) {
	opts, err := parseOptions([]string{"--shell=gtk"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.shell != "gtk" {
		t.Fatalf("shell = %q, want gtk", opts.shell)
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

func TestDevUIURLPointsAtViteOnlyWithADevToken(t *testing.T) {
	if got := devUIURL(""); got != "" {
		t.Fatalf("devUIURL with no token = %q, want empty", got)
	}
	want := "http://localhost:9245/#t=dev"
	if got := devUIURL("dev"); got != want {
		t.Fatalf("devUIURL = %q, want %q", got, want)
	}
}
