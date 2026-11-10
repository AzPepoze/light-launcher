package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/webengine"
)

// bridgeScript defines window.llShell. It is injected into every page the shell loads.
//
//go:embed bridge.js
var bridgeScript string

// bridgePrefix marks the console messages that bridge.js sends to the shell.
const bridgePrefix = "llshell:"

type bridgeRequest struct {
	ID     int    `json:"id"`
	Method string `json:"method"`
}

// installBridge answers window.llShell requests. Only pages on the launcher origin get answers.
//
// miqt cannot define custom QWebChannel objects, so the bridge uses console messages instead.
func installBridge(page *webengine.QWebEnginePage, parent *qt.QWidget, launcherOrigin string) {
	page.Profile().Scripts().Insert(bridgeScriptObject())

	page.OnJavaScriptConsoleMessage(func(super func(level webengine.QWebEnginePage__JavaScriptConsoleMessageLevel, message string, lineNumber int, sourceID string), level webengine.QWebEnginePage__JavaScriptConsoleMessageLevel, message string, lineNumber int, sourceID string) {
		req, ok := parseBridgeRequest(message)
		if !ok {
			super(level, message, lineNumber, sourceID)
			return
		}
		if qUrlOrigin(page.Url()) != launcherOrigin {
			log.Warn("bridge", "ignored %s from another origin", req.Method)
			return
		}
		answerLater(page, parent, req)
	})
}

func bridgeScriptObject() *webengine.QWebEngineScript {
	script := webengine.NewQWebEngineScript()
	script.SetName("llshell")
	script.SetSourceCode(bridgeScript)
	script.SetInjectionPoint(webengine.QWebEngineScript__DocumentCreation)
	script.SetWorldId(0) // 0 is the page's main world, the one the launcher UI runs in.
	script.SetRunsOnSubFrames(false)
	return script
}

func parseBridgeRequest(message string) (bridgeRequest, bool) {
	if !strings.HasPrefix(message, bridgePrefix) {
		return bridgeRequest{}, false
	}
	var req bridgeRequest
	if err := json.Unmarshal([]byte(strings.TrimPrefix(message, bridgePrefix)), &req); err != nil {
		return bridgeRequest{}, false
	}
	return req, true
}

// answerLater opens the dialog on the next event loop tick, outside the console message callback.
func answerLater(page *webengine.QWebEnginePage, parent *qt.QWidget, req bridgeRequest) {
	timer := qt.NewQTimer()
	timer.SetSingleShot(true)
	timer.OnTimeout(func() {
		answer(page, parent, req)
		timer.DeleteLater()
	})
	timer.Start(0)
}

func answer(page *webengine.QWebEnginePage, parent *qt.QWidget, req bridgeRequest) {
	path, ok := showPicker(parent, req.Method)
	if !ok {
		log.Warn("bridge", "unknown method %q", req.Method)
		return
	}
	script, err := resolveScript(req.ID, path)
	if err != nil {
		log.Error("bridge", "cannot build reply: %v", err)
		return
	}
	page.RunJavaScriptWithScriptSource(script)
}

// showPicker opens the native dialog for a method. An empty path means the user cancelled.
func showPicker(parent *qt.QWidget, method string) (string, bool) {
	switch method {
	case "pickFolder":
		return qt.QFileDialog_GetExistingDirectory3(parent, "Choose a folder", ""), true
	case "pickFile":
		return qt.QFileDialog_GetOpenFileName3(parent, "Choose a file", ""), true
	}
	return "", false
}

// resolveScript builds the JavaScript that settles one pending promise. A cancel resolves to null.
func resolveScript(id int, path string) (string, error) {
	var value any
	if path != "" {
		value = path
	}
	literal, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("window.llShell._resolve(%d, %s)", id, literal), nil
}
