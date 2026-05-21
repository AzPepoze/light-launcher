package gamescope

import (
	"flag"
	"light-launcher/internal/system"
	"light-launcher/internal/types"
	"strings"
)

type GamescopeAdapter struct {
	enabled        bool
	width          string
	height         string
	outputWidth    string
	outputHeight   string
	refreshRate    string
	framerateLimit string
	windowMode     string
	scaler         string
	filter         string
	sharpness      string
	hdr            bool
	adaptiveSync   bool
	mangoapp       bool
	customArgs     string
}

func NewAdapter() *GamescopeAdapter {
	return &GamescopeAdapter{}
}

func (a *GamescopeAdapter) Name() string {
	return "gamescope"
}

func (a *GamescopeAdapter) BuildInstanceArgs(options types.LaunchOptions) []string {
	var args []string
	if options.Extras.Gamescope.Enabled {
		args = append(args, "--gamescope")
		gs := options.Extras.Gamescope
		if gs.Width != "" {
			args = append(args, "--gs-w", gs.Width)
		}
		if gs.Height != "" {
			args = append(args, "--gs-h", gs.Height)
		}
		if gs.OutputWidth != "" {
			args = append(args, "--gs-out-w", gs.OutputWidth)
		}
		if gs.OutputHeight != "" {
			args = append(args, "--gs-out-h", gs.OutputHeight)
		}
		if gs.RefreshRate != "" {
			args = append(args, "--gs-r", gs.RefreshRate)
		}
		if gs.FramerateLimit != "" {
			args = append(args, "--gs-fr-limit", gs.FramerateLimit)
		}
		if gs.WindowMode != "" {
			args = append(args, "--gs-window-mode", gs.WindowMode)
		}
		if gs.Scaler != "" {
			args = append(args, "--gs-scaler", gs.Scaler)
		}
		if gs.Filter != "" {
			args = append(args, "--gs-filter", gs.Filter)
		}
		if gs.Sharpness != "" {
			args = append(args, "--gs-sharpness", gs.Sharpness)
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
	}
	return args
}

func (a *GamescopeAdapter) ModifyEnv(options types.LaunchOptions, env []string) []string {
	return env
}

func (a *GamescopeAdapter) WrapCommand(options types.LaunchOptions, args []string) []string {
	if options.Extras.Gamescope.Enabled && system.IsCommandAvailable("gamescope") {
		wrapper := []string{"gamescope"}
		gs := options.Extras.Gamescope
		if gs.Width != "" {
			wrapper = append(wrapper, "-w", gs.Width)
		}
		if gs.Height != "" {
			wrapper = append(wrapper, "-h", gs.Height)
		}
		if gs.OutputWidth != "" {
			wrapper = append(wrapper, "-W", gs.OutputWidth)
		}
		if gs.OutputHeight != "" {
			wrapper = append(wrapper, "-H", gs.OutputHeight)
		}
		if gs.RefreshRate != "" {
			wrapper = append(wrapper, "-r", gs.RefreshRate)
		}
		if gs.FramerateLimit != "" {
			wrapper = append(wrapper, "--framerate-limit", gs.FramerateLimit)
		}
		switch gs.WindowMode {
		case "fullscreen":
			wrapper = append(wrapper, "-f")
		case "borderless":
			wrapper = append(wrapper, "-b")
		}
		if gs.Scaler != "" && gs.Scaler != "auto" {
			wrapper = append(wrapper, "-S", gs.Scaler)
		}
		if gs.Filter != "" {
			wrapper = append(wrapper, "-F", gs.Filter)
		}
		if gs.Sharpness != "" {
			wrapper = append(wrapper, "--sharpness", gs.Sharpness)
		}
		if gs.HDR {
			wrapper = append(wrapper, "--hdr-enabled")
		}
		if gs.AdaptiveSync {
			wrapper = append(wrapper, "--adaptive-sync")
		}
		if gs.Mangoapp {
			wrapper = append(wrapper, "--mangoapp")
		}
		if gs.CustomArgs != "" {
			customTokens := strings.Fields(gs.CustomArgs)
			wrapper = append(wrapper, customTokens...)
		}
		wrapper = append(wrapper, "--", "env")
		return append(wrapper, args...)
	}
	return args
}

func (a *GamescopeAdapter) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&a.enabled, "gamescope", false, "Enable Gamescope")
	fs.StringVar(&a.width, "gs-w", "", "Width")
	fs.StringVar(&a.height, "gs-h", "", "Height")
	fs.StringVar(&a.outputWidth, "gs-out-w", "", "Output Width")
	fs.StringVar(&a.outputHeight, "gs-out-h", "", "Output Height")
	fs.StringVar(&a.refreshRate, "gs-r", "", "Refresh Rate")
	fs.StringVar(&a.framerateLimit, "gs-fr-limit", "", "Framerate Limit")
	fs.StringVar(&a.windowMode, "gs-window-mode", "", "Window Mode")
	fs.StringVar(&a.scaler, "gs-scaler", "", "Scaler")
	fs.StringVar(&a.filter, "gs-filter", "", "Filter")
	fs.StringVar(&a.sharpness, "gs-sharpness", "", "Sharpness")
	fs.BoolVar(&a.hdr, "gs-hdr", false, "Enable HDR")
	fs.BoolVar(&a.adaptiveSync, "gs-adaptive-sync", false, "Enable Adaptive Sync")
	fs.BoolVar(&a.mangoapp, "gs-mangoapp", false, "Enable Mangoapp")
	fs.StringVar(&a.customArgs, "gs-custom-args", "", "Custom Gamescope Arguments")
}

func (a *GamescopeAdapter) ExtractOptions(options *types.LaunchOptions) {
	options.Extras.Gamescope.Enabled = a.enabled
	options.Extras.Gamescope.Width = a.width
	options.Extras.Gamescope.Height = a.height
	options.Extras.Gamescope.OutputWidth = a.outputWidth
	options.Extras.Gamescope.OutputHeight = a.outputHeight
	options.Extras.Gamescope.RefreshRate = a.refreshRate
	options.Extras.Gamescope.FramerateLimit = a.framerateLimit
	options.Extras.Gamescope.WindowMode = a.windowMode
	options.Extras.Gamescope.Scaler = a.scaler
	options.Extras.Gamescope.Filter = a.filter
	options.Extras.Gamescope.Sharpness = a.sharpness
	options.Extras.Gamescope.HDR = a.hdr
	options.Extras.Gamescope.AdaptiveSync = a.adaptiveSync
	options.Extras.Gamescope.Mangoapp = a.mangoapp
	options.Extras.Gamescope.CustomArgs = a.customArgs
}
