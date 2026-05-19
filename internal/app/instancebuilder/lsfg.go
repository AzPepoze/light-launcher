package instancebuilder

func (b *Builder) applyLsfg() {
	if b.options.Extras.Lsfg.Enabled {
		b.args = append(b.args, "--lsfg", "--lsfg-mult", b.options.Extras.Lsfg.Multiplier)
		if b.options.Extras.Lsfg.PerfMode {
			b.args = append(b.args, "--lsfg-perf")
		}
		if b.options.Extras.Lsfg.DllPath != "" {
			b.args = append(b.args, "--lsfg-dll-path", b.options.Extras.Lsfg.DllPath)
		}
	}
}
