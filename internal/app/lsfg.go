package app

import (
	"os"
	"path/filepath"

	"light-launcher/internal/executor"
	"light-launcher/internal/system"
	"light-launcher/internal/types"
	"light-launcher/lib/lsfg"

	"github.com/pelletier/go-toml/v2"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func (app *App) GetUtilsStatus() types.UtilsStatus {
	return types.UtilsStatus{
		IsLsfgInstalled: lsfg.IsInstalled(),
		LsfgVersion:     lsfg.GetVersion(),
	}
}

func (app *App) GetLsfgProfileForGame(name, gamePath string) (*types.LsfgProfileData, error) {
	profile, _, err := lsfg.FindProfileForGame(gamePath)
	if err != nil {
		return nil, nil
	}

	var dllPath string
	var allowFp16 bool
	configPath, err := lsfg.GetConfigPath()
	if err == nil {
		data, err := os.ReadFile(configPath)
		if err == nil {
			var configFile lsfg.ConfigFile
			if err := toml.Unmarshal(data, &configFile); err == nil {
				dllPath = configFile.Global.DLL
				allowFp16 = configFile.Global.AllowFP16
			}
		}
	}

	return &types.LsfgProfileData{
		Name:            profile.Name,
		Multiplier:      profile.Multiplier,
		PerformanceMode: profile.PerformanceMode,
		GPU:             profile.GPU,
		FlowScale:       profile.FlowScale,
		Pacing:          profile.Pacing,
		DllPath:         dllPath,
		AllowFp16:       allowFp16,
	}, nil
}

func (app *App) InstallLsfg() error {
	return lsfg.Install(func(percent int, message string) {
		application.Get().Event.Emit("lsfg-install-progress", map[string]interface{}{
			"percent": percent,
			"message": message,
		})
	})
}

func (app *App) DetectLosslessDll() string {
	homeDirectory, _ := os.UserHomeDir()
	potentialPaths := []string{
		filepath.Join(homeDirectory, ".steam/root/steamapps/common/Lossless Scaling/Lossless.dll"),
		filepath.Join(homeDirectory, ".local/share/Steam/steamapps/common/Lossless Scaling/Lossless.dll"),
	}

	for _, path := range potentialPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func (app *App) GetListGpus() []string {
	return system.GetListGpus()
}

func (app *App) UninstallLsfg() error {
	return lsfg.Uninstall(executor.DebugLog)
}

func (app *App) SaveLsfgProfile(profileName, gamePath string, multiplier int, performanceMode bool, dllPath, gpu, flowScale, pacing string, allowFp16 bool) error {
	if gpu == "" {
		gpuList := system.GetListGpus()
		if len(gpuList) > 0 {
			gpu = gpuList[0]
			executor.DebugLog("SaveLsfgProfile() GPU was blank, using first GPU: " + gpu)
		}
	}

	configPath, err := lsfg.GetConfigPath()
	if err != nil {
		return err
	}

	return lsfg.SaveProfileToPath(profileName, gamePath, configPath, multiplier, performanceMode, dllPath, gpu, flowScale, pacing, allowFp16)
}

func (app *App) DisableLsfgProfile(profileName, gamePath string) error {
	return lsfg.DisableProfileInConfig(profileName, gamePath)
}

func (app *App) RemoveProfile(mainExecutablePath string) error {
	return lsfg.RemoveProfileFromConfig(mainExecutablePath)
}

func (app *App) EditLsfgConfigForGame(mainExecutablePath string) error {
	_, _, err := lsfg.FindProfileForGame(mainExecutablePath)
	return err
}
