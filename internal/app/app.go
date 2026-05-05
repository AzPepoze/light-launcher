package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"light-launcher/internal/system"
	"light-launcher/internal/types"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//wails:service
type App struct {
	context context.Context
}

func NewApp() *App {
	return &App{}
}

func (app *App) Startup(ctx context.Context) {
	app.context = ctx
}

func (app *App) GetInitialLauncherPath() string {
	return os.Getenv("LIGHT_LAUNCHER_LAUNCHER_PATH")
}

func (app *App) GetInitialGamePath() string {
	return os.Getenv("LIGHT_LAUNCHER_GAME_PATH")
}

func (app *App) GetShouldEditLsfg() bool {
	return os.Getenv("LIGHT_LAUNCHER_EDIT_LSFG") == "1"
}

func (app *App) CloseWindow() {
	application.Get().Quit()
}

func (app *App) GetExeIcon(executablePath string) string {
	if _, err := os.Stat(executablePath); os.IsNotExist(err) {
		return ""
	}

	temporaryDirectory, err := os.MkdirTemp("", "light-launcher-icon-*")
	if err != nil {
		return ""
	}
	defer os.RemoveAll(temporaryDirectory)

	wrestoolCommand := exec.Command("wrestool", "-x", "--output="+temporaryDirectory, executablePath)
	if err := wrestoolCommand.Run(); err == nil {
		matches, _ := filepath.Glob(filepath.Join(temporaryDirectory, "*.ico"))
		if len(matches) > 0 {
			if icon := tryReadIcon(matches[0]); icon != "" {
				return icon
			}
		}
	}

	icoextractCommand := exec.Command("icoextract", executablePath, filepath.Join(temporaryDirectory, "icon.ico"))
	if err := icoextractCommand.Run(); err == nil {
		if icon := tryReadIcon(filepath.Join(temporaryDirectory, "icon.ico")); icon != "" {
			return icon
		}
		matches, _ := filepath.Glob(filepath.Join(temporaryDirectory, "*.ico"))
		if len(matches) > 0 {
			return tryReadIcon(matches[0])
		}
	}

	return ""
}

func tryReadIcon(iconPath string) string {
	data, err := os.ReadFile(iconPath)
	if err != nil || len(data) == 0 {
		return ""
	}
	return "data:image/x-icon;base64," + base64.StdEncoding.EncodeToString(data)
}

func (app *App) GetSystemToolsStatus() types.SystemToolsStatus {
	return system.GetSystemToolsStatus()
}

func (app *App) GetSystemInfo() types.SystemInfo {
	return system.GetSystemInfo()
}

func (app *App) GetSystemUsage() types.SystemUsage {
	return system.GetSystemUsage()
}

func (app *App) GetShaderCacheSize() string {
	return system.GetShaderCacheSize()
}

func (app *App) ClearShaderCache() error {
	return system.ClearShaderCache()
}

func (app *App) DropCaches() error {
	return system.DropCaches()
}

func (app *App) ClearSwap() error {
	return system.ClearSwap()
}

func (app *App) CleanupProcesses() error {
	commands := []string{
		"umu-run",
		"pressure-vessel",
		"gamescopereaper",
		"steam-runtime-launcher-service",
		"srt-bwrap",
		"reaper",
	}
	for _, command := range commands {
		_ = exec.Command("pkill", "-f", command).Run()
	}
	return nil
}

func (app *App) GetTotalRam() (int, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var memoryTotalKb int
	var label string
	for {
		_, err := fmt.Fscanf(file, "%s %d kB\n", &label, &memoryTotalKb)
		if err != nil || label == "MemTotal:" {
			break
		}
	}

	if memoryTotalKb == 0 {
		return 0, fmt.Errorf("failed to parse MemTotal")
	}

	return memoryTotalKb / 1024 / 1024, nil
}

func (app *App) GetImageBase64(imagePath string) string {
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return ""
	}
	data, err := os.ReadFile(imagePath)
	if err != nil || len(data) == 0 {
		return ""
	}
	extension := filepath.Ext(imagePath)
	mimeType := "image/png"
	switch extension {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".svg":
		mimeType = "image/svg+xml"
	case ".webp":
		mimeType = "image/webp"
	}
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data))
}
