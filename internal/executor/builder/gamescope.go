package builder

import "light-launcher/internal/system"

func (builder *CommandBuilder) applyGamescope() {
	if builder.Options.Extras.Gamescope.Enabled && system.IsCommandAvailable("gamescope") {
		builder.Arguments = append(builder.Arguments, "gamescope")
		if builder.Options.Extras.Gamescope.Width != "" {
			builder.Arguments = append(builder.Arguments, "-w", builder.Options.Extras.Gamescope.Width)
		}
		if builder.Options.Extras.Gamescope.Height != "" {
			builder.Arguments = append(builder.Arguments, "-h", builder.Options.Extras.Gamescope.Height)
		}
		if builder.Options.Extras.Gamescope.RefreshRate != "" {
			builder.Arguments = append(builder.Arguments, "-r", builder.Options.Extras.Gamescope.RefreshRate)
		}
		builder.Arguments = append(builder.Arguments, "--", "env")
	}
}
