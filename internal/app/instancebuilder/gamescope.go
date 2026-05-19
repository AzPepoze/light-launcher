package instancebuilder

func (b *Builder) applyGamescope() {
	if b.options.Extras.Gamescope.Enabled {
		b.args = append(b.args, "--gamescope")
		if b.options.Extras.Gamescope.Width != "" {
			b.args = append(b.args, "--gs-w", b.options.Extras.Gamescope.Width)
		}
		if b.options.Extras.Gamescope.Height != "" {
			b.args = append(b.args, "--gs-h", b.options.Extras.Gamescope.Height)
		}
		if b.options.Extras.Gamescope.OutputWidth != "" {
			b.args = append(b.args, "--gs-out-w", b.options.Extras.Gamescope.OutputWidth)
		}
		if b.options.Extras.Gamescope.OutputHeight != "" {
			b.args = append(b.args, "--gs-out-h", b.options.Extras.Gamescope.OutputHeight)
		}
		if b.options.Extras.Gamescope.RefreshRate != "" {
			b.args = append(b.args, "--gs-r", b.options.Extras.Gamescope.RefreshRate)
		}
		if b.options.Extras.Gamescope.FramerateLimit != "" {
			b.args = append(b.args, "--gs-fr-limit", b.options.Extras.Gamescope.FramerateLimit)
		}
		if b.options.Extras.Gamescope.WindowMode != "" {
			b.args = append(b.args, "--gs-window-mode", b.options.Extras.Gamescope.WindowMode)
		}
		if b.options.Extras.Gamescope.Scaler != "" {
			b.args = append(b.args, "--gs-scaler", b.options.Extras.Gamescope.Scaler)
		}
		if b.options.Extras.Gamescope.Filter != "" {
			b.args = append(b.args, "--gs-filter", b.options.Extras.Gamescope.Filter)
		}
		if b.options.Extras.Gamescope.Sharpness != "" {
			b.args = append(b.args, "--gs-sharpness", b.options.Extras.Gamescope.Sharpness)
		}
		if b.options.Extras.Gamescope.HDR {
			b.args = append(b.args, "--gs-hdr")
		}
		if b.options.Extras.Gamescope.AdaptiveSync {
			b.args = append(b.args, "--gs-adaptive-sync")
		}
		if b.options.Extras.Gamescope.Mangoapp {
			b.args = append(b.args, "--gs-mangoapp")
		}
		if b.options.Extras.Gamescope.CustomArgs != "" {
			b.args = append(b.args, "--gs-custom-args", b.options.Extras.Gamescope.CustomArgs)
		}
	}
}
