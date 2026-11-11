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

const blinkFeaturesPrefix = "--enable-blink-features="

// blinkFeatures are Blink features that are off by default on Linux. MiddleClickAutoscroll gives the
// Chrome-style middle-click drag scrolling that Windows browsers have.
var blinkFeatures = []string{"MiddleClickAutoscroll"}

// chromiumFlags adds the rendering flags to the ones the user already set. A flag that is already
// there is not added twice, and the user's own flags always stay.
func chromiumFlags(existing string, disableGPU bool) string {
	wanted := gpuFlags
	if disableGPU {
		wanted = []string{"--disable-gpu"}
	}
	flags := addMissing(strings.Fields(existing), wanted)
	flags = mergeBlinkFeatures(flags, blinkFeatures)
	return strings.Join(flags, " ")
}

// addMissing appends the wanted flags that are not in flags yet.
func addMissing(flags, wanted []string) []string {
	present := map[string]bool{}
	for _, flag := range flags {
		present[flag] = true
	}
	for _, flag := range wanted {
		if !present[flag] {
			flags = append(flags, flag)
		}
	}
	return flags
}

// mergeBlinkFeatures puts the features into one --enable-blink-features flag, keeping any the user listed.
func mergeBlinkFeatures(flags, features []string) []string {
	for index, flag := range flags {
		if !strings.HasPrefix(flag, blinkFeaturesPrefix) {
			continue
		}
		listed := strings.Split(strings.TrimPrefix(flag, blinkFeaturesPrefix), ",")
		flags[index] = blinkFeaturesPrefix + strings.Join(addMissing(listed, features), ",")
		return flags
	}
	return append(flags, blinkFeaturesPrefix+strings.Join(features, ","))
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
