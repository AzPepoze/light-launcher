package plugins

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"time"

	"light-launcher/core/internal/executor"
)

// childProcess is one running plugin program with its stdio pipes.
type childProcess struct {
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	stderrDone chan struct{}
}

// startChild launches the program at entry. Each stderr line goes to onStderr.
// The child dies with its process group or Job Object when the launcher stops it.
func startChild(entry, dir string, onStderr func(line string)) (*childProcess, error) {
	cmd := exec.Command(entry)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := executor.StartCommand(cmd); err != nil {
		return nil, err
	}
	child := &childProcess{cmd: cmd, stdin: stdin, stdout: stdout, stderrDone: make(chan struct{})}
	go pumpLines(stderr, onStderr, child.stderrDone)
	return child, nil
}

// pumpLines sends each line of r to sink until r ends, then closes done.
func pumpLines(r io.Reader, sink func(line string), done chan struct{}) {
	defer close(done)
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64<<10), 1<<20)
	for scanner.Scan() {
		sink(scanner.Text())
	}
}

// wait reaps the process. Stdio must be drained first. After grace, the process group is killed.
func (c *childProcess) wait(grace time.Duration) error {
	done := make(chan error, 1)
	go func() { done <- c.cmd.Wait() }()
	defer executor.ReleaseProcess(c.cmd.Process)
	select {
	case err := <-done:
		return err
	case <-time.After(grace):
	}
	c.kill()
	select {
	case err := <-done:
		return err
	case <-time.After(grace):
		_ = c.cmd.Process.Kill()
		return <-done
	}
}

// kill asks the process group to stop, which also stops the plugin's children.
func (c *childProcess) kill() {
	_ = executor.StopProcessGroup(c.cmd.Process)
}
