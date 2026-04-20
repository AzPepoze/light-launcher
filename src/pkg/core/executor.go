package core

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var debugLogFile *os.File

func InitDebugLog() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	logPath := filepath.Join(homeDir, "LightLauncher", "debug.log")
	os.MkdirAll(filepath.Dir(logPath), 0755)

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		debugLogFile = f
		DebugLog("=== DEBUG LOG STARTED ===")
	}
}

func DebugLog(msg string) {
	if debugLogFile == nil {
		InitDebugLog()
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] %s\n", timestamp, msg)

	// Write to console
	fmt.Print(line)

	// Write to file
	if debugLogFile != nil {
		debugLogFile.WriteString(line)
		debugLogFile.Sync()
	}
}

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func buildLsfgEnv(opts LaunchOptions) []string {
	if !opts.EnableLsfgVk {
		return nil
	}

	DebugLog("buildLsfgEnv() called")

	// lsfg-vk reads from ~/.config/lsfg-vk/conf.toml by default
	// We don't pass LSFG_CONFIG env var - just enable LSFG
	return []string{}
}

func buildBaseEnv(opts LaunchOptions) []string {
	baseEnv := []string{
		fmt.Sprintf("WINEPREFIX=%s", ExpandPath(opts.PrefixPath)),
		fmt.Sprintf("UMU_PROTON_PATTERN=%s", opts.ProtonPattern),
	}

	if opts.ProtonPath != "" {
		baseEnv = append(baseEnv, fmt.Sprintf("PROTONPATH=%s", ExpandPath(opts.ProtonPath)))
	}

	return baseEnv
}

func BuildCommand(opts LaunchOptions) ([]string, []string) {
	// DEBUG: Log received options
	DebugLog("BuildCommand() called")
	DebugLog("  LauncherPath: " + opts.LauncherPath)
	DebugLog("  MainExecutablePath: " + opts.MainExecutablePath)
	DebugLog("  HaveGameExe: " + fmt.Sprintf("%v", opts.HaveGameExe))
	DebugLog("  EnableLsfgVk: " + fmt.Sprintf("%v", opts.EnableLsfgVk))

	var cmdArgs []string
	env := append(os.Environ(), buildBaseEnv(opts)...)

	lsfgEnv := buildLsfgEnv(opts)

	if !opts.EnableGamescope {
		if opts.EnableMangoHud {
			env = append(env, "MANGOHUD=1")
		}
		env = append(env, lsfgEnv...)
	}

	if opts.EnableGamemode && isCommandAvailable("gamemoderun") {
		cmdArgs = append(cmdArgs, "gamemoderun")
	}

	if opts.EnableGamescope && isCommandAvailable("gamescope") {
		cmdArgs = append(cmdArgs, "gamescope")
		if opts.GamescopeW != "" {
			cmdArgs = append(cmdArgs, "-W", opts.GamescopeW)
		}
		if opts.GamescopeH != "" {
			cmdArgs = append(cmdArgs, "-H", opts.GamescopeH)
		}
		if opts.GamescopeR != "" {
			cmdArgs = append(cmdArgs, "-r", opts.GamescopeR)
		}
		cmdArgs = append(cmdArgs, "--", "env")

		if opts.EnableMangoHud {
			env = append(env, "MANGOHUD=1")
		}
		env = append(env, lsfgEnv...)
	}

	cmdArgs = append(cmdArgs, "umu-run")

	// Always use launcher path for execution
	// MainExecutablePath is only used for LSFG-VK profile matching
	DebugLog("About to resolve exePath:")
	DebugLog("  opts.LauncherPath = " + opts.LauncherPath)
	DebugLog("  opts.MainExecutablePath = " + opts.MainExecutablePath)

	exePath := opts.LauncherPath
	if exePath == "" {
		DebugLog("LauncherPath is empty, falling back to MainExecutablePath")
		exePath = opts.MainExecutablePath
	}
	DebugLog("Final exePath: " + exePath)
	cmdArgs = append(cmdArgs, exePath)

	if opts.CustomArgs != "" {
		args := strings.Fields(opts.CustomArgs)
		cmdArgs = append(cmdArgs, args...)
	}

	if opts.EnableMemoryMin && opts.MemoryMinValue != "" && isCommandAvailable("systemd-run") {
		wrappedArgs := []string{"systemd-run", "--user", "--scope", fmt.Sprintf("-pMemoryMin=%s", opts.MemoryMinValue), "--"}
		cmdArgs = append(wrappedArgs, cmdArgs...)
	}

	return cmdArgs, env
}

func RunGameWithLog(cmdArgs []string, env []string, onLog func(string), onExit func()) (*os.Process, error) {
	if len(cmdArgs) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Env = env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	readPipe := func(r io.Reader, prefix string) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			onLog(fmt.Sprintf("%s%s", prefix, scanner.Text()))
		}
	}

	go readPipe(stdout, "")
	go readPipe(stderr, "")

	go func() {
		_ = cmd.Wait()
		onLog("\n--- Process Exited ---")
		if onExit != nil {
			onExit()
		}
	}()

	return cmd.Process, nil
}

func StopProcessGroup(proc *os.Process) error {
	if proc == nil {
		return nil
	}
	pgid := -proc.Pid
	_ = syscall.Kill(pgid, syscall.SIGINT)
	time.Sleep(200 * time.Millisecond)
	return syscall.Kill(pgid, syscall.SIGTERM)
}

func FormatCommandForDisplay(cmdArgs []string, opts LaunchOptions) string {
	var sb strings.Builder
	if opts.EnableMemoryMin && opts.MemoryMinValue != "" {
		sb.WriteString(fmt.Sprintf("[MemMin:%s] ", opts.MemoryMinValue))
	}
	sb.WriteString("WINEPREFIX=" + opts.PrefixPath + " ")
	sb.WriteString("UMU_PROTON_PATTERN=" + opts.ProtonPattern + " ")
	if opts.EnableMangoHud {
		sb.WriteString("MANGOHUD=1 ")
	}
	sb.WriteString(strings.Join(cmdArgs, " "))
	return sb.String()
}
