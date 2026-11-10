package systemsvc

import (
	"context"
	"os/exec"
)

// Runner runs external programs. Tests replace it with a fake so no real command runs.
type Runner interface {
	Output(ctx context.Context, name string, args ...string) ([]byte, error)
}

type execRunner struct{}

func (execRunner) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).Output()
}
