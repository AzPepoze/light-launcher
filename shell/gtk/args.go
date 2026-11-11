package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
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
	if !isWebURL(target) {
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
