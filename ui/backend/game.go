package backend

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"light-launcher/internal/core"
	"light-launcher/internal/lsfg_utils"
)

func (a *App) RunGame(opts core.LaunchOptions, showLogs bool) error {
	core.DebugLog("[APP.RunGame] Called with options:")
	core.DebugLog("[APP.RunGame]   LauncherPath: " + opts.LauncherPath)
	core.DebugLog("[APP.RunGame]   MainExecutablePath: " + opts.MainExecutablePath)
	core.DebugLog("[APP.RunGame]   HaveGameExe: " + fmt.Sprintf("%v", opts.HaveGameExe))
	core.DebugLog("[APP.RunGame]   EnableLsfgVk: " + fmt.Sprintf("%v", opts.EnableLsfgVk))
	core.DebugLog("[APP.RunGame] Full opts: " + fmt.Sprintf("%+v", opts))

	if _, err := os.Stat(opts.MainExecutablePath); os.IsNotExist(err) {
		return fmt.Errorf("game executable not found at: %s", opts.MainExecutablePath)
	}

	if !opts.HaveGameExe && opts.LauncherPath != "" {
		core.DebugLog("[APP.RunGame] NORMALIZE: HaveGameExe=false, enforcing MainExecutablePath=LauncherPath")
		opts.MainExecutablePath = opts.LauncherPath
	}

	_ = core.SaveGameConfig(opts)

	if opts.EnableLsfgVk {
		core.DebugLog("[APP.RunGame] LSFG-VK enabled, ensuring profile exists")
		configPath, err := lsfg_utils.GetConfigPath()
		if err == nil {
			_, _, err := lsfg_utils.FindProfileForGameAtPath(opts.MainExecutablePath, configPath)
			if err != nil {
				core.DebugLog("[APP.RunGame] No profile found, creating one with current options")
				gpu := opts.LsfgGpu
				if gpu == "" {
					gpuList := core.GetListGpus()
					if len(gpuList) > 0 {
						gpu = gpuList[0]
						core.DebugLog("[APP.RunGame] GPU was blank, using first GPU: " + gpu)
					}
				}

				_ = lsfg_utils.SaveProfileToPath(opts.MainExecutablePath, configPath,
					parseMultiplier(opts.LsfgMultiplier),
					opts.LsfgPerfMode,
					opts.LsfgDllPath,
					gpu,
					opts.LsfgFlowScale,
					opts.LsfgPacing,
					opts.LsfgAllowFp16)
			}
		}
	}

	instanceName := "light-launcher-instance"
	exeDir := "."
	if exe, err := os.Executable(); err == nil {
		exeDir = filepath.Dir(exe)
	}

	potentialPaths := []string{
		// 1. Same directory as current executable
		filepath.Join(exeDir, instanceName),
		// 2. Project bin directory (relative to ui/ frontend dev)
		filepath.Join(exeDir, "../bin", instanceName),
		// 3. Project bin directory (relative to ui/backend)
		"../bin/" + instanceName,
		// 4. Local bin or current dir
		"./bin/" + instanceName,
		"./" + instanceName,
		// 5. System path (fallback)
		"/usr/bin/" + instanceName,
	}

	foundPath := ""
	for _, p := range potentialPaths {
		if _, err := os.Stat(p); err == nil {
			if abs, err := filepath.Abs(p); err == nil {
				foundPath = abs
			} else {
				foundPath = p
			}
			break
		}
	}

	if foundPath == "" {
		return fmt.Errorf("instance manager '%s' not found in potential paths", instanceName)
	}

	core.DebugLog("[APP.RunGame] Using instance manager: " + foundPath)

	args := []string{
		"--game", opts.MainExecutablePath,
		"--launcher", opts.LauncherPath,
		"--prefix", opts.PrefixPath,
		"--proton-pattern", opts.ProtonPattern,
		"--proton-path", opts.ProtonPath,
	}
	if opts.EnableMangoHud {
		args = append(args, "--mango")
	}
	if opts.EnableGamemode {
		args = append(args, "--gamemode")
	}
	if opts.EnableLsfgVk {
		args = append(args, "--lsfg")
		args = append(args, "--lsfg-mult", opts.LsfgMultiplier)
		if opts.LsfgPerfMode {
			args = append(args, "--lsfg-perf")
		}
		if opts.LsfgDllPath != "" {
			args = append(args, "--lsfg-dll-path", opts.LsfgDllPath)
		}
	}
	if opts.EnableMemoryMin {
		args = append(args, "--memory-min")
		if opts.MemoryMinValue != "" {
			args = append(args, "--memory-min-value", opts.MemoryMinValue)
		}
	}
	if opts.EnableGamescope {
		args = append(args, "--gamescope")
		args = append(args, "--gs-w", opts.GamescopeW)
		args = append(args, "--gs-h", opts.GamescopeH)
		args = append(args, "--gs-r", opts.GamescopeR)
	}
	if !showLogs {
		args = append(args, "--logs=false")
	}

	cmd := exec.Command(foundPath, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start instance manager: %w", err)
	}
	go cmd.Process.Release()

	return nil
}

