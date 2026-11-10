//go:build windows

package main

import "os/exec"

// openInBrowser opens a URL in the default browser and returns without waiting for it.
func openInBrowser(target string) error {
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait()
	return nil
}
