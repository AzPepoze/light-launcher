package main

import (
	"context"
	"slices"
	"testing"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/library"
)

// useTempConfig keeps the databases and folders of a test out of the real config folder.
func useTempConfig(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
}

func TestWireServicesRegistersEveryService(t *testing.T) {
	useTempConfig(t)
	router := api.NewRouter()
	svc, err := wireServices(context.Background(), router, api.NewHub())
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Close()

	methods := router.Methods()
	for _, want := range []string{
		"GetInitialGamePath", "GetAppSettings", "GetAllGames", "RunGame", "GetPlatform",
		"GetSystemInfo", "LibraryList", "PluginsList", "DownloadsList", "DownloadEnqueue",
		"SchedulerStatus", "SchedulerRunNow",
	} {
		if !slices.Contains(methods, want) {
			t.Errorf("method %s is not registered", want)
		}
	}
}

func TestPickUpdateFile(t *testing.T) {
	name := func(value string) *string { return &value }
	files := []library.GameFile{
		{URL: "https://example.test/old", Label: "Build 1.0", Filename: name("game-1.0.zip")},
		{URL: "https://example.test/new", Label: "Build 1.2", Filename: name("game-1.2.zip")},
		{URL: "https://example.test/login", Label: "Beta", RequiresAuth: 1},
	}

	file, ok := pickUpdateFile(files, "1.2")
	if !ok || file.URL != "https://example.test/new" {
		t.Fatalf("picked %+v, %v; want the 1.2 file", file, ok)
	}
	if _, ok := pickUpdateFile(files, "9.9"); ok {
		t.Fatal("picked a file for a version no file names")
	}
	single := []library.GameFile{{URL: "https://example.test/only", Label: "Main"}}
	if file, ok := pickUpdateFile(single, "9.9"); !ok || file.URL != "https://example.test/only" {
		t.Fatalf("single open file was not picked: %+v, %v", file, ok)
	}
}
