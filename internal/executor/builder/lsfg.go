package builder

func (builder *CommandBuilder) applyLsfg() {
	if !builder.Options.Extras.Lsfg.Enabled {
		builder.Environment = append(builder.Environment, "DISABLE_LSFGVK=1")
		return
	}
}
