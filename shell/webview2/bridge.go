//go:build windows

package main

import (
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jchv/go-webview2"

	"light-launcher/shell/common/bridge"
)

// bridgeTemplate defines window.llShell and sends outside links to the browser. It runs
// before each page. bridgeScriptFor fills in the launcher origin and per-run secret.
//
//go:embed bridge.js
var bridgeTemplate string

// nativeActions are the shell actions a page may call through the bridge.
type nativeActions struct {
	pick func(method string, id int)
	open func(target string) error
}

// newBridgeSecret returns a random token that only the injected script holds, so a
// page cannot call the native bindings directly.
func newBridgeSecret() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate bridge secret: %w", err)
	}
	return hex.EncodeToString(raw), nil
}

// bridgeScriptFor fills the launcher origin and secret into the injected script.
func bridgeScriptFor(launcherOrigin, secret string) (string, error) {
	origin, err := json.Marshal(launcherOrigin)
	if err != nil {
		return "", fmt.Errorf("encode launcher origin: %w", err)
	}
	token, err := json.Marshal(secret)
	if err != nil {
		return "", fmt.Errorf("encode bridge secret: %w", err)
	}
	return strings.NewReplacer(
		"__LL_ORIGIN__", string(origin),
		"__LL_SECRET__", string(token),
	).Replace(bridgeTemplate), nil
}

// bindNativeFunctions exposes llRequestPick and llOpenExternal to the page. The
// injected script passes the per-run secret and the page origin with every call,
// and the native side accepts calls only from the launcher origin.
func bindNativeFunctions(w webview2.WebView, launcherOrigin, secret string, actions nativeActions) error {
	if err := w.Bind("llRequestPick", func(pageSecret, pageOrigin, method string, id int) {
		if !allowsCall(pageSecret, pageOrigin, launcherOrigin, secret) {
			log.Warn("bridge", "ignored a request from another origin")
			return
		}
		actions.pick(method, id)
	}); err != nil {
		return err
	}
	return w.Bind("llOpenExternal", func(pageSecret, pageOrigin, target string) error {
		if !allowsCall(pageSecret, pageOrigin, launcherOrigin, secret) {
			log.Warn("bridge", "ignored a link from another origin")
			return nil
		}
		return actions.open(target)
	})
}

func allowsCall(pageSecret, pageOrigin, launcherOrigin, secret string) bool {
	return pageSecret == secret && bridge.Allows(pageOrigin, launcherOrigin)
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
	resolve, err := bridge.ResolveScript(id, path)
	if err != nil {
		log.Error("bridge", "cannot build reply: %v", err)
		return
	}
	w.Dispatch(func() {
		w.Eval(resolve)
	})
}
