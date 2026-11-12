//go:build linux

// Command light-launcher-shell shows the launcher UI in a native GTK4 WebKitGTK window.
//
// The launcher passes the URL in the LL_URL environment variable. A URL argument is the fallback.
// Usage: [LL_URL=<url>] light-launcher-shell [--transparent] [url]
//
// The --assist mode instead shows one web page for a human to use and reports the download the page
// starts. It never acts on the page: the user logs in, presses the buttons and solves any challenge.
package main

import (
	"os"
	"runtime"

	"light-launcher/shell/common/logx"
)

var log = logx.New(os.Stderr)

func main() {
	// GTK and WebKit must have their calls on the thread that ran gtk_init.
	runtime.LockOSThread()
	logRendering()
	args := os.Args[1:]
	if wantsAssist(args) {
		os.Exit(runAssist(args))
	}

	opts, err := parseOptions(args)
	if err != nil {
		log.Error("args", "%v", err)
		log.Error("args", "usage: LL_URL=<url> light-launcher-shell [--transparent]")
		os.Exit(exitBadArgs)
	}
	os.Exit(runNormal(opts))
}

func runNormal(opts options) int {
	session := newNormalSession(opts)
	log.Info("window", "opened %s", session.launcherOrigin)
	path := geometryPath()
	geometry := defaultGeometry()
	if saved, ok := loadGeometry(path); ok {
		geometry = usableGeometry(saved, screenAreas())
	}
	config := gtkConfig{
		url:          opts.url.String(),
		title:        windowTitle,
		bridgeScript: bridgeScript,
		transparent:  opts.transparent,
		devtools:     developerTools(),
		disableGPU:   gpuDisabled(),
		geometry:     geometry,
	}
	if _, err := runGTK(session, config); err != nil {
		log.Error("window", "%v", err)
		return exitFailure
	}
	if session.geometrySet {
		if err := saveGeometry(path, session.geometry); err != nil {
			log.Warn("window", "cannot save the window geometry: %v", err)
		}
	}
	return exitOK
}
