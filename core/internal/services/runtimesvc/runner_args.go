package runtimesvc

import (
	"os"
	"path/filepath"
	"strings"

	"light-launcher/core/internal/platform"
	"light-launcher/core/internal/types"
)

// buildInstanceArgs lists the flags for light-launcher-instance. Tool flags are left out when the platform lacks the tool.
func buildInstanceArgs(launch types.LaunchOptions, showLogs bool, features platform.Features, settings types.AppSettings) []string {
	args := []string{
		"--game", launch.GamePath,
		"--launcher", firstNonEmpty(launch.LauncherPath, launch.GamePath),
	}
	if features.Proton {
		args = append(args,
			"--prefix", launch.PrefixPath,
			"--proton-pattern", nodeBasename(launch.ProtonPath),
			"--proton-path", expandPath(launch.ProtonPath),
		)
	}
	args = append(args,
		"--game-name", firstNonEmpty(launch.Name, nameWithoutExt(filepath.Base(launch.GamePath))),
	)
	args = append(args, discordArgs(settings)...)
	if !showLogs {
		args = append(args, "--logs=false")
	}
	return append(args, extrasArgs(launch.Extras, features)...)
}

// discordArgs passes the Discord client ID unless Rich Presence is turned off.
func discordArgs(settings types.AppSettings) []string {
	if settings.DiscordRichPresence != nil && !*settings.DiscordRichPresence {
		return nil
	}
	clientID := strings.TrimSpace(firstNonEmpty(settings.DiscordClientId, os.Getenv("LIGHT_LAUNCHER_DISCORD_CLIENT_ID")))
	if clientID == "" {
		return nil
	}
	return []string{"--discord-client-id", clientID}
}

func extrasArgs(extras types.ExtrasConfig, features platform.Features) []string {
	var args []string
	if features.MangoHud && extras.EnableMangoHud {
		args = append(args, "--mangohud")
	}
	if features.GameMode && extras.EnableGamemode {
		args = append(args, "--gamemode")
	}
	if extras.Memory.Enabled {
		args = append(args, "--memory-min")
		if extras.Memory.Value != "" {
			args = append(args, "--memory-min-value", extras.Memory.Value)
		}
	}
	if features.Gamescope && extras.Gamescope.Enabled {
		args = append(args, gamescopeArgs(extras.Gamescope)...)
	}
	if features.Lsfg && extras.Lsfg.Enabled {
		args = append(args, lsfgArgs(extras.Lsfg)...)
	}
	return args
}

func gamescopeArgs(gs types.GamescopeConfig) []string {
	args := []string{"--gamescope"}
	valued := []struct {
		flag  string
		value string
	}{
		{"--gs-w", gs.Width},
		{"--gs-h", gs.Height},
		{"--gs-out-w", gs.OutputWidth},
		{"--gs-out-h", gs.OutputHeight},
		{"--gs-r", gs.RefreshRate},
		{"--gs-fr-limit", gs.FramerateLimit},
		{"--gs-window-mode", gs.WindowMode},
		{"--gs-scaler", gs.Scaler},
		{"--gs-filter", gs.Filter},
		{"--gs-sharpness", gs.Sharpness},
	}
	for _, option := range valued {
		if option.value != "" {
			args = append(args, option.flag, option.value)
		}
	}
	if gs.HDR {
		args = append(args, "--gs-hdr")
	}
	if gs.AdaptiveSync {
		args = append(args, "--gs-adaptive-sync")
	}
	if gs.Mangoapp {
		args = append(args, "--gs-mangoapp")
	}
	if gs.CustomArgs != "" {
		args = append(args, "--gs-custom-args", gs.CustomArgs)
	}
	return args
}

func lsfgArgs(ls types.LsfgConfig) []string {
	args := []string{"--lsfg"}
	if ls.Multiplier != "" {
		args = append(args, "--lsfg-multiplier", ls.Multiplier)
	}
	if ls.PerfMode {
		args = append(args, "--lsfg-perf")
	}
	if ls.DllPath != "" {
		args = append(args, "--lsfg-dll", ls.DllPath)
	}
	if ls.Gpu != "" {
		args = append(args, "--lsfg-gpu", ls.Gpu)
	}
	if ls.FlowScale != "" {
		args = append(args, "--lsfg-flow", ls.FlowScale)
	}
	if ls.Pacing != "" {
		args = append(args, "--lsfg-pacing", ls.Pacing)
	}
	if ls.AllowFp16 {
		args = append(args, "--lsfg-fp16")
	}
	return args
}
