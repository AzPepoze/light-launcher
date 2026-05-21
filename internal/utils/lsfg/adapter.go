package lsfg

import (
	"flag"
	"light-launcher/internal/types"
)

type LsfgAdapter struct {
	enabled    bool
	multiplier string
	perfMode   bool
	dllPath    string
}

func NewAdapter() *LsfgAdapter {
	return &LsfgAdapter{}
}

func (a *LsfgAdapter) Name() string {
	return "lsfg"
}

func (a *LsfgAdapter) BuildInstanceArgs(options types.LaunchOptions) []string {
	var args []string
	if options.Extras.Lsfg.Enabled {
		args = append(args, "--lsfg", "--lsfg-mult", options.Extras.Lsfg.Multiplier)
		if options.Extras.Lsfg.PerfMode {
			args = append(args, "--lsfg-perf")
		}
		if options.Extras.Lsfg.DllPath != "" {
			args = append(args, "--lsfg-dll-path", options.Extras.Lsfg.DllPath)
		}
	}
	return args
}

func (a *LsfgAdapter) ModifyEnv(options types.LaunchOptions, env []string) []string {
	if !options.Extras.Lsfg.Enabled {
		return append(env, "DISABLE_LSFGVK=1")
	}
	return env
}

func (a *LsfgAdapter) WrapCommand(options types.LaunchOptions, args []string) []string {
	return args
}

func (a *LsfgAdapter) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&a.enabled, "lsfg", false, "Enable LSFG-VK")
	fs.StringVar(&a.multiplier, "lsfg-mult", "2", "LSFG Multiplier")
	fs.BoolVar(&a.perfMode, "lsfg-perf", false, "Enable LSFG Performance Mode")
	fs.StringVar(&a.dllPath, "lsfg-dll-path", "", "Path to Lossless.dll")
}

func (a *LsfgAdapter) ExtractOptions(options *types.LaunchOptions) {
	options.Extras.Lsfg.Enabled = a.enabled
	options.Extras.Lsfg.Multiplier = a.multiplier
	options.Extras.Lsfg.PerfMode = a.perfMode
	options.Extras.Lsfg.DllPath = a.dllPath
}
