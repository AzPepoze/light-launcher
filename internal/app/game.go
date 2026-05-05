package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"light-launcher/internal/config"
	"light-launcher/internal/executor"
	"light-launcher/internal/executor/builder"
	"light-launcher/internal/system"
	"light-launcher/internal/types"
	"light-launcher/lib/lsfg"
)

func (app *App) RunGame(options types.LaunchOptions, showLogs bool) error {
	executor.DebugLog("RunGame called with options for: " + options.GamePath)

	if _, err := os.Stat(options.GamePath); os.IsNotExist(err) {
		return fmt.Errorf("game executable not found at: %s", options.GamePath)
	}

	if !options.UseGamePath && options.RunnerPath != "" {
		options.GamePath = options.RunnerPath
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
		"--launcher", options.RunnerPath,
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

func (app *App) GetAllGames() ([]types.GameInfo, error) {
	configs, err := config.ListGameConfigs()
	if err != nil {
		return nil, err
	}

	games := make([]types.GameInfo, 0)
	for _, gameConfig := range configs {
		name := gameConfig.Name
		if name == "" {
			name = filepath.Base(gameConfig.GamePath)
			name = strings.TrimSuffix(name, filepath.Ext(name))
		}

		cleanedPath := filepath.Clean(gameConfig.RunnerPath)
		if cleanedPath == "" {
			cleanedPath = filepath.Clean(gameConfig.GamePath)
		}
		if absolutePath, err := filepath.Abs(cleanedPath); err == nil {
			cleanedPath = absolutePath
		}

		games = append(games, types.GameInfo{
			Name:   name,
			Path:   cleanedPath,
			Config: gameConfig,
		})
	}
	return games, nil
}

func (app *App) GetRunningSessions() ([]types.RunningSession, error) {
	output, _ := exec.Command("pgrep", "light-launcher-instance").Output()
	if len(output) == 0 {
		output, _ = exec.Command("pgrep", "light-launcher-instan").Output()
	}

	pids := strings.Split(strings.TrimSpace(string(output)), "\n")
	sessions := make([]types.RunningSession, 0)

	for _, pidString := range pids {
		if pidString == "" {
			continue
		}
		pid, err := strconv.Atoi(pidString)
		if err != nil {
			continue
		}

		cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
		content, err := os.ReadFile(cmdlinePath)
		if err != nil {
			continue
		}

		arguments := strings.Split(string(content), "\x00")
		gamePath := ""
		for index, argument := range arguments {
			if argument == "--game" && index+1 < len(arguments) {
				gamePath = arguments[index+1]
				break
			}
		}

		if gamePath != "" {
			name := filepath.Base(gamePath)
			name = strings.TrimSuffix(name, filepath.Ext(name))

			cleanedPath := filepath.Clean(gamePath)
			if absolutePath, err := filepath.Abs(cleanedPath); err == nil {
				cleanedPath = absolutePath
			}

			sessions = append(sessions, types.RunningSession{
				Pid:      pid,
				GamePath: cleanedPath,
				GameName: name,
			})
		}
	}
	return sessions, nil
}

func (app *App) KillSession(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Signal(os.Interrupt)
}

func (app *App) RemoveGame(executablePath string) error {
	cfg, err := app.GetConfig(executablePath)
	if err != nil {
		return fmt.Errorf("could not find game to remove: %w", err)
	}

	configDirectory := config.GetExecutableConfigPath(cfg.Name, cfg.ID)

	if _, err := os.Stat(configDirectory); err == nil {
		if err := os.RemoveAll(configDirectory); err != nil {
			return fmt.Errorf("failed to remove game config: %w", err)
		}
	}

	_ = lsfg.DisableProfileInConfig(cfg.Name, executablePath)
	return nil
}

func (app *App) RunPrefixTool(prefixPath, toolName, protonPath string) error {
	options := types.LaunchOptions{
		GamePath:   toolName,
		PrefixPath: prefixPath,
		ProtonPath: protonPath,
	}
	commandArguments, environment := builder.BuildCommand(options)
	command := exec.Command(commandArguments[0], commandArguments[1:]...)
	command.Env = environment
	return command.Start()
}

func (app *App) GetConfig(executablePath string) (*types.LaunchOptions, error) {
	return config.LoadGameConfig(executablePath)
}

func (app *App) SavePrefixConfig(prefixName string, options types.LaunchOptions) error {
	return config.SavePrefixConfig(prefixName, options)
}

func (app *App) LoadPrefixConfig(prefixName string) (*types.LaunchOptions, error) {
	return config.LoadPrefixConfig(prefixName)
}

func (app *App) SaveGameConfig(options types.LaunchOptions) error {
	return config.SaveGameConfig(options)
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
