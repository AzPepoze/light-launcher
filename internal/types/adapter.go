package types

import "flag"

// LaunchAdapter defines the interface that each tool integration must implement
// to build its command-line arguments and modify execution settings.
type LaunchAdapter interface {
	// Name returns the name of this adapter.
	Name() string

	// BuildInstanceArgs constructs arguments for the light-launcher-instance command line.
	BuildInstanceArgs(options LaunchOptions) []string

	// ModifyEnv appends or modifies environment variables for execution.
	ModifyEnv(options LaunchOptions, env []string) []string

	// WrapCommand wraps the command arguments (e.g. prepending wrappers).
	WrapCommand(options LaunchOptions, args []string) []string

	// RegisterFlags registers the tool's command line flags in the runner.
	RegisterFlags(fs *flag.FlagSet)

	// ExtractOptions populates LaunchOptions with the values parsed from the flags.
	ExtractOptions(options *LaunchOptions)
}
