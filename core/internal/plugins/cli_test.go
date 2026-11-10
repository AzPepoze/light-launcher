package plugins

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIInstallListAndRemove(t *testing.T) {
	host := newTestHost(t)
	source := filepath.Join(t.TempDir(), "fixture.llplugin")
	writeFixturePackage(t, source, "0.1.0")
	ctx := context.Background()

	var out bytes.Buffer
	if code := runCommand(ctx, host, []string{"plugin", "install", source}, &out); code != 1 {
		t.Fatalf("unsigned install exit = %d, want 1", code)
	}
	if !strings.Contains(out.String(), "--allow-unsigned") {
		t.Fatalf("unsigned install should explain the flag: %q", out.String())
	}

	out.Reset()
	if code := runCommand(ctx, host, []string{"plugin", "install", source, "--allow-unsigned"}, &out); code != 0 {
		t.Fatalf("install exit = %d, output %q", code, out.String())
	}
	if !strings.Contains(out.String(), "installed fixture 0.1.0") {
		t.Fatalf("install output = %q", out.String())
	}

	out.Reset()
	runCommand(ctx, host, []string{"plugin", "list"}, &out)
	if !strings.Contains(out.String(), "fixture") || !strings.Contains(out.String(), "0.1.0") {
		t.Fatalf("list output = %q", out.String())
	}

	out.Reset()
	if code := runCommand(ctx, host, []string{"plugin", "remove", "fixture"}, &out); code != 0 {
		t.Fatalf("remove exit = %d, output %q", code, out.String())
	}
}

func TestCLIRejectsUnknownCommandsAndFlags(t *testing.T) {
	host := newTestHost(t)
	ctx := context.Background()
	cases := [][]string{
		nil,
		{"launcher", "list"},
		{"plugin", "explode"},
		{"plugin", "install"},
		{"plugin", "list", "--force"},
	}
	for _, args := range cases {
		var out bytes.Buffer
		if code := runCommand(ctx, host, args, &out); code != 2 {
			t.Errorf("runCommand(%v) = %d, want 2", args, code)
		}
		if !strings.Contains(out.String(), "Usage:") {
			t.Errorf("runCommand(%v) did not print usage", args)
		}
	}
}
