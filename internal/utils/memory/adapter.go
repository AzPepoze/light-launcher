package memory

import (
	"flag"
	"fmt"
	"light-launcher/internal/system"
	"light-launcher/internal/types"
)

type MemoryAdapter struct {
	enabled bool
	value   string
}

func NewAdapter() *MemoryAdapter {
	return &MemoryAdapter{}
}

func (a *MemoryAdapter) Name() string {
	return "memory"
}

func (a *MemoryAdapter) BuildInstanceArgs(options types.LaunchOptions) []string {
	var args []string
	if options.Extras.Memory.Enabled {
		args = append(args, "--memory-min")
		if options.Extras.Memory.Value != "" {
			args = append(args, "--memory-min-value", options.Extras.Memory.Value)
		}
	}
	return args
}

func (a *MemoryAdapter) ModifyEnv(options types.LaunchOptions, env []string) []string {
	return env
}

func (a *MemoryAdapter) WrapCommand(options types.LaunchOptions, args []string) []string {
	if options.Extras.Memory.Enabled && options.Extras.Memory.Value != "" && system.IsCommandAvailable("systemd-run") {
		wrapped := []string{
			"systemd-run",
			"--user",
			"--scope",
			fmt.Sprintf("-pMemoryMin=%s", options.Extras.Memory.Value),
			"--",
		}
		return append(wrapped, args...)
	}
	return args
}

func (a *MemoryAdapter) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&a.enabled, "memory-min", false, "Enable Memory Protection (min RAM)")
	fs.StringVar(&a.value, "memory-min-value", "", "Memory Protection Value (e.g. 4G)")
}

func (a *MemoryAdapter) ExtractOptions(options *types.LaunchOptions) {
	options.Extras.Memory.Enabled = a.enabled
	options.Extras.Memory.Value = a.value
}
