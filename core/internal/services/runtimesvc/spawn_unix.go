//go:build !windows

package runtimesvc

import (
	"os/exec"
	"syscall"
)

// startDetached starts a command in a new session with no terminal. It does not wait for it to exit.
func startDetached(name string, args []string, env []string) error {
	command := exec.Command(name, args...)
	command.Env = env
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := command.Start(); err != nil {
		return err
	}
	go func() { _ = command.Wait() }()
	return nil
}
