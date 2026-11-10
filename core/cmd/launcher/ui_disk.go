//go:build !embedui

package main

import "io/fs"

// embeddedUI reports no embedded UI: normal builds read the UI from disk.
func embeddedUI() (fs.FS, bool) {
	return nil, false
}
