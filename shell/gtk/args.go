//go:build linux

package main

import (
	"flag"
	"net/url"
	"os"

	"light-launcher/shell/common/urlx"
)

const (
	exitOK      = 0
	exitFailure = 1
	exitBadArgs = 2

	windowTitle   = "LightLauncher"
	minWidth      = 800
	minHeight     = 600
	defaultWidth  = 1100
	defaultHeight = 720
)

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
