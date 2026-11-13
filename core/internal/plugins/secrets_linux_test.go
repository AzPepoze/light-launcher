//go:build linux

package plugins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeSecretTool puts a script named secret-tool first on PATH. It logs each call's arguments,
// then prints message to stderr when it is set and exits with code.
func fakeSecretTool(t *testing.T, code int, message string) string {
	t.Helper()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "calls.log")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> '" + logPath + "'\n"
	if message != "" {
		script += "printf '%s\\n' '" + message + "' >&2\n"
	}
	script += fmt.Sprintf("exit %d\n", code)
	if err := os.WriteFile(filepath.Join(dir, "secret-tool"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return logPath
}

// wantSecretToolCall fails unless the fake secret-tool ran exactly once with these arguments.
func wantSecretToolCall(t *testing.T, logPath, want string) {
	t.Helper()
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read secret-tool calls: %v", err)
	}
	if got := strings.TrimSpace(string(data)); got != want {
		t.Fatalf("secret-tool calls = %q, want %q", got, want)
	}
}

func TestKeyringDeleteAllClearsOnlyThatPlugin(t *testing.T) {
	logPath := fakeSecretTool(t, 0, "")
	if err := (keyringSecrets{}).DeleteAll(context.Background(), "demo"); err != nil {
		t.Fatalf("DeleteAll: %v", err)
	}
	wantSecretToolCall(t, logPath, "clear service light-launcher plugin demo")
}

func TestKeyringDeleteAllWithoutMatchesIsNotAnError(t *testing.T) {
	fakeSecretTool(t, 1, "")
	if err := (keyringSecrets{}).DeleteAll(context.Background(), "demo"); err != nil {
		t.Fatalf("DeleteAll with no match: %v", err)
	}
}

func TestKeyringDeleteAllReportsFailures(t *testing.T) {
	fakeSecretTool(t, 2, "collection is locked")
	err := (keyringSecrets{}).DeleteAll(context.Background(), "demo")
	if err == nil || !strings.Contains(err.Error(), "collection is locked") {
		t.Fatalf("DeleteAll error = %v, want it to report the secret-tool message", err)
	}
}

func TestKeyringDeleteWithoutMatchesIsNotAnError(t *testing.T) {
	fakeSecretTool(t, 1, "")
	if err := (keyringSecrets{}).Delete(context.Background(), "demo", "token"); err != nil {
		t.Fatalf("Delete with no match: %v", err)
	}
}
