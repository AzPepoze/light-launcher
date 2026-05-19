package instancebuilder

import (
	"path/filepath"
	"light-launcher/internal/types"
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

	b.applyMangoHud()
	b.applyGamemode()
	b.applyLsfg()
	b.applyMemory()
	b.applyGamescope()

	if !b.showLogs {
		b.args = append(b.args, "--logs=false")
	}

	return b.args
}


