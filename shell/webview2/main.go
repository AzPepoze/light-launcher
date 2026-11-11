//go:build windows

// Command light-launcher-shell.exe shows the launcher UI in a WebView2 window.
//
// The launcher passes the URL in the LL_URL environment variable. A URL argument is the fallback.
// Usage: [LL_URL=<url>] light-launcher-shell.exe [--transparent] [url]
package main

import (
	"flag"
	"net/url"
	"os"

	"github.com/jchv/go-webview2"

	"light-launcher/shell/common/assist"
	"light-launcher/shell/common/logx"
	"light-launcher/shell/common/urlx"
)

const (
	windowTitle   = "LightLauncher"
	minWidth      = 800
	minHeight     = 600
	defaultWidth  = 1100
	defaultHeight = 720
)

var log = logx.New(os.Stderr)

type options struct {
	url         *url.URL
	transparent bool
}

func parseOptions(args []string) (options, error) {
	flags := flag.NewFlagSet("light-launcher-shell", flag.ContinueOnError)
	transparent := flags.Bool("transparent", false, "use a transparent window background (not supported on Windows yet)")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	raw, err := urlx.ShellURL(os.Getenv(urlx.EnvURL), flags.Args())
	if err != nil {
		return options{}, err
	}
	target, err := urlx.ParseTarget(raw)
	if err != nil {
		return options{}, err
	}
	return options{url: target, transparent: *transparent}, nil
}

func main() {
	if assistRequested(os.Args[1:]) {
		os.Exit(runAssist(os.Args[1:]))
	}

	opts, err := parseOptions(os.Args[1:])
	if err != nil {
		log.Error("args", "%v", err)
		log.Error("args", "usage: LL_URL=<url> light-launcher-shell.exe [--transparent]")
		log.Error("args", assist.Usage)
		os.Exit(2)
	}
	if opts.transparent {
		log.Warn("window", "--transparent is not supported by the WebView2 shell yet, ignored")
	}

	if !webView2Installed() {
		offerRuntimeDownload()
		log.Error("window", "WebView2 runtime is not installed")
		os.Exit(1)
	}

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		WindowOptions: webview2.WindowOptions{
			Title:  windowTitle,
			Width:  defaultWidth,
			Height: defaultHeight,
			Center: true,
		},
	})
	if w == nil {
		offerRuntimeDownload()
		log.Error("window", "WebView2 runtime could not start")
		os.Exit(1)
	}
	defer w.Destroy()

	w.SetSize(minWidth, minHeight, webview2.HintMin)
	launcherOrigin := urlx.Origin(opts.url)
	secret, err := newBridgeSecret()
	if err != nil {
		log.Error("bridge", "%v", err)
		os.Exit(1)
	}
	if err := bindNativeFunctions(w, launcherOrigin, secret, nativeActions{
		pick: func(method string, id int) { go pickInBackground(w, method, id) },
		open: openInBrowser,
	}); err != nil {
		log.Error("bridge", "cannot bind native functions: %v", err)
		os.Exit(1)
	}
	script, err := bridgeScriptFor(launcherOrigin, secret)
	if err != nil {
		log.Error("bridge", "%v", err)
		os.Exit(1)
	}
	w.Init(script)
	w.Navigate(opts.url.String())
	log.Info("window", "opened %s", launcherOrigin)
	w.Run()
}
