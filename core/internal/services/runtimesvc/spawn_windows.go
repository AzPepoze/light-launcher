//go:build windows

package runtimesvc

import (
	"os/exec"
	"syscall"
)

// startDetached starts a command as a detached process. It does not wait for it to exit.
func startDetached(name string, args []string, env []string) error {
	command := exec.Command(name, args...)
	command.Env = env
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | syscall.DETACHED_PROCESS}
	if err := command.Start(); err != nil {
		return err
	}
	go func() { _ = command.Wait() }()
	return nil
}
