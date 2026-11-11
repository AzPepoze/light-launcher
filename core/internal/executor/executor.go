package executor

import (
	"os/exec"
)

// StartCommand starts a command so that StopProcessGroup can stop it and its children.
// Call ReleaseProcess after the command has exited.
func StartCommand(command *exec.Cmd) error {
	PrepareCommand(command)
	if err := command.Start(); err != nil {
		return err
	}
	if err := attachProcess(command.Process); err != nil {
		_ = command.Process.Kill()
		return err
	}
	return nil
}
