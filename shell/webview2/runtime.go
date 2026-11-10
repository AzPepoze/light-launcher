//go:build windows

package main

import (
	"fmt"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// webView2ClientID is the Evergreen WebView2 Runtime product ID in the Edge update registry.
const (
	webView2ClientID    = `Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`
	webView2DownloadURL = "https://developer.microsoft.com/microsoft-edge/webview2/"
	// messageBoxYes is the MessageBox result for the Yes button (IDYES).
	messageBoxYes = 6
)

// webView2Installed reports whether the runtime is registered for the machine or the current user.
func webView2Installed() bool {
	keys := []struct {
		root registry.Key
		path string
	}{
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\` + webView2ClientID},
		{registry.LOCAL_MACHINE, `SOFTWARE\` + webView2ClientID},
		{registry.CURRENT_USER, `Software\` + webView2ClientID},
	}
	for _, key := range keys {
		if hasRuntimeVersion(key.root, key.path) {
			return true
		}
	}
	return false
}

// hasRuntimeVersion reads the "pv" value, which holds the installed version.
func hasRuntimeVersion(root registry.Key, path string) bool {
	key, err := registry.OpenKey(root, path, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()
	version, _, err := key.GetStringValue("pv")
	return err == nil && version != "" && version != "0.0.0.0"
}

// offerRuntimeDownload asks whether to open the WebView2 download page in the browser.
func offerRuntimeDownload() {
	text := fmt.Sprintf(
		"The Microsoft Edge WebView2 Runtime is missing.\n\n"+
			"LightLauncher needs it to show its window.\n\n"+
			"Download: %s\n\nOpen the download page now?",
		webView2DownloadURL,
	)
	choice, _ := windows.MessageBox(
		0,
		windows.StringToUTF16Ptr(text),
		windows.StringToUTF16Ptr(windowTitle),
		windows.MB_YESNO|windows.MB_ICONWARNING,
	)
	if choice == messageBoxYes {
		if err := openInBrowser(webView2DownloadURL); err != nil {
			log.Error("window", "cannot open download page: %v", err)
		}
	}
}
