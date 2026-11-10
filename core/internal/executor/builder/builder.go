package builder

import (
	"fmt"
	"light-launcher/core/internal/adapter"
	"light-launcher/core/internal/config"
	"light-launcher/core/internal/platform"
	"light-launcher/core/internal/types"
	"os"
	"path/filepath"
	"strings"
)

// BuildCommand returns the command arguments and environment used to start the game.
func BuildCommand(options types.LaunchOptions) ([]string, []string) {
	if !platform.Current().Proton {
		return directCommand(options), os.Environ()
	}
	return protonCommand(options)
}

// WorkingDirectory returns the folder to start the game in, or "" to keep the default.
func WorkingDirectory(options types.LaunchOptions) string {
	if platform.Current().Proton {
		return ""
	}
	return filepath.Dir(options.GamePath)
}

// directCommand starts the game executable itself. Used where Proton is not available.
func directCommand(options types.LaunchOptions) []string {
	arguments := []string{options.GamePath}
	return append(arguments, strings.Fields(options.CustomArgs)...)
}

// protonCommand starts the game through umu-run with Proton, then applies the adapters.
func protonCommand(options types.LaunchOptions) ([]string, []string) {
	// 1. Build initial command arguments (base execution command)
	var arguments []string
	arguments = append(arguments, "umu-run")
	executablePath := options.LauncherPath
	if executablePath == "" {
		executablePath = options.GamePath
	}
	arguments = append(arguments, executablePath)

	if options.CustomArgs != "" {
		arguments = append(arguments, strings.Fields(options.CustomArgs)...)
	}

	// 2. Build initial environment
	environment := os.Environ()
	environment = append(environment,
		fmt.Sprintf("WINEPREFIX=%s", config.ExpandPath(options.PrefixPath)),
	)
	if options.ProtonPath != "" {
		protonPattern := filepath.Base(options.ProtonPath)
		environment = append(environment,
			fmt.Sprintf("UMU_PROTON_PATTERN=%s", protonPattern),
			fmt.Sprintf("PROTONPATH=%s", config.ExpandPath(options.ProtonPath)),
		)
	}

	// 3. Apply adapters
	adapters := adapter.GetAdapters()

	// Apply ModifyEnv
	for _, a := range adapters {
		environment = a.ModifyEnv(options, environment)
	}

	// Apply WrapCommand
	for _, a := range adapters {
		arguments = a.WrapCommand(options, arguments)
	}

	return arguments, environment
}
