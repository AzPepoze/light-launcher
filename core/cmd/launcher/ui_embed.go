//go:build embedui

package main

import (
	"embed"
	"io/fs"
)

// The release build copies src/renderer/dist into core/cmd/launcher/ui, then builds with -tags embedui.
//
//go:embed all:ui
var embeddedFiles embed.FS

func embeddedUI() (fs.FS, bool) {
	ui, err := fs.Sub(embeddedFiles, "ui")
	return ui, err == nil
}
