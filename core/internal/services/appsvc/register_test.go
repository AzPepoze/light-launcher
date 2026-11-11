package appsvc

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"light-launcher/core/internal/api"
)

var registeredMethods = []string{
	"GetInitialLauncherPath", "GetInitialGamePath", "GetShouldEditLsfg",
	"IsDir", "GetExeIcon", "GetTotalRam", "GetImageBase64",
	"OpenURL", "OpenFileLocation", "CloseWindow", "RestartApp",
	"GetAppSettings", "SaveAppSettings", "GetConfig", "SaveGameConfig",
	"GetAllGames", "RemoveGame", "SearchExecutables", "GetAutoScannedGames",
	"AddScanFolder", "RemoveScanFolder", "UpdateScanFolderConfig", "RenameScanFolder",
	"GetScanFolderConfig", "SetPreferredLauncher", "ClearPreferredLauncher", "BlacklistGame", "UnblacklistGame",
	"GetGameActivity", "PickFile", "PickFolder", "PickFileCustom",
}

type recordingEvents struct {
	channels []string
}

func (e *recordingEvents) Emit(channel string, _ any) {
	e.channels = append(e.channels, channel)
}

func TestRegisterAddsExactlyTheUIMethods(t *testing.T) {
	router := api.NewRouter()
	Register(router, &recordingEvents{})

	got := router.Methods()
	slices.Sort(got)
	want := slices.Clone(registeredMethods)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("registered methods differ\n got: %v\nwant: %v", got, want)
	}
}

func TestRegisteredHandlersDecodeRouterPayloads(t *testing.T) {
	isolateConfigHome(t)
	router := api.NewRouter()
	events := &recordingEvents{}
	Register(router, events)
	ctx := context.Background()

	folder := t.TempDir()
	payload, err := json.Marshal(map[string]string{"path": folder})
	if err != nil {
		t.Fatal(err)
	}
	isDir, err := router.Dispatch(ctx, "IsDir", payload)
	if err != nil || isDir != true {
		t.Fatalf("IsDir = %v, %v", isDir, err)
	}

	settings, err := router.Dispatch(ctx, "GetAppSettings", nil)
	if err != nil {
		t.Fatalf("GetAppSettings: %v", err)
	}
	encoded, err := json.Marshal(settings)
	if err != nil || !json.Valid(encoded) {
		t.Fatalf("GetAppSettings is not JSON: %v", err)
	}

	if _, err := router.Dispatch(ctx, "CloseWindow", nil); err != nil {
		t.Fatalf("CloseWindow: %v", err)
	}
	if _, err := router.Dispatch(ctx, "RestartApp", nil); err != nil {
		t.Fatalf("RestartApp: %v", err)
	}
	if !slices.Equal(events.channels, []string{"window-close", "restart-app"}) {
		t.Fatalf("events = %v", events.channels)
	}
}

func TestSetInitialArgsOverridesEnvironment(t *testing.T) {
	t.Setenv("LIGHT_LAUNCHER_GAME_PATH", "/from/env.exe")
	SetInitialArgs("/launcher.exe", "/game.exe", true)
	t.Cleanup(func() { SetInitialArgs("", "", false) })

	if got := initialGamePath(); got != "/game.exe" {
		t.Errorf("initialGamePath = %q", got)
	}
	if !shouldEditLsfg() {
		t.Error("shouldEditLsfg must be true when set")
	}
}
