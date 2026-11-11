package shellbin

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPathNextToFindsTheShell(t *testing.T) {
	dir := t.TempDir()
	name := Name
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	want := filepath.Join(dir, name)
	if err := os.WriteFile(want, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := PathNextTo(filepath.Join(dir, "light-launcher"))
	if err != nil {
		t.Fatalf("PathNextTo: %v", err)
	}
	if got != want {
		t.Fatalf("PathNextTo = %q, want %q", got, want)
	}
}

func TestPathNextToRejectsAMissingShell(t *testing.T) {
	if _, err := PathNextTo(filepath.Join(t.TempDir(), "light-launcher")); err == nil {
		t.Fatal("missing shell was accepted")
	}
}
