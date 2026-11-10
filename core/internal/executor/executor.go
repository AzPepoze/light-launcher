package executor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// RunGameWithLog starts a command and streams its output lines to onLog.
func RunGameWithLog(commandArguments []string, environment []string, onLog func(string), onExit func()) (*os.Process, error) {
	if len(commandArguments) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	command := exec.Command(commandArguments[0], commandArguments[1:]...)
	command.Env = environment

	stdout, _ := command.StdoutPipe()
	stderr, _ := command.StderrPipe()

	if err := StartCommand(command); err != nil {
		return nil, err
	}

	readPipe := func(reader io.Reader, prefix string) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			onLog(fmt.Sprintf("%s%s", prefix, scanner.Text()))
		}
	}

	go readPipe(stdout, "")
	go readPipe(stderr, "")

	go func() {
		_ = command.Wait()
		ReleaseProcess(command.Process)
		onLog("\n--- Process Exited ---")
		if onExit != nil {
			onExit()
		}
	}()

	return command.Process, nil
}

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
