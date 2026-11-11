package library

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDetectExecutableSkipsInstallersAndPrefersShallowNative(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip("launcher suffixes are platform specific; this test covers Linux")
	}
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "uninstall.exe"))
	writeFile(t, filepath.Join(dir, "redist", "vc_redist.x64.exe"))
	writeFile(t, filepath.Join(dir, "bin", "Game.exe"))
	writeFile(t, filepath.Join(dir, "run.sh"))

	if got := DetectExecutable(dir); got != filepath.Join(dir, "run.sh") {
		t.Errorf("DetectExecutable = %q, want run.sh", got)
	}
}

func TestDetectExecutableFindsNothingInEmptyFolder(t *testing.T) {
	if got := DetectExecutable(t.TempDir()); got != "" {
		t.Errorf("DetectExecutable(empty) = %q, want empty", got)
	}
}
