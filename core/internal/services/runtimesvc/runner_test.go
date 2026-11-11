package runtimesvc

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"light-launcher/core/internal/platform"
	"light-launcher/core/internal/types"
)

func linuxLaunch() types.LaunchOptions {
	return types.LaunchOptions{
		Name:       "Game",
		GamePath:   "/games/Game.exe",
		PrefixPath: "/prefixes/Game",
		ProtonPath: "/proton/GE-Proton9",
		Extras: types.ExtrasConfig{
			EnableMangoHud: true,
			Memory:         types.MemoryConfig{Enabled: true, Value: "8G"},
			Gamescope: types.GamescopeConfig{
				Enabled:    true,
				Width:      "1280",
				HDR:        true,
				CustomArgs: "--x",
			},
			Lsfg: types.LsfgConfig{
				Enabled:    true,
				Multiplier: "3",
				PerfMode:   true,
				FlowScale:  "1.5",
				AllowFp16:  true,
			},
		},
	}
}

func TestBuildInstanceArgsOnLinux(t *testing.T) {
	skipUnlessLinux(t)
	t.Setenv("LIGHT_LAUNCHER_DISCORD_CLIENT_ID", "")
	settings := types.AppSettings{DiscordClientId: "123"}

	got := buildInstanceArgs(linuxLaunch(), false, platform.Current(), settings)

	want := []string{
		"--game", "/games/Game.exe",
		"--launcher", "/games/Game.exe",
		"--prefix", "/prefixes/Game",
		"--proton-pattern", "GE-Proton9",
		"--proton-path", "/proton/GE-Proton9",
		"--game-name", "Game",
		"--discord-client-id", "123",
		"--logs=false",
		"--mangohud",
		"--memory-min", "--memory-min-value", "8G",
		"--gamescope", "--gs-w", "1280", "--gs-hdr", "--gs-custom-args", "--x",
		"--lsfg", "--lsfg-multiplier", "3", "--lsfg-perf", "--lsfg-flow", "1.5", "--lsfg-fp16",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("args =\n%v\nwant\n%v", got, want)
	}
}

func TestBuildInstanceArgsWithoutToolsSkipsToolFlags(t *testing.T) {
	t.Setenv("LIGHT_LAUNCHER_DISCORD_CLIENT_ID", "")
	launch := linuxLaunch()

	got := buildInstanceArgs(launch, true, platform.Features{}, types.AppSettings{})

	want := []string{
		"--game", "/games/Game.exe",
		"--launcher", "/games/Game.exe",
		"--game-name", "Game",
		"--memory-min", "--memory-min-value", "8G",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("args =\n%v\nwant\n%v", got, want)
	}
}

func TestDiscordArgsRespectRichPresenceSetting(t *testing.T) {
	t.Setenv("LIGHT_LAUNCHER_DISCORD_CLIENT_ID", "env-id")
	off := false

	if got := discordArgs(types.AppSettings{DiscordRichPresence: &off, DiscordClientId: "123"}); got != nil {
		t.Errorf("rich presence off gave %v, want none", got)
	}
	if got := discordArgs(types.AppSettings{}); !reflect.DeepEqual(got, []string{"--discord-client-id", "env-id"}) {
		t.Errorf("env fallback gave %v", got)
	}
}

func TestParseLaunchMultiplier(t *testing.T) {
	tests := map[string]int{"": 2, "4": 4, "x": 2}
	for input, want := range tests {
		if got := parseLaunchMultiplier(input); got != want {
			t.Errorf("parseLaunchMultiplier(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestResolveProtonPathFallsBackToFirstInstall(t *testing.T) {
	skipUnlessLinux(t)
	home := isolateHome(t)
	protonDir := filepath.Join(home, ".steam", "root", "compatibilitytools.d", "GE-Proton9")
	writeFile(t, filepath.Join(protonDir, "proton"), "")

	launch := types.LaunchOptions{UseCustomProton: true, ProtonPath: "/missing/Other"}
	if err := resolveProtonPath(&launch); err != nil {
		t.Fatalf("resolveProtonPath: %v", err)
	}
	if launch.ProtonPath != protonDir {
		t.Errorf("ProtonPath = %q, want the first installed tool %q", launch.ProtonPath, protonDir)
	}
}

func TestResolveProtonPathWithoutInstallsFails(t *testing.T) {
	isolateHome(t)

	launch := types.LaunchOptions{}
	if err := resolveProtonPath(&launch); !errors.Is(err, errNoProtonRuntime) {
		t.Errorf("error = %v, want no Proton runtime error", err)
	}
}

func TestResolvePrefixPathMovesLegacyPrefix(t *testing.T) {
	base := isolateConfigHome(t)
	launch := types.LaunchOptions{PrefixPath: "/old/home/LightLauncher/prefixes/Game"}

	if err := resolvePrefixPath(&launch); err != nil {
		t.Fatalf("resolvePrefixPath: %v", err)
	}
	want := filepath.Join(base, "Game")
	if launch.PrefixPath != want {
		t.Errorf("PrefixPath = %q, want %q", launch.PrefixPath, want)
	}
	if !pathExists(want) {
		t.Error("prefix folder was not created")
	}
}

func TestRunGameReportsMissingExecutable(t *testing.T) {
	isolateConfigHome(t)
	missing := filepath.Join(t.TempDir(), "missing.exe")

	err := runGame(types.LaunchOptions{GamePath: missing}, true)
	if err == nil || err.Error() != "Game executable not found at: "+missing {
		t.Errorf("error = %v, want missing executable error", err)
	}
}
