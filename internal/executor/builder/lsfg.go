package builder

func (builder *CommandBuilder) applyLsfg() {
	if !builder.Options.Extras.Lsfg.Enabled {
		return
	}
	// Currently LSFG logic is handled by the launcher environment, 
	// but we can add specific LSFG env vars here if needed in the future.
}
