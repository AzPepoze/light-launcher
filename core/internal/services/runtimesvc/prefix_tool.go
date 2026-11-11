package runtimesvc

import (
	"errors"
	"fmt"
	"os"

	"light-launcher/core/internal/config"
)

// runPrefixTool starts a Windows tool, such as an installer or config app, inside a prefix through umu-run.
func runPrefixTool(prefixPath string, toolName string, protonPath string) error {
	if toolName == "" {
		return errors.New("Tool name cannot be empty")
	}

	resolvedPrefix := config.ExpandPath(prefixPath)
	if !config.PathExists(resolvedPrefix) {
		if err := os.MkdirAll(resolvedPrefix, 0755); err != nil {
			return err
		}
	}

	environment := append(os.Environ(), "WINEPREFIX="+resolvedPrefix)
	if protonPath != "" {
		expandedProton := config.ExpandPath(protonPath)
		environment = append(environment,
			"PROTONPATH="+expandedProton,
			"UMU_PROTON_PATTERN="+config.BaseName(expandedProton),
		)
	}

	command, args := setsidCommand("umu-run", []string{toolName})
	if err := startDetached(command, args, environment); err != nil {
		return fmt.Errorf("Failed to launch prefix tool \"%s\": %s", toolName, err)
	}
	return nil
}
