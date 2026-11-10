// Package systemsvc serves the system methods: tool checks, host info, usage and maintenance.
package systemsvc

import (
	"context"
	"os"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/platform"
)

// Service holds the state the system methods share between calls.
type Service struct {
	runner Runner
	root   string // filesystem root: empty in production, a temp dir in tests
	home   string
	cpu    cpuSampler
	gpus   gpuCache
}

// PlatformInfo tells the UI which OS it runs on and which parts it should show.
type PlatformInfo struct {
	OS       string            `json:"os"`
	Features platform.Features `json:"features"`
}

func newService(runner Runner, root, home string) *Service {
	return &Service{runner: runner, root: root, home: home}
}

// Register adds the system methods to r. The events hook is reserved for progress pushes.
func Register(r *api.Router, _ api.Events) {
	home, _ := os.UserHomeDir()
	newService(execRunner{}, "", home).register(r)
}

func (s *Service) register(r *api.Router) {
	r.Handle("GetSystemToolsStatus", returning(systemToolsStatus))
	r.Handle("GetSystemInfo", returning(s.systemInfo))
	r.Handle("GetSystemUsage", returning(s.systemUsage))
	r.Handle("GetShaderCacheSize", returning(s.shaderCacheSize))
	r.Handle("ClearShaderCache", withoutResult(s.clearShaderCache))
	r.Handle("DropCaches", withoutResult(s.dropCaches))
	r.Handle("ClearSwap", withoutResult(s.clearSwap))
	r.Handle("CleanupProcesses", withoutResult(s.cleanupProcesses))
	r.Handle("GetListGpus", returning(s.listGpus))
	r.Handle("GetUtilsStatus", returning(s.utilsStatus))
	r.Handle("GetPlatform", returning(platformInfo))
}

func platformInfo(_ context.Context) PlatformInfo {
	return PlatformInfo{OS: platform.OS(), Features: platform.Current()}
}

// returning wraps a method that cannot fail.
func returning[Out any](fn func(context.Context) Out) api.Handler {
	return api.NoInput(func(ctx context.Context) (Out, error) {
		return fn(ctx), nil
	})
}

// withoutResult wraps a method that only reports success or failure. The UI gets null.
func withoutResult(fn func(context.Context) error) api.Handler {
	return api.NoInput(func(ctx context.Context) (any, error) {
		return nil, fn(ctx)
	})
}
