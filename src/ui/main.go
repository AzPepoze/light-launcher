package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"light-launcher-wails/backend"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if len(os.Args) > 1 {
		gamePath := os.Args[1]

		if _, err := os.Stat(gamePath); err == nil {
			if absPath, err := filepath.Abs(gamePath); err == nil {
				os.Setenv("LIGHT_LAUNCHER_LAUNCHER_PATH", absPath)
				fmt.Printf("Pre-selecting launcher path: %s\n", absPath)
			}
		}
	}

	app := backend.NewApp()

	wailsApp := application.New(application.Options{
		Name: "LightLauncher",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "LightLauncher",
		Width:            1024,
		Height:           768,
		BackgroundColour: application.NewRGB(24, 24, 27),
	})

	wailsApp.RegisterService(application.NewService(app))

	err := wailsApp.Run()

	if err != nil {
		println("Error:", err.Error())
	}
}
