package systemsvc

import (
	"context"

	"light-launcher/core/internal/system"
	"light-launcher/core/internal/types"
)

func systemToolsStatus(_ context.Context) types.SystemToolsStatus {
	return types.SystemToolsStatus{
		HasGamescope:  system.IsCommandAvailable("gamescope"),
		HasMangoHud:   system.IsCommandAvailable("mangohud"),
		HasGameMode:   system.IsCommandAvailable("gamemoderun"),
		HasVulkanInfo: system.IsCommandAvailable("vulkaninfo"),
	}
}

func (s *Service) utilsStatus(_ context.Context) types.UtilsStatus {
	installed := s.lsfgInstalled()
	status := types.UtilsStatus{IsLsfgInstalled: installed}
	if installed {
		status.LsfgVersion = "Installed"
	}
	return status
}
