//go:build windows

package main

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/jchv/go-webview2"
)

// bridgeScript defines window.llShell and sends outside links to the browser. It runs before each page.
//
//go:embed bridge.js
var bridgeScript string

// bindNativeFunctions exposes llRequestPick and llOpenExternal to the page.
func bindNativeFunctions(w webview2.WebView) error {
	if err := w.Bind("llRequestPick", func(method string, id int) {
		go pickInBackground(w, method, id)
	}); err != nil {
		return err
	}
	return w.Bind("llOpenExternal", openInBrowser)
}

// pickInBackground runs a dialog off the UI thread, then settles the page promise on the UI thread.
func pickInBackground(w webview2.WebView, method string, id int) {
	script, ok := pickerScripts[method]
	if !ok {
		log.Warn("bridge", "unknown method %q", method)
		return
	}
	path, err := showPicker(script)
	if err != nil {
		log.Error("bridge", "%s failed: %v", method, err)
	}
	w.Dispatch(func() {
		w.Eval(fmt.Sprintf("window.llShell._resolve(%d, %s)", id, jsValue(path)))
	})
}

// jsValue returns a JavaScript literal for a path. A cancel is null.
func jsValue(path string) string {
	if path == "" {
		return "null"
	}
	literal, err := json.Marshal(path)
	if err != nil {
		return "null"
	}
	return string(literal)
}
