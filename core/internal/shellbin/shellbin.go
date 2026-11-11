// Package shellbin locates the native window binary that ships next to the launcher.
package shellbin

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Name is the shell binary name without the platform suffix.
const Name = "light-launcher-shell"

// Path returns the shell binary next to the running launcher.
func Path() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return PathNextTo(exe)
}

// PathNextTo returns the shell binary in the same folder as exe.
func PathNextTo(exe string) (string, error) {
	name := Name
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	path := filepath.Join(filepath.Dir(exe), name)
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("shell binary %s not found next to the launcher", name)
	}
	return path, nil
}
