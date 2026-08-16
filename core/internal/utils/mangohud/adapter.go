package mangohud

import (
	"flag"
	"light-launcher/core/internal/types"
)

type MangoHudAdapter struct {
	enabled bool
}

func NewAdapter() *MangoHudAdapter {
	return &MangoHudAdapter{}
}

func (a *MangoHudAdapter) Name() string {
	return "mangohud"
}

func (a *MangoHudAdapter) BuildInstanceArgs(options types.LaunchOptions) []string {
	if options.Extras.EnableMangoHud {
		return []string{"--mango"}
	}
	return nil
}

func (a *MangoHudAdapter) ModifyEnv(options types.LaunchOptions, env []string) []string {
	if options.Extras.EnableMangoHud {
		return append(env, "MANGOHUD=1")
	}
	return env
}

func (a *MangoHudAdapter) WrapCommand(options types.LaunchOptions, args []string) []string {
	return args
}

func (a *MangoHudAdapter) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&a.enabled, "mango", false, "Enable MangoHud")
	fs.BoolVar(&a.enabled, "mangohud", false, "Enable MangoHud")
}

func (a *MangoHudAdapter) ExtractOptions(options *types.LaunchOptions) {
	options.Extras.EnableMangoHud = a.enabled
}
