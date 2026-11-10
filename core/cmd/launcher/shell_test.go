package main

import (
	"context"
	"slices"
	"testing"
)

func TestShellCommandPassesURLInEnvironment(t *testing.T) {
	const url = "http://127.0.0.1:4000/#t=secret-token"
	cmd := shellCommand(context.Background(), "light-launcher-shell", url)

	if len(cmd.Args) != 1 {
		t.Fatalf("shell got arguments %q, want none besides the program name", cmd.Args[1:])
	}
	if !slices.Contains(cmd.Env, shellURLEnv+"="+url) {
		t.Fatalf("%s missing from the shell environment", shellURLEnv)
	}
}

func TestShellCommandKeepsParentEnvironment(t *testing.T) {
	t.Setenv("LL_TEST_KEEP", "yes")
	cmd := shellCommand(context.Background(), "light-launcher-shell", "http://127.0.0.1:1/")

	if !slices.Contains(cmd.Env, "LL_TEST_KEEP=yes") {
		t.Fatal("parent environment was dropped")
	}
}
