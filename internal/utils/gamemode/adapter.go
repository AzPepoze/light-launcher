package gamemode

import (
	"flag"
	"light-launcher/internal/system"
	"light-launcher/internal/types"
)

type GameModeAdapter struct {
	enabled bool
}

func NewAdapter() *GameModeAdapter {
	return &GameModeAdapter{}
}

func (a *GameModeAdapter) Name() string {
	return "gamemode"
}

func (a *GameModeAdapter) BuildInstanceArgs(options types.LaunchOptions) []string {
	if options.Extras.EnableGamemode {
		return []string{"--gamemode"}
	}
	return nil
}

func (a *GameModeAdapter) ModifyEnv(options types.LaunchOptions, env []string) []string {
	return env
}

func (a *GameModeAdapter) WrapCommand(options types.LaunchOptions, args []string) []string {
	if options.Extras.EnableGamemode && system.IsCommandAvailable("gamemoderun") {
		return append([]string{"gamemoderun"}, args...)
	}
	return args
}

func (a *GameModeAdapter) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&a.enabled, "gamemode", false, "Enable GameMode")
}

func (a *GameModeAdapter) ExtractOptions(options *types.LaunchOptions) {
	options.Extras.EnableGamemode = a.enabled
}
