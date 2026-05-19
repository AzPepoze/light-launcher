package builder

func (builder *CommandBuilder) addUmuRun() {
	builder.Arguments = append(builder.Arguments, "umu-run")

	executablePath := builder.Options.LauncherPath
	if executablePath == "" {
		executablePath = builder.Options.GamePath
	}
	builder.Arguments = append(builder.Arguments, executablePath)
}
