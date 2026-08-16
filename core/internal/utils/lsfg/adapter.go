package lsfg

import (
	"flag"
	"light-launcher/core/internal/types"
)

type LsfgAdapter struct {
	enabled    bool
	multiplier string
	perfMode   bool
	dllPath    string
	gpu        string
	flowScale  string
	pacing     string
	allowFp16  bool
}

func NewAdapter() *LsfgAdapter {
	return &LsfgAdapter{
		multiplier: "2",
		flowScale:  "1.0",
		pacing:     "smooth",
	}
}

func (a *LsfgAdapter) Name() string {
	return "lsfg"
}

func (a *LsfgAdapter) BuildInstanceArgs(options types.LaunchOptions) []string {
	var args []string
	if options.Extras.Lsfg.Enabled {
		args = append(args, "--lsfg", "--lsfg-multiplier", options.Extras.Lsfg.Multiplier)
		if options.Extras.Lsfg.PerfMode {
			args = append(args, "--lsfg-perf")
		}
		if options.Extras.Lsfg.DllPath != "" {
			args = append(args, "--lsfg-dll", options.Extras.Lsfg.DllPath)
		}
		if options.Extras.Lsfg.Gpu != "" {
			args = append(args, "--lsfg-gpu", options.Extras.Lsfg.Gpu)
		}
		if options.Extras.Lsfg.FlowScale != "" {
			args = append(args, "--lsfg-flow", options.Extras.Lsfg.FlowScale)
		}
		if options.Extras.Lsfg.Pacing != "" {
			args = append(args, "--lsfg-pacing", options.Extras.Lsfg.Pacing)
		}
		if options.Extras.Lsfg.AllowFp16 {
			args = append(args, "--lsfg-fp16")
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
	fs.StringVar(&a.multiplier, "lsfg-multiplier", "2", "LSFG Multiplier")
	fs.StringVar(&a.multiplier, "lsfg-mult", "2", "LSFG Multiplier alias")
	fs.BoolVar(&a.perfMode, "lsfg-perf", false, "Enable LSFG Performance Mode")
	fs.StringVar(&a.dllPath, "lsfg-dll", "", "Path to Lossless.dll")
	fs.StringVar(&a.dllPath, "lsfg-dll-path", "", "Path to Lossless.dll alias")
	fs.StringVar(&a.gpu, "lsfg-gpu", "", "GPU index")
	fs.StringVar(&a.flowScale, "lsfg-flow", "1.0", "Flow scale")
	fs.StringVar(&a.pacing, "lsfg-pacing", "smooth", "Pacing method")
	fs.BoolVar(&a.allowFp16, "lsfg-fp16", false, "Allow FP16")
}

func (a *LsfgAdapter) ExtractOptions(options *types.LaunchOptions) {
	options.Extras.Lsfg.Enabled = a.enabled
	options.Extras.Lsfg.Multiplier = a.multiplier
	options.Extras.Lsfg.PerfMode = a.perfMode
	options.Extras.Lsfg.DllPath = a.dllPath
	options.Extras.Lsfg.Gpu = a.gpu
	options.Extras.Lsfg.FlowScale = a.flowScale
	options.Extras.Lsfg.Pacing = a.pacing
	options.Extras.Lsfg.AllowFp16 = a.allowFp16
}
