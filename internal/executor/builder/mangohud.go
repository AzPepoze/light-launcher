package builder

func (builder *CommandBuilder) applyMangoHud() {
	if builder.Options.Extras.EnableMangoHud {
		builder.Environment = append(builder.Environment, "MANGOHUD=1")
	}
}
