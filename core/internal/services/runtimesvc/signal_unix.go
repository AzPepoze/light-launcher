//go:build !windows

package runtimesvc

import "syscall"

// errNoSuchProcess is returned when the process is already gone.
var errNoSuchProcess = syscall.ESRCH

func interruptProcess(pid int) error {
	return syscall.Kill(pid, syscall.SIGINT)
}
