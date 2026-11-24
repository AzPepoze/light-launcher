package builder

import (
	"light-launcher/core/internal/platform"
	"light-launcher/core/internal/types"
	"path/filepath"
	"testing"
)

func TestLaunchExecutablePrefersLauncher(t *testing.T) {
	options := types.LaunchOptions{GamePath: "/games/GTA5.exe", LauncherPath: "/games/PlayGTAV.exe"}
	if got := launchExecutable(options); got != "/games/PlayGTAV.exe" {
		t.Fatalf("launchExecutable() = %q, want launcher path", got)
	}
}

func TestLaunchExecutableFallsBackToGame(t *testing.T) {
	options := types.LaunchOptions{GamePath: "/games/GTA5.exe"}
	if got := launchExecutable(options); got != "/games/GTA5.exe" {
		t.Fatalf("launchExecutable() = %q, want game path", got)
	}
}

func TestWorkingDirectoryIsExecutableFolder(t *testing.T) {
	options := types.LaunchOptions{GamePath: "/games/GTAV/GTA5.exe", LauncherPath: "/games/GTAV/PlayGTAV.exe"}
	want := filepath.Dir(options.LauncherPath)
	if !platform.Current().Proton {
		want = filepath.Dir(options.GamePath)
	}
	if got := WorkingDirectory(options); got != want {
		t.Fatalf("WorkingDirectory() = %q, want %q", got, want)
	}
}

func TestProtonCommandStartsLauncher(t *testing.T) {
	if !platform.Current().Proton {
		t.Skip("Proton command is only used on Proton platforms")
	}
	options := types.LaunchOptions{GamePath: "/games/GTAV/GTA5.exe", LauncherPath: "/games/GTAV/PlayGTAV.exe"}
	arguments, _ := protonCommand(options)
	if len(arguments) < 2 || arguments[0] != "umu-run" || arguments[1] != "/games/GTAV/PlayGTAV.exe" {
		t.Fatalf("protonCommand() arguments = %v, want umu-run with launcher path", arguments)
	}
}

func TestUserEnvKeepsOnlyAssignments(t *testing.T) {
	options := types.LaunchOptions{EnvVars: "USE_FAKE_EPIC_EXE=true  stray DXVK_HUD=fps"}
	want := []string{"USE_FAKE_EPIC_EXE=true", "DXVK_HUD=fps"}
	got := userEnv(options)
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("userEnv() = %v, want %v", got, want)
	}
}
