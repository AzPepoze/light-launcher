package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"light-launcher/internal/config"
	"light-launcher/internal/executor"
	"light-launcher/internal/system"
	"light-launcher/internal/types"
	"light-launcher/lib/lsfg"
)

func (app *App) RunGame(options types.LaunchOptions, showLogs bool) error {
	executor.DebugLog("RunGame called with options for: " + options.GamePath)

	if !options.UseGamePath && options.LauncherPath != "" {
		options.GamePath = options.LauncherPath
	}

	if options.PrefixPath == "" {
		options.PrefixPath = filepath.Join(config.GetPrefixBaseDirectory(), "Default")
	}

	if _, err := os.Stat(options.GamePath); os.IsNotExist(err) {
		return fmt.Errorf("game executable not found at: %s", options.GamePath)
	}

	_ = config.SaveGameConfig(options)

	if options.Extras.Lsfg.Enabled {
		configPath, err := lsfg.GetConfigPath()
		if err == nil {
			// Always update/save profile during launch if enabled
			gpu := options.Extras.Lsfg.Gpu
			if gpu == "" {
				gpuList := system.GetListGpus()
				if len(gpuList) > 0 {
					gpu = gpuList[0]
				}
			}

			_ = lsfg.SaveProfileToPath(options.Name, options.GamePath, configPath,
				parseMultiplier(options.Extras.Lsfg.Multiplier),
				options.Extras.Lsfg.PerfMode,
				options.Extras.Lsfg.DllPath,
				gpu,
				options.Extras.Lsfg.FlowScale,
				options.Extras.Lsfg.Pacing,
				options.Extras.Lsfg.AllowFp16)
		}
	} else {
		_ = lsfg.DisableProfileInConfig(options.Name, options.GamePath)
	}

	instanceManagerPath := findInstanceManager()
	if instanceManagerPath == "" {
		return fmt.Errorf("instance manager not found")
	}

	arguments := buildInstanceManagerArgs(options, showLogs)
	command := exec.Command(instanceManagerPath, arguments...)
	if err := command.Start(); err != nil {
		return fmt.Errorf("failed to start instance manager: %w", err)
	}
	go command.Process.Release()

	return nil
}

func findInstanceManager() string {
	instanceName := "light-launcher-instance"
	executablePath, err := os.Executable()
	if err != nil {
		return ""
	}
	executableDirectory := filepath.Dir(executablePath)

	potentialPaths := []string{
		filepath.Join(executableDirectory, instanceName),
		"./bin/" + instanceName,
		filepath.Join(executableDirectory, "../bin", instanceName),
		"../bin/" + instanceName,
		"./" + instanceName,
		"/usr/bin/" + instanceName,
	}

	for _, path := range potentialPaths {
		if _, err := os.Stat(path); err == nil {
			if absolutePath, err := filepath.Abs(path); err == nil {
				return absolutePath
			}
			return path
		}
	}
	return ""
}

func buildInstanceManagerArgs(options types.LaunchOptions, showLogs bool) []string {
	arguments := []string{
		"--game", options.GamePath,
		"--launcher", options.LauncherPath,
		"--prefix", options.PrefixPath,
		"--proton-pattern", filepath.Base(options.ProtonPath),
		"--proton-path", options.ProtonPath,
	}
	if options.Extras.EnableMangoHud {
		arguments = append(arguments, "--mango")
	}
	if options.Extras.EnableGamemode {
		arguments = append(arguments, "--gamemode")
	}
	if options.Extras.Lsfg.Enabled {
		arguments = append(arguments, "--lsfg", "--lsfg-mult", options.Extras.Lsfg.Multiplier)
		if options.Extras.Lsfg.PerfMode {
			arguments = append(arguments, "--lsfg-perf")
		}
		if options.Extras.Lsfg.DllPath != "" {
			arguments = append(arguments, "--lsfg-dll-path", options.Extras.Lsfg.DllPath)
		}
	}
	if options.Extras.Memory.Enabled {
		arguments = append(arguments, "--memory-min")
		if options.Extras.Memory.Value != "" {
			arguments = append(arguments, "--memory-min-value", options.Extras.Memory.Value)
		}
	}
	if options.Extras.Gamescope.Enabled {
		arguments = append(arguments, "--gamescope",
			"--gs-w", options.Extras.Gamescope.Width,
			"--gs-h", options.Extras.Gamescope.Height,
			"--gs-r", options.Extras.Gamescope.RefreshRate)
	}
	if !showLogs {
		arguments = append(arguments, "--logs=false")
	}
	return arguments
}

func parseMultiplier(multiplier string) int {
	value := 2
	if multiplier != "" {
		if _, err := fmt.Sscanf(multiplier, "%d", &value); err != nil {
			value = 2
		}
	}
	return value
}
