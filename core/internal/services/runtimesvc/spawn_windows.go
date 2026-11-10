//go:build windows

package runtimesvc

import (
	"os/exec"
	"syscall"
)

// detachedProcess is DETACHED_PROCESS from the Windows API. Package syscall does not export it.
const detachedProcess = 0x00000008

// startDetached starts a command as a detached process. It does not wait for it to exit.
func startDetached(name string, args []string, env []string) error {
	command := exec.Command(name, args...)
	command.Env = env
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | detachedProcess}
	if err := command.Start(); err != nil {
		return err
	}
	go func() { _ = command.Wait() }()
	return nil
}
