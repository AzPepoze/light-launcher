package builder

import (
	"fmt"
	"light-launcher/internal/config"
	"light-launcher/internal/types"
	"os"
	"path/filepath"
	"strings"
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

func (builder *CommandBuilder) buildBaseEnvironment() {
	builder.Environment = append(builder.Environment,
		fmt.Sprintf("WINEPREFIX=%s", config.ExpandPath(builder.Options.PrefixPath)),
	)

	if builder.Options.ProtonPath != "" {
		protonPattern := filepath.Base(builder.Options.ProtonPath)
		builder.Environment = append(builder.Environment,
			fmt.Sprintf("UMU_PROTON_PATTERN=%s", protonPattern),
			fmt.Sprintf("PROTONPATH=%s", config.ExpandPath(builder.Options.ProtonPath)),
		)
	}
}

func (builder *CommandBuilder) addUmuRun() {
	builder.Arguments = append(builder.Arguments, "umu-run")

	executablePath := builder.Options.RunnerPath
	if executablePath == "" {
		executablePath = builder.Options.GamePath
	}
	builder.Arguments = append(builder.Arguments, executablePath)
}

func (builder *CommandBuilder) addCustomArgs() {
	if builder.Options.CustomArgs != "" {
		arguments := strings.Fields(builder.Options.CustomArgs)
		builder.Arguments = append(builder.Arguments, arguments...)
	}
}

func FormatCommandForDisplay(commandArguments []string, options types.LaunchOptions) string {
	var builder strings.Builder
	if options.Extras.Memory.Enabled && options.Extras.Memory.Value != "" {
		builder.WriteString(fmt.Sprintf("[MemMin:%s] ", options.Extras.Memory.Value))
	}
	builder.WriteString("WINEPREFIX=" + options.PrefixPath + " ")
	if options.ProtonPath != "" {
		builder.WriteString("UMU_PROTON_PATTERN=" + filepath.Base(options.ProtonPath) + " ")
	}
	if options.Extras.EnableMangoHud {
		builder.WriteString("MANGOHUD=1 ")
	}
	builder.WriteString(strings.Join(commandArguments, " "))
	return builder.String()
}
