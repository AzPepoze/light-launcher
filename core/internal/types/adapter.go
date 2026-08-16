package types

import "flag"

type LaunchAdapter interface {
	Name() string
	BuildInstanceArgs(options LaunchOptions) []string
	ModifyEnv(options LaunchOptions, env []string) []string
	WrapCommand(options LaunchOptions, args []string) []string
	RegisterFlags(fs *flag.FlagSet)
	ExtractOptions(options *LaunchOptions)
}
