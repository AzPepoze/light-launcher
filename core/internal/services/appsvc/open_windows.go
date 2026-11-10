package appsvc

import (
	"os/exec"
	"syscall"
)

// openExternalURL opens a link in the default browser through Explorer.
func openExternalURL(url string) error {
	return startDetached(exec.Command("explorer.exe", url))
}

// openPathInShell opens a folder in Explorer.
func openPathInShell(path string) error {
	return startDetached(exec.Command("explorer.exe", path))
}

// revealInFolder selects the file in Explorer. Explorer needs the quoted path on the raw command line.
func revealInFolder(path string) error {
	cmd := exec.Command("explorer.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: `explorer.exe /select,"` + path + `"`}
	return startDetached(cmd)
}
