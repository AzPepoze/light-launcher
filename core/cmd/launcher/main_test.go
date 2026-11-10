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
