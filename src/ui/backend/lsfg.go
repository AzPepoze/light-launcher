package backend

import (
	"os"
	"path/filepath"

	"light-launcher/pkg/core"
	"light-launcher/pkg/lsfg_utils"

	"github.com/pelletier/go-toml/v2"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func (a *App) GetUtilsStatus() core.UtilsStatus {
	return core.UtilsStatus{
		IsLsfgInstalled: lsfg_utils.IsInstalled(),
		LsfgVersion:     lsfg_utils.GetVersion(),
	}
}

func (a *App) InstallLsfg() error {
	return lsfg_utils.Install(func(percent int, msg string) {
		application.Get().Event.Emit("lsfg-install-progress", map[string]interface{}{
			"percent": percent,
			"message": msg,
		})
	})
}

func (a *App) DetectLosslessDll() string {
	home, _ := os.UserHomeDir()
	paths := []string{
		filepath.Join(home, ".steam/root/steamapps/common/Lossless Scaling/Lossless.dll"),
		filepath.Join(home, ".local/share/Steam/steamapps/common/Lossless Scaling/Lossless.dll"),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func (a *App) GetListGpus() []string {
	return core.GetListGpus()
}

func (a *App) UninstallLsfg() error {
	return lsfg_utils.Uninstall(core.DebugLog)
}

func (a *App) GetLsfgProfileForGame(mainExePath string) (*LsfgProfileData, error) {
	profile, _, err := lsfg_utils.FindProfileForGame(mainExePath)
	if err != nil {
		return nil, nil
	}

	var dllPath string
	var allowFp16 bool
	configPath, err := lsfg_utils.GetConfigPath()
	if err == nil {
		data, err := os.ReadFile(configPath)
		if err == nil {
			var config lsfg_utils.ConfigFile
			if err := toml.Unmarshal(data, &config); err == nil {
				dllPath = config.Global.DLL
				allowFp16 = config.Global.AllowFP16
			}
		}
	}

	return &LsfgProfileData{
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

func (a *App) SaveLsfgProfile(mainExePath string, multiplier int, perfMode bool, dllPath, gpu, flowScale, pacing string, allowFp16 bool) error {
	if gpu == "" {
		gpuList := core.GetListGpus()
		if len(gpuList) > 0 {
			gpu = gpuList[0]
			core.DebugLog("SaveLsfgProfile() GPU was blank, using first GPU: " + gpu)
		}
	}

	configPath, err := lsfg_utils.GetConfigPath()
	if err != nil {
		return err
	}

	return lsfg_utils.SaveProfileToPath(mainExePath, configPath, multiplier, perfMode, dllPath, gpu, flowScale, pacing, allowFp16)
}

func (a *App) RemoveProfile(mainExePath string) error {
	return lsfg_utils.RemoveProfileFromConfig(mainExePath)
}

func (a *App) EditLsfgConfigForGame(mainExePath string) error {
	_, _, err := lsfg_utils.FindProfileForGame(mainExePath)
	return err
}
