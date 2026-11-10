package systemsvc

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var errFakeCommand = errors.New("fake command failed")

// fakeRunner answers commands from a table keyed by "name arg1 arg2" and records every argv.
// A command missing from the table fails, like a program that is not installed.
type fakeRunner struct {
	outputs  map[string]string
	failures map[string]bool
	calls    [][]string
}

func newFakeRunner() *fakeRunner {
	return &fakeRunner{outputs: map[string]string{}, failures: map[string]bool{}}
}

func (f *fakeRunner) on(command, output string) *fakeRunner {
	f.outputs[command] = output
	return f
}

func (f *fakeRunner) fail(command string) *fakeRunner {
	f.failures[command] = true
	return f
}

func (f *fakeRunner) Output(_ context.Context, name string, args ...string) ([]byte, error) {
	argv := append([]string{name}, args...)
	f.calls = append(f.calls, argv)
	command := strings.Join(argv, " ")
	output, known := f.outputs[command]
	if f.failures[command] || !known {
		return []byte(output), errFakeCommand
	}
	return []byte(output), nil
}

// callLines lists the recorded commands as "name arg1 arg2" strings.
func (f *fakeRunner) callLines() []string {
	lines := make([]string, 0, len(f.calls))
	for _, argv := range f.calls {
		lines = append(lines, strings.Join(argv, " "))
	}
	return lines
}

// newTestService builds a service over a temp filesystem root and a temp home, so no real path is touched.
func newTestService(t *testing.T, runner Runner) *Service {
	t.Helper()
	return newService(runner, t.TempDir(), t.TempDir())
}

// writeHostFile writes content at an absolute system path inside the test root.
func writeHostFile(t *testing.T, s *Service, absolute, content string) {
	t.Helper()
	path := s.path(absolute)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create dir for %s: %v", absolute, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", absolute, err)
	}
}
