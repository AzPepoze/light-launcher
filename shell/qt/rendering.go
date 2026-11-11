package main

import (
	"os"
	"strings"

	"github.com/mappu/miqt/qt6/webengine"
)

const (
	chromiumFlagsEnv = "QTWEBENGINE_CHROMIUM_FLAGS"
	disableGPUEnv    = "LL_DISABLE_GPU"
	renderLogName    = "render"
)

// gpuFlags keep the page on the GPU. The blocklist can silently drop a Linux driver to software
// rendering, which makes scrolling and animation choppy.
var gpuFlags = []string{"--ignore-gpu-blocklist", "--enable-gpu-rasterization"}

// chromiumFlags adds the rendering flags to the ones the user already set. A flag that is already
// there is not added twice, and the user's own flags always stay.
func chromiumFlags(existing string, disableGPU bool) string {
	wanted := gpuFlags
	if disableGPU {
		wanted = []string{"--disable-gpu"}
	}
	present := map[string]bool{}
	for _, flag := range strings.Fields(existing) {
		present[flag] = true
	}
	flags := strings.Fields(existing)
	for _, flag := range wanted {
		if !present[flag] {
			flags = append(flags, flag)
		}
	}
	return strings.Join(flags, " ")
}

// configureRendering must run before Qt starts, because Chromium reads its flags once.
func configureRendering() {
	disable := os.Getenv(disableGPUEnv) == "1"
	if err := os.Setenv(chromiumFlagsEnv, chromiumFlags(os.Getenv(chromiumFlagsEnv), disable)); err != nil {
		log.Warn(renderLogName, "cannot set the rendering flags: %v", err)
		return
	}
	if disable {
		log.Warn(renderLogName, "GPU is off because %s=1", disableGPUEnv)
	}
}

// enableSmoothScrolling turns on animated scrolling, which Qt WebEngine leaves off by default.
func enableSmoothScrolling(page *webengine.QWebEnginePage) {
	page.Settings().SetAttribute(webengine.QWebEngineSettings__ScrollAnimatorEnabled, true)
}
