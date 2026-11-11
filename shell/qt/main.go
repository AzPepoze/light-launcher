// Command light-launcher-shell shows the launcher UI in a native Qt6 WebEngine window.
//
// The launcher passes the URL in the LL_URL environment variable. A URL argument is the fallback.
// Usage: [LL_URL=<url>] light-launcher-shell [--transparent] [url]
//
// The --assist mode instead shows one web page for a human to use and reports the download the page
// starts. It never acts on the page: the user logs in, presses the buttons and solves any challenge.
package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"runtime"

	qt "github.com/mappu/miqt/qt6"

	"light-launcher/shell/logx"
)

var log = logx.New(os.Stderr)

type options struct {
	url         *url.URL
	transparent bool
}

func parseOptions(args []string) (options, error) {
	flags := flag.NewFlagSet("light-launcher-shell", flag.ContinueOnError)
	transparent := flags.Bool("transparent", false, "use a transparent window background")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	raw, err := shellURL(flags.Args())
	if err != nil {
		return options{}, err
	}
	target, err := parseTarget(raw)
	if err != nil {
		return options{}, err
	}
	return options{url: target, transparent: *transparent}, nil
}

// parseTarget validates a page URL. url.Parse errors repeat the input, which can hold a token,
// so the error is replaced.
func parseTarget(raw string) (*url.URL, error) {
	target, err := url.Parse(raw)
	if err != nil {
		return nil, errors.New("URL is not valid")
	}
	if (target.Scheme != "http" && target.Scheme != "https") || target.Host == "" {
		return nil, errors.New("URL must be http:// or https:// with a host")
	}
	return target, nil
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
	args := os.Args[1:]
	if wantsAssist(args) {
		os.Exit(runAssist(args))
	}

	opts, err := parseOptions(args)
	if err != nil {
		log.Error("args", "%v", err)
		log.Error("args", "usage: LL_URL=<url> light-launcher-shell [--transparent]")
		os.Exit(2)
	}

	// Only the program name goes to Qt. The launcher flags are already parsed.
	qt.NewQApplication([]string{os.Args[0]})
	win := newWindow(opts)
	win.view.Load(qt.NewQUrl3(opts.url.String()))
	win.view.Show()
	log.Info("window", "opened %s", origin(opts.url))

	code := qt.QApplication_Exec()
	// The Go side must stay reachable while Qt runs, or its finalizers can free the view.
	runtime.KeepAlive(win)
	os.Exit(code)
}

// origin returns scheme and host without the path, query or token. Logs use this, never the full URL.
func origin(u *url.URL) string {
	return u.Scheme + "://" + u.Host
}
