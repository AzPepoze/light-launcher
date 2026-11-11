//go:build windows

// Command light-launcher-shell.exe shows the launcher UI in a WebView2 window.
//
// The launcher passes the URL in the LL_URL environment variable. A URL argument is the fallback.
// Usage: [LL_URL=<url>] light-launcher-shell.exe [--transparent] [url]
package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/jchv/go-webview2"

	"light-launcher/shell/logx"
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
	raw, err := shellURL(flags.Args())
	if err != nil {
		return options{}, err
	}
	// url.Parse errors repeat the input, which holds the token, so the error is replaced.
	target, err := url.Parse(raw)
	if err != nil {
		return options{}, errors.New("URL is not valid")
	}
	if (target.Scheme != "http" && target.Scheme != "https") || target.Host == "" {
		return options{}, errors.New("URL must be http:// or https:// with a host")
	}
	return options{url: target, transparent: *transparent}, nil
}

// shellURL returns the UI URL from LL_URL, or from the single argument when LL_URL is not set.
func shellURL(args []string) (string, error) {
	if value := os.Getenv("LL_URL"); value != "" {
		return value, nil
	}
	if len(args) != 1 {
		return "", fmt.Errorf("expected one URL, got %d arguments", len(args))
	}
	return args[0], nil
}

func main() {
	if assistRequested(os.Args[1:]) {
		os.Exit(runAssist(os.Args[1:]))
	}

	opts, err := parseOptions(os.Args[1:])
	if err != nil {
		log.Error("args", "%v", err)
		log.Error("args", "usage: LL_URL=<url> light-launcher-shell.exe [--transparent]")
		log.Error("args", assistUsage)
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
	if err := bindNativeFunctions(w); err != nil {
		log.Error("bridge", "cannot bind native functions: %v", err)
		os.Exit(1)
	}
	w.Init(bridgeScript)
	w.Navigate(opts.url.String())
	log.Info("window", "opened %s", origin(opts.url))
	w.Run()
}

// origin returns scheme and host without the path, query or token. Logs use this, never the full URL.
func origin(u *url.URL) string {
	return u.Scheme + "://" + u.Host
}
