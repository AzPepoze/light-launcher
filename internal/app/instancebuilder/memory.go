package instancebuilder

func (b *Builder) applyMemory() {
	if b.options.Extras.Memory.Enabled {
		b.args = append(b.args, "--memory-min")
		if b.options.Extras.Memory.Value != "" {
			b.args = append(b.args, "--memory-min-value", b.options.Extras.Memory.Value)
		}
	}
}
