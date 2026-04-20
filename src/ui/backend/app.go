package backend

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"light-launcher/pkg/core"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type LsfgProfileData struct {
	Name            string  `json:"name"`
	Multiplier      int     `json:"multiplier"`
	PerformanceMode bool    `json:"performanceMode"`
	GPU             string  `json:"gpu"`
	FlowScale       float32 `json:"flowScale"`
	Pacing          string  `json:"pacing"`
	DllPath         string  `json:"dllPath"`
	AllowFp16       bool    `json:"allowFp16"`
}

type GameInfo struct {
	Name     string             `json:"name"`
	Path     string             `json:"path"`
	Icon     string             `json:"icon"`
	Config   core.LaunchOptions `json:"config"`
	IsRecent bool               `json:"isRecent"`
}

type RunningSession struct {
	Pid      int    `json:"pid"`
	GamePath string `json:"gamePath"`
	GameName string `json:"gameName"`
}

// App struct
//
//wails:service
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetInitialLauncherPath() string {
	return os.Getenv("LIGHT_LAUNCHER_LAUNCHER_PATH")
}

func (a *App) GetInitialGamePath() string {
	return os.Getenv("LIGHT_LAUNCHER_GAME_PATH")
}

func (a *App) GetShouldEditLsfg() bool {
	return os.Getenv("LIGHT_LAUNCHER_EDIT_LSFG") == "1"
}

func (a *App) CloseWindow() {
	application.Get().Quit()
}

func (a *App) GetExeIcon(exePath string) string {
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		return ""
	}

	tempDir, err := os.MkdirTemp("", "light-launcher-icon-*")
	if err != nil {
		return ""
	}
	defer os.RemoveAll(tempDir)

	cmd := exec.Command("wrestool", "-x", "--output="+tempDir, exePath)
	if err := cmd.Run(); err == nil {
		matches, _ := filepath.Glob(filepath.Join(tempDir, "*.ico"))
		if len(matches) > 0 && tryReadIcon(matches[0]) != "" {
			return tryReadIcon(matches[0])
		}
	}

	cmd = exec.Command("icoextract", exePath, filepath.Join(tempDir, "icon.ico"))
	if err := cmd.Run(); err == nil {
		if icon := tryReadIcon(filepath.Join(tempDir, "icon.ico")); icon != "" {
			return icon
		}
		matches, _ := filepath.Glob(filepath.Join(tempDir, "*.ico"))
		if len(matches) > 0 {
			return tryReadIcon(matches[0])
		}
	}

	return ""
}

func tryReadIcon(icoPath string) string {
	data, err := os.ReadFile(icoPath)
	if err != nil || len(data) == 0 {
		return ""
	}
	return "data:image/x-icon;base64," + base64.StdEncoding.EncodeToString(data)
}

func (a *App) GetSystemToolsStatus() core.SystemToolsStatus {
	return core.GetSystemToolsStatus()
}

func (a *App) GetSystemInfo() core.SystemInfo {
	return core.GetSystemInfo()
}

func (a *App) GetSystemUsage() core.SystemUsage {
	return core.GetSystemUsage()
}

func (a *App) GetShaderCacheSize() string {
	return core.GetShaderCacheSize()
}

func (a *App) ClearShaderCache() error {
	return core.ClearShaderCache()
}

func (a *App) DropCaches() error {
	return core.DropCaches()
}

func (a *App) ClearSwap() error {
	return core.ClearSwap()
}

func (a *App) CleanupProcesses() error {
	commands := []string{
		"umu-run",
		"pressure-vessel",
		"gamescopereaper",
		"steam-runtime-launcher-service",
		"srt-bwrap",
		"reaper",
	}
	for _, cmd := range commands {
		_ = exec.Command("pkill", "-f", cmd).Run()
	}
	return nil
}

func (a *App) GetTotalRam() (int, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var memTotalKb int
	var line string
	for {
		_, err := fmt.Fscanf(file, "%s %d kB\n", &line, &memTotalKb)
		if err != nil || line == "MemTotal:" {
			break
		}
	}

	if memTotalKb == 0 {
		return 0, fmt.Errorf("failed to parse MemTotal")
	}

	return memTotalKb / 1024 / 1024, nil
}
