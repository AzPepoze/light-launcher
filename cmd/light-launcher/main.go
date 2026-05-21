package main

import (
	"fmt"
	"os"
	"path/filepath"

	"light-launcher/internal/app"
	"light-launcher/internal/config"
	"light-launcher/ui"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func main() {
	if len(os.Args) > 1 {
		gamePath := os.Args[1]

		if _, err := os.Stat(gamePath); err == nil {
			if absolutePath, err := filepath.Abs(gamePath); err == nil {
				os.Setenv("LIGHT_LAUNCHER_LAUNCHER_PATH", absolutePath)
				fmt.Printf("Pre-selecting launcher path: %s\n", absolutePath)
			}
		}
	}

	backendApp := app.NewApp()

	wailsApp := application.New(application.Options{
		Name: "LightLauncher",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(ui.Assets),
		},
	})

	appSettings := config.LoadAppSettings()

	backgroundType := application.BackgroundTypeSolid
	backgroundColour := application.NewRGBA(24, 24, 27, 255) // Default dark solid
	if appSettings.TransparentMode {
		backgroundType = application.BackgroundTypeTransparent
		backgroundColour = application.NewRGBA(0, 0, 0, 0)
	}

	wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "LightLauncher",
		Width:            1024,
		Height:           768,
		BackgroundColour: backgroundColour,
		BackgroundType:   backgroundType,
		EnableFileDrop:   true,
	})

	wailsApp.RegisterService(application.NewService(backendApp))

	if err := wailsApp.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
