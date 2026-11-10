//go:build windows

package runtimesvc

import "os"

// errNoSuchProcess is returned when the process is already gone.
var errNoSuchProcess = os.ErrProcessDone

// interruptProcess has no SIGINT on Windows, so it stops the process.
func interruptProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Kill()
}
