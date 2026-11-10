package runtimesvc

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/platform"
	"light-launcher/core/internal/types"
)

const instanceName = "light-launcher-instance"

var (
	errNoProtonRuntime = errors.New("No Proton runtime was found. Install a Proton version first.")
	errUmuMissing      = errors.New("umu-run is required to launch games via Proton. Please install umu-launcher (e.g., 'yay -S umu-launcher').")
	errInstanceMissing = errors.New("light-launcher-instance binary not found. Please build or reinstall LightLauncher.")
)

// runGame starts a game through the light-launcher-instance binary.
// On Linux it resolves the prefix and Proton first. On Windows it starts the game directly.
func runGame(options types.LaunchOptions, showLogs bool) error {
	features := platform.Current()
	launch := options
	if !launch.UseGamePath && launch.LauncherPath != "" {
		launch.GamePath = launch.LauncherPath
	}

	if features.Prefix {
		if err := resolvePrefixPath(&launch); err != nil {
			return err
		}
	}
	if !pathExists(launch.GamePath) {
		logger.Error("Runner", "Game executable not found at: %s", launch.GamePath)
		return fmt.Errorf("Game executable not found at: %s", launch.GamePath)
	}
	if features.Proton {
		if err := resolveProtonPath(&launch); err != nil {
			return err
		}
	}

	logger.Info("Runner", "Launching \"%s\" (game: %s, prefix: %s, proton: %s)",
		firstNonEmpty(launch.Name, nameWithoutExt(filepath.Base(launch.GamePath))),
		launch.GamePath, launch.PrefixPath, stringOrDefault(launch.ProtonPath, "default"))

	if err := saveGameConfig(&launch); err != nil {
		return err
	}
	if features.Lsfg {
		if err := syncLsfgProfile(launch); err != nil {
			return err
		}
	}
	if features.Proton {
		if err := requireUmuRun(); err != nil {
			return err
		}
	}

	instancePath, err := findInstanceBinary()
	if err != nil {
		return err
	}
	command, args := setsidCommand(instancePath, buildInstanceArgs(launch, showLogs, features, loadAppSettings()))
	logger.Info("Runner", "Executing: %s %s", command, strings.Join(args, " "))
	if err := startDetached(command, args, nil); err != nil {
		return err
	}
	return recordLaunch(launch)
}

// findInstanceBinary looks next to the running executable first, then in PATH.
func findInstanceBinary() (string, error) {
	name := instanceName
	if platform.OS() == "windows" {
		name += ".exe"
	}
	if executable, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(executable), name)
		if pathExists(candidate) {
			return candidate, nil
		}
	}
	if found, err := exec.LookPath(name); err == nil {
		return found, nil
	}
	logger.Error("Runner", "%s binary not found", instanceName)
	return "", errInstanceMissing
}

func requireUmuRun() error {
	if _, err := exec.LookPath("umu-run"); err != nil {
		logger.Error("Runner", "umu-run command not found in PATH")
		return errUmuMissing
	}
	return nil
}