func (a *App) GetAllGames() ([]GameInfo, error) {
	configs, err := core.ListGameConfigs()
	if err != nil {
		return nil, err
	}

	games := make([]GameInfo, 0)
	for _, cfg := range configs {
		name := filepath.Base(cfg.MainExecutablePath)
		name = strings.TrimSuffix(name, filepath.Ext(name))

		cleanedPath := filepath.Clean(cfg.MainExecutablePath)
		if abs, err := filepath.Abs(cleanedPath); err == nil {
			cleanedPath = abs
		}

		games = append(games, GameInfo{
			Name:   name,
			Path:   cleanedPath,
			Config: cfg,
		})
	}
	return games, nil
}

func (a *App) GetRunningSessions() ([]RunningSession, error) {
	out, _ := exec.Command("pgrep", "light-launcher-instance").Output()
	if len(out) == 0 {
		out, _ = exec.Command("pgrep", "light-launcher-instan").Output()
	}

	pids := strings.Split(strings.TrimSpace(string(out)), "\n")
	sessions := make([]RunningSession, 0)

	for _, pidStr := range pids {
		if pidStr == "" {
			continue
		}
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
		content, err := os.ReadFile(cmdlinePath)
		if err != nil {
			continue
		}

		args := strings.Split(string(content), "\x00")
		gamePath := ""
		for i, arg := range args {
			if arg == "--game" && i+1 < len(args) {
				gamePath = args[i+1]
				break
			}
		}

		if gamePath != "" {
			name := filepath.Base(gamePath)
			name = strings.TrimSuffix(name, filepath.Ext(name))

			cleanedPath := filepath.Clean(gamePath)
			if abs, err := filepath.Abs(cleanedPath); err == nil {
				cleanedPath = abs
			}

			sessions = append(sessions, RunningSession{
				Pid:      pid,
				GamePath: cleanedPath,
				GameName: name,
			})
		}
	}
	return sessions, nil
}

func (a *App) KillSession(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Signal(os.Interrupt)
}

func (a *App) RemoveGame(exePath string) error {
	core.DebugLog("[APP.RemoveGame] Removing game config for: " + exePath)
	configDir := core.GetConfigPath(exePath)

	if _, err := os.Stat(configDir); err == nil {
		err = os.RemoveAll(configDir)
		if err != nil {
			return fmt.Errorf("failed to remove game config: %w", err)
		}
	}

	// Also try to remove LSFG profile if it exists (don't error if it doesn't)
	_ = lsfg_utils.RemoveProfileFromConfig(exePath)

	return nil
}

func (a *App) RunPrefixTool(prefixPath, toolName, protonPattern string) error {
	opts := core.LaunchOptions{
		MainExecutablePath: toolName,
		PrefixPath:         prefixPath,
		ProtonPattern:      protonPattern,
	}
	cmdArgs, env := core.BuildCommand(opts)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Env = env
	return cmd.Start()
}

func (a *App) GetConfig(exePath string) (*core.LaunchOptions, error) {
	return core.LoadGameConfig(exePath)
}

func (a *App) SavePrefixConfig(prefixName string, opts core.LaunchOptions) error {
	return core.SavePrefixConfig(prefixName, opts)
}

func (a *App) LoadPrefixConfig(prefixName string) (*core.LaunchOptions, error) {
	return core.LoadPrefixConfig(prefixName)
}

func (a *App) SaveGameConfig(opts core.LaunchOptions) error {
	return core.SaveGameConfig(opts)
}

func parseMultiplier(mult string) int {
	var val int = 2
	if mult != "" {
		if _, err := fmt.Sscanf(mult, "%d", &val); err != nil {
			val = 2
		}
	}
	return val
}
