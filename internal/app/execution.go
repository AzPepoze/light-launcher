package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"light-launcher/internal/app/instancebuilder"
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

	protonTools, _ := system.GetProtonTools()
	if match := app.findProtonMatch(options.ProtonPath, protonTools); match != nil {
		options.ProtonPath = match.Path
	}

	arguments := instancebuilder.NewBuilder(options, showLogs).Build()
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



func parseMultiplier(multiplier string) int {
	value := 2
	if multiplier != "" {
		if _, err := fmt.Sscanf(multiplier, "%d", &value); err != nil {
			value = 2
		}
	}
	return value
}
