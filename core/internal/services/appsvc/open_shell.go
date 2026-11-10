package appsvc

import "os/exec"

// startDetached starts a program and reaps it in the background, so the UI call never waits on it.
func startDetached(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
