package main

import (
	"os/exec"
	"runtime"
)

// openBrowser opens url in the default browser and returns without waiting for it.
func openBrowser(url string) error {
	cmd := browserCommand(url)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}

func browserCommand(url string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	return exec.Command("xdg-open", url)
}
