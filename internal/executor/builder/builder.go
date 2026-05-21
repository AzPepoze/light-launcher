package builder

import (
	"fmt"
	"light-launcher/internal/adapter"
	"light-launcher/internal/config"
	"light-launcher/internal/types"
	"os"
	"path/filepath"
	"strings"
)

func BuildCommand(options types.LaunchOptions) ([]string, []string) {
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
