//go:build linux

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGeometryPathUsesTheConfigFolder(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	want := filepath.Join(dir, launcherDirName, geometryFileName)
	if got := geometryPath(); got != want {
		t.Fatalf("geometryPath = %q, want %q", got, want)
	}
}

func TestSaveAndLoadGeometry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", geometryFileName)
	want := windowGeometry{
		X: 120, Y: 80, Width: 1280, Height: 800, Positioned: true,
	}
	if err := saveGeometry(path, want); err != nil {
		t.Fatalf("saveGeometry: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("saved file is not JSON: %v", err)
	}
	if got, ok := raw["version"].(float64); !ok || int(got) != geometryVersion {
		t.Fatalf("saved version = %v, want %d", raw["version"], geometryVersion)
	}

	got, ok := loadGeometry(path)
	if !ok {
		t.Fatal("loadGeometry ignored a file it just saved")
	}
	want.Version = geometryVersion
	if got != want {
		t.Fatalf("loaded = %+v, want %+v", got, want)
	}
}

func TestLoadGeometryIgnoresBadFiles(t *testing.T) {
	dir := t.TempDir()
	tests := []struct {
		name string
		data string
	}{
		{name: "malformed", data: "{"},
		{name: "unversioned", data: `{"width":1200,"height":800}`},
		{name: "future version", data: `{"version":2,"width":1200,"height":800}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(dir, test.name+".json")
			if err := os.WriteFile(path, []byte(test.data), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, ok := loadGeometry(path); ok {
				t.Fatalf("loadGeometry accepted %s", test.name)
			}
		})
	}
	if _, ok := loadGeometry(filepath.Join(dir, "missing.json")); ok {
		t.Fatal("loadGeometry accepted a missing file")
	}
}

func TestUsableGeometry(t *testing.T) {
	screens := []screenArea{
		{X: 0, Y: 0, Width: 1920, Height: 1080},
		{X: 1920, Y: 0, Width: 1920, Height: 1080},
	}
	tests := []struct {
		name   string
		saved  windowGeometry
		screen []screenArea
		want   windowGeometry
	}{
		{
			name:   "visible position is kept",
			saved:  windowGeometry{X: 120, Y: 80, Width: 1280, Height: 800, Positioned: true},
			screen: screens,
			want:   windowGeometry{X: 120, Y: 80, Width: 1280, Height: 800, Positioned: true, Version: geometryVersion},
		},
		{
			name:   "position on the second monitor is kept",
			saved:  windowGeometry{X: 2000, Y: 40, Width: 1000, Height: 700, Positioned: true},
			screen: screens,
			want:   windowGeometry{X: 2000, Y: 40, Width: 1000, Height: 700, Positioned: true, Version: geometryVersion},
		},
		{
			name:   "off-screen position falls back to the default",
			saved:  windowGeometry{X: 5000, Y: 80, Width: 1280, Height: 800, Positioned: true},
			screen: screens,
			want:   defaultGeometry(),
		},
		{
			name:   "negative position falls back to the default",
			saved:  windowGeometry{X: -2000, Y: -100, Width: 1280, Height: 800, Positioned: true},
			screen: screens,
			want:   defaultGeometry(),
		},
		{
			name:   "unknown screen falls back to the default",
			saved:  windowGeometry{X: 120, Y: 80, Width: 1280, Height: 800, Positioned: true},
			screen: nil,
			want:   defaultGeometry(),
		},
		{
			name:   "unpositioned size is kept",
			saved:  windowGeometry{Width: 1400, Height: 900},
			screen: screens,
			want:   windowGeometry{Width: 1400, Height: 900, Version: geometryVersion},
		},
		{
			name:   "too small falls back to the default",
			saved:  windowGeometry{X: 120, Y: 80, Width: minWidth - 1, Height: 800, Positioned: true},
			screen: screens,
			want:   defaultGeometry(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := usableGeometry(test.saved, test.screen); got != test.want {
				t.Fatalf("usableGeometry = %+v, want %+v", got, test.want)
			}
		})
	}
}
