//go:build !windows

package executor

import (
	"os"
	"os/exec"
	"syscall"
	"time"
)

// PrepareCommand puts the command in its own process group.
func PrepareCommand(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func attachProcess(process *os.Process) error {
	return nil
}

// ReleaseProcess frees what StartCommand kept for the process. Nothing is kept on Unix.
func ReleaseProcess(process *os.Process) {}

// StopProcessGroup sends SIGINT, then SIGTERM, to the process group.
func StopProcessGroup(process *os.Process) error {
	if process == nil {
		return nil
	}
	processGroupId := -process.Pid
	_ = syscall.Kill(processGroupId, syscall.SIGINT)
	time.Sleep(200 * time.Millisecond)
	return syscall.Kill(processGroupId, syscall.SIGTERM)
}
