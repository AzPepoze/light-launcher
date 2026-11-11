//go:build linux

package main

/*
#cgo pkg-config: gtk4 webkitgtk-6.0
#include <stdlib.h>
#include "shell.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime/cgo"
	"unsafe"
)

type gtkConfig struct {
	url          string
	title        string
	bridgeScript string
	profileDir   string
	message      string
	transparent  bool
	assist       bool
	devtools     bool
	disableGPU   bool
}

// runGTK opens the window on the main thread and blocks until it closes.
func runGTK(session *shellSession, config gtkConfig) (int, error) {
	handle := cgo.NewHandle(session)
	defer handle.Delete()

	url := C.CString(config.url)
	defer C.free(unsafe.Pointer(url))
	title := C.CString(config.title)
	defer C.free(unsafe.Pointer(title))
	bridge := optionalCString(config.bridgeScript)
	defer freeCString(bridge)
	profile := optionalCString(config.profileDir)
	defer freeCString(profile)
	banner := optionalCString(config.message)
	defer freeCString(banner)

	code := int(C.ll_shell_run(
		C.uintptr_t(handle),
		url,
		bridge,
		profile,
		title,
		banner,
		boolInt(config.transparent),
		boolInt(config.assist),
		boolInt(config.devtools),
		boolInt(config.disableGPU),
	))
	if code != 0 {
		return code, fmt.Errorf("GTK window failed to start (code %d)", code)
	}
	return 0, nil
}

func openPicker(id int, folder bool) {
	C.ll_shell_pick(C.int(id), boolInt(folder))
}

func evalJS(script string) {
	value := C.CString(script)
	defer C.free(unsafe.Pointer(value))
	C.ll_shell_eval(value)
}

func quitShell() {
	C.ll_shell_quit()
}

func openURI(rawURL string) error {
	value := C.CString(rawURL)
	defer C.free(unsafe.Pointer(value))
	if C.ll_open_uri(value) == 0 {
		return errors.New("no application accepted the address")
	}
	return nil
}

func optionalCString(value string) *C.char {
	if value == "" {
		return nil
	}
	return C.CString(value)
}

func freeCString(value *C.char) {
	if value != nil {
		C.free(unsafe.Pointer(value))
	}
}

func boolInt(value bool) C.int {
	if value {
		return 1
	}
	return 0
}
