package appsvc

import (
	"context"
	"os/exec"
	"sync"
)

type toolCandidate struct {
	name  string
	probe []string
}

// toolResolver finds the first installed tool once and remembers the answer until reset.
type toolResolver struct {
	mu         sync.Mutex
	candidates []toolCandidate
	probed     bool
	tool       string
}

func newToolResolver(candidates []toolCandidate) *toolResolver {
	return &toolResolver{candidates: candidates}
}

// get returns the tool name, or "" when none of the candidates runs.
func (r *toolResolver) get(ctx context.Context) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.probed {
		return r.tool
	}
	r.tool = firstWorkingTool(ctx, r.candidates)
	r.probed = true
	return r.tool
}

func (r *toolResolver) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.probed = false
	r.tool = ""
}

func firstWorkingTool(ctx context.Context, candidates []toolCandidate) string {
	for _, candidate := range candidates {
		if runTool(ctx, candidate.name, candidate.probe...) == nil {
			return candidate.name
		}
	}
	return ""
}

// runTool runs an external program and ignores its output.
func runTool(ctx context.Context, name string, args ...string) error {
	return exec.CommandContext(ctx, name, args...).Run()
}
