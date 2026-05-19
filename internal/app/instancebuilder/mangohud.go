package instancebuilder

func (b *Builder) applyMangoHud() {
	if b.options.Extras.EnableMangoHud {
		b.args = append(b.args, "--mango")
	}
}
