package instancebuilder

func (b *Builder) applyGamemode() {
	if b.options.Extras.EnableGamemode {
		b.args = append(b.args, "--gamemode")
	}
}
