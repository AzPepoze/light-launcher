package app

import (
	"light-launcher/internal/system"
	"light-launcher/internal/types"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func (app *App) ScanProtonVersions() ([]types.ProtonTool, error) {
	return system.GetProtonTools()
}

func (app *App) GetProtonVariants() []system.ProtonVariant {
	return system.GetKnownVariants()
}

func (app *App) GetProtonReleases(variantID string) ([]system.GitHubRelease, error) {
	return system.FetchReleases(variantID)
}

func (app *App) InstallProtonVersion(url, version string) error {
	return system.InstallProton(url, version, func(percent int, message string) {
		application.Get().Event.Emit("install-proton-progress", map[string]interface{}{
			"percent": percent,
			"message": message,
		})
	})
}
