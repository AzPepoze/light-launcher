package main

import (
	"os"
	"path/filepath"

	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/webengine"
)

const (
	windowTitle   = "LightLauncher"
	minWidth      = 800
	minHeight     = 600
	defaultWidth  = 1100
	defaultHeight = 720
)

// shellWindow owns the Qt view and page. Keep it referenced until the event loop ends.
type shellWindow struct {
	view         *webengine.QWebEngineView
	page         *webengine.QWebEnginePage
	geometryFile string
}

func newWindow(opts options) *shellWindow {
	view := webengine.NewQWebEngineView(nil)
	// miqt only allows overriding page methods on a page the shell creates itself.
	page := webengine.NewQWebEnginePage()
	view.SetPage(page)
	enableSmoothScrolling(page)
	view.SetWindowTitle(windowTitle)
	view.SetMinimumSize2(minWidth, minHeight)
	view.Resize(defaultWidth, defaultHeight)

	win := &shellWindow{view: view, page: page, geometryFile: geometryPath()}
	win.restoreGeometry()
	if opts.transparent {
		makeTransparent(view, page)
	}
	view.OnCloseEvent(func(super func(event *qt.QCloseEvent), event *qt.QCloseEvent) {
		win.saveGeometry()
		super(event)
	})

	launcherOrigin := origin(opts.url)
	routeExternalLinks(page, launcherOrigin)
	installBridge(page, view.QWidget, launcherOrigin)
	return win
}

func makeTransparent(view *webengine.QWebEngineView, page *webengine.QWebEnginePage) {
	view.SetAttribute2(qt.WA_TranslucentBackground, true)
	page.SetBackgroundColor(qt.NewQColor2(qt.Transparent))
}

// geometryPath returns the file that keeps the window position and size between runs.
func geometryPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "light-launcher", "window-geometry")
}

func (win *shellWindow) restoreGeometry() {
	if win.geometryFile == "" {
		return
	}
	data, err := os.ReadFile(win.geometryFile)
	if err != nil {
		return
	}
	win.view.RestoreGeometry(data)
}

func (win *shellWindow) saveGeometry() {
	if win.geometryFile == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(win.geometryFile), 0o755); err != nil {
		log.Warn("window", "cannot create config folder: %v", err)
		return
	}
	if err := os.WriteFile(win.geometryFile, win.view.SaveGeometry(), 0o600); err != nil {
		log.Warn("window", "cannot save geometry: %v", err)
	}
}
