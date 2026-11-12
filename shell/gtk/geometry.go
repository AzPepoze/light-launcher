//go:build linux

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	geometryVersion  = 1
	geometryFileName = "window-geometry.json"
	launcherDirName  = "light-launcher"
)

// screenArea is one monitor in logical pixels.
type screenArea struct {
	X      int
	Y      int
	Width  int
	Height int
}

// windowGeometry is the saved size and position of the launcher window.
type windowGeometry struct {
	Version    int  `json:"version"`
	X          int  `json:"x"`
	Y          int  `json:"y"`
	Width      int  `json:"width"`
	Height     int  `json:"height"`
	Positioned bool `json:"positioned"`
}

func defaultGeometry() windowGeometry {
	return windowGeometry{Version: geometryVersion, Width: defaultWidth, Height: defaultHeight}
}

// geometryPath matches the launcher config directory, which this module cannot import.
func geometryPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, launcherDirName, geometryFileName)
}

func loadGeometry(path string) (windowGeometry, bool) {
	if path == "" {
		return windowGeometry{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return windowGeometry{}, false
	}
	saved, err := decodeGeometry(data)
	if err != nil {
		log.Warn("window", "ignoring the saved window geometry: %v", err)
		return windowGeometry{}, false
	}
	return saved, true
}

func decodeGeometry(data []byte) (windowGeometry, error) {
	var saved windowGeometry
	if err := json.Unmarshal(data, &saved); err != nil {
		return windowGeometry{}, fmt.Errorf("read window geometry: %w", err)
	}
	return upgradeGeometry(saved)
}

// upgradeGeometry moves an older saved file to the current shape.
func upgradeGeometry(saved windowGeometry) (windowGeometry, error) {
	if saved.Version != geometryVersion {
		return windowGeometry{}, fmt.Errorf("unsupported window geometry version %d", saved.Version)
	}
	return saved, nil
}

func saveGeometry(path string, geometry windowGeometry) error {
	if path == "" {
		return errors.New("no config folder for the window geometry")
	}
	geometry.Version = geometryVersion
	data, err := json.Marshal(geometry)
	if err != nil {
		return fmt.Errorf("encode window geometry: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create the config folder: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write the window geometry: %w", err)
	}
	return nil
}

// usableGeometry keeps a saved size and position only when they still fit on a screen.
func usableGeometry(saved windowGeometry, screens []screenArea) windowGeometry {
	if saved.Width < minWidth || saved.Height < minHeight {
		return defaultGeometry()
	}
	if saved.Positioned && !onScreen(saved, screens) {
		return defaultGeometry()
	}
	saved.Version = geometryVersion
	return saved
}

// onScreen reports whether the top-left corner is on one of the monitors.
func onScreen(geometry windowGeometry, screens []screenArea) bool {
	for _, screen := range screens {
		if geometry.X >= screen.X && geometry.X < screen.X+screen.Width &&
			geometry.Y >= screen.Y && geometry.Y < screen.Y+screen.Height {
			return true
		}
	}
	return false
}
