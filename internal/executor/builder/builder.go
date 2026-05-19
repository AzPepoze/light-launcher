package builder

import (
	"light-launcher/internal/types"
	"os"
)

type CommandBuilder struct {
	Options     types.LaunchOptions
	Arguments   []string
	Environment []string
}

func NewCommandBuilder(options types.LaunchOptions) *CommandBuilder {
	return &CommandBuilder{
		Options:     options,
		Environment: os.Environ(),
	}
}

func BuildCommand(options types.LaunchOptions) ([]string, []string) {
	builder := NewCommandBuilder(options)

	builder.buildBaseEnvironment()
	builder.applyLsfg()

	if !options.Extras.Gamescope.Enabled {
		builder.applyMangoHud()
	}

	builder.applyGameMode()

	if options.Extras.Gamescope.Enabled {
		builder.applyGamescope()
		builder.applyMangoHud()
	}

	builder.addUmuRun()
	builder.addCustomArgs()
	builder.applyMemoryProtection()

	return builder.Arguments, builder.Environment
}
