package builder

import (
	"strings"

	"light-launcher/internal/system"
)

func (builder *CommandBuilder) applyGamescope() {
	if builder.Options.Extras.Gamescope.Enabled && system.IsCommandAvailable("gamescope") {
		builder.Arguments = append(builder.Arguments, "gamescope")
		gs := builder.Options.Extras.Gamescope

		if gs.Width != "" {
			builder.Arguments = append(builder.Arguments, "-w", gs.Width)
		}
		if gs.Height != "" {
			builder.Arguments = append(builder.Arguments, "-h", gs.Height)
		}
		if gs.OutputWidth != "" {
			builder.Arguments = append(builder.Arguments, "-W", gs.OutputWidth)
		}
		if gs.OutputHeight != "" {
			builder.Arguments = append(builder.Arguments, "-H", gs.OutputHeight)
		}
		if gs.RefreshRate != "" {
			builder.Arguments = append(builder.Arguments, "-r", gs.RefreshRate)
		}
		if gs.FramerateLimit != "" {
			builder.Arguments = append(builder.Arguments, "--framerate-limit", gs.FramerateLimit)
		}
		if gs.WindowMode == "fullscreen" {
			builder.Arguments = append(builder.Arguments, "-f")
		} else if gs.WindowMode == "borderless" {
			builder.Arguments = append(builder.Arguments, "-b")
		}
		if gs.Scaler != "" && gs.Scaler != "auto" {
			builder.Arguments = append(builder.Arguments, "-S", gs.Scaler)
		}
		if gs.Filter != "" {
			builder.Arguments = append(builder.Arguments, "-F", gs.Filter)
		}
		if gs.Sharpness != "" {
			builder.Arguments = append(builder.Arguments, "--sharpness", gs.Sharpness)
		}
		if gs.HDR {
			builder.Arguments = append(builder.Arguments, "--hdr-enabled")
		}
		if gs.AdaptiveSync {
			builder.Arguments = append(builder.Arguments, "--adaptive-sync")
		}
		if gs.Mangoapp {
			builder.Arguments = append(builder.Arguments, "--mangoapp")
		}
		if gs.CustomArgs != "" {
			customTokens := strings.Fields(gs.CustomArgs)
			builder.Arguments = append(builder.Arguments, customTokens...)
		}

		builder.Arguments = append(builder.Arguments, "--", "env")
	}
}
