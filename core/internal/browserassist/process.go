package browserassist

import (
	"io"
	"os"
	"os/exec"
)

// Process is one running shell, as the assist service sees it.
type Process interface {
	Stdout() io.Reader
	Kill() error
	Wait() error
}

// Starter starts the shell binary. Tests replace it with a fake.
type Starter interface {
	Start(bin string, args []string) (Process, error)
}

type execStarter struct{}

func (execStarter) Start(bin string, args []string) (Process, error) {
	cmd := exec.Command(bin, args...)
	cmd.Stderr = os.Stderr
	configureProcess(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &execProcess{cmd: cmd, stdout: stdout}, nil
}

type execProcess struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
}

func (p *execProcess) Stdout() io.Reader { return p.stdout }

func (p *execProcess) Kill() error { return killProcess(p.cmd) }

func (p *execProcess) Wait() error { return p.cmd.Wait() }
