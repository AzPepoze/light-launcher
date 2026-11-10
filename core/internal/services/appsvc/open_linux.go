package appsvc

import (
	"os/exec"
	"path/filepath"
)

// openExternalURL opens a link with the desktop's default handler.
func openExternalURL(url string) error {
	return startDetached(exec.Command("xdg-open", url))
}

// openPathInShell opens a folder in the file manager.
func openPathInShell(path string) error {
	return startDetached(exec.Command("xdg-open", path))
}

// revealInFolder opens the folder that holds the file, since xdg-open cannot select the file.
func revealInFolder(path string) error {
	return startDetached(exec.Command("xdg-open", filepath.Dir(path)))
}
