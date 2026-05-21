package instancebuilder

import (
	"light-launcher/internal/adapter"
	"light-launcher/internal/types"
	"path/filepath"
)

type Builder struct {
	options  types.LaunchOptions
	showLogs bool
	args     []string
}

func NewBuilder(options types.LaunchOptions, showLogs bool) *Builder {
	return &Builder{
		options:  options,
		showLogs: showLogs,
		args:     []string{},
	}
}

func (b *Builder) Build() []string {
	b.args = append(b.args,
		"--game", b.options.GamePath,
		"--launcher", b.options.LauncherPath,
		"--prefix", b.options.PrefixPath,
		"--proton-pattern", filepath.Base(b.options.ProtonPath),
		"--proton-path", b.options.ProtonPath,
	)

	// Build using adapters
	for _, a := range adapter.GetAdapters() {
		b.args = append(b.args, a.BuildInstanceArgs(b.options)...)
	}

	if !b.showLogs {
		b.args = append(b.args, "--logs=false")
	}

	return b.args
}
