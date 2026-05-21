package executor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func RunGameWithLog(commandArguments []string, environment []string, onLog func(string), onExit func()) (*os.Process, error) {
	if len(commandArguments) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	command := exec.Command(commandArguments[0], commandArguments[1:]...)
	command.Env = environment
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, _ := command.StdoutPipe()
	stderr, _ := command.StderrPipe()

	if err := command.Start(); err != nil {
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
		onLog("\n--- Process Exited ---")
		if onExit != nil {
			onExit()
		}
	}()

	return command.Process, nil
}

func StopProcessGroup(process *os.Process) error {
	if process == nil {
		return nil
	}
	processGroupId := -process.Pid
	_ = syscall.Kill(processGroupId, syscall.SIGINT)
	time.Sleep(200 * time.Millisecond)
	return syscall.Kill(processGroupId, syscall.SIGTERM)
}
