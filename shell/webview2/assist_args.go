package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const (
	assistFlag        = "--assist"
	assistUsage       = "usage: light-launcher-shell --assist <url> --profile <dir> [--title <text>] [--message <text>]"
	assistTitle       = "LightLauncher"
	assistProfileFlag = "--profile"
	assistTitleFlag   = "--title"
	assistMessageFlag = "--message"
)

// assistOptions is the parsed command line of the assist mode.
type assistOptions struct {
	url     *url.URL
	profile string
	title   string
	message string
}

// assistRequested reports whether the command line selects assist mode.
func assistRequested(args []string) bool {
	return len(args) > 0 && args[0] == assistFlag
}

// parseAssistOptions reads the assist arguments. A human drives the page; the shell only
// watches the download the page starts and cancels it, so the launcher can download the file.
func parseAssistOptions(args []string) (assistOptions, error) {
	if len(args) == 0 || args[0] != assistFlag {
		return assistOptions{}, fmt.Errorf("assist mode starts with %s", assistFlag)
	}
	opts := assistOptions{title: assistTitle}
	var rawURL string
	for i := 1; i < len(args); i++ {
		switch arg := args[i]; arg {
		case assistProfileFlag:
			value, next, err := flagValue(args, i, arg)
			if err != nil {
				return assistOptions{}, err
			}
			opts.profile = value
			i = next
		case assistTitleFlag:
			value, next, err := flagValue(args, i, arg)
			if err != nil {
				return assistOptions{}, err
			}
			opts.title = value
			i = next
		case assistMessageFlag:
			value, next, err := flagValue(args, i, arg)
			if err != nil {
				return assistOptions{}, err
			}
			opts.message = value
			i = next
		default:
			if strings.HasPrefix(arg, "-") {
				return assistOptions{}, fmt.Errorf("unknown option %s", arg)
			}
			if rawURL != "" {
				return assistOptions{}, fmt.Errorf("expected one URL, got extra argument %q", arg)
			}
			rawURL = arg
		}
	}
	if rawURL == "" {
		return assistOptions{}, errors.New("expected one URL after --assist")
	}
	if opts.profile == "" {
		return assistOptions{}, errors.New("--profile is required")
	}
	if opts.title == "" {
		opts.title = assistTitle
	}
	target, err := url.Parse(rawURL)
	if err != nil {
		// url.Parse repeats the input, which may hold a token, so the error is replaced.
		return assistOptions{}, errors.New("URL is not valid")
	}
	if !isHTTPURL(target) {
		return assistOptions{}, errors.New("URL must be http:// or https:// with a host")
	}
	opts.url = target
	return opts, nil
}

// flagValue returns the value that follows a flag and the index of that value.
func flagValue(args []string, index int, name string) (string, int, error) {
	if index+1 >= len(args) {
		return "", index, fmt.Errorf("%s needs a value", name)
	}
	return args[index+1], index + 1, nil
}

// isHTTPURL reports whether target is an http or https address with a host.
func isHTTPURL(target *url.URL) bool {
	if target == nil {
		return false
	}
	return (target.Scheme == "http" || target.Scheme == "https") && target.Host != ""
}

// isHTTPAddress applies isHTTPURL to a raw address, such as a download or navigation URI.
func isHTTPAddress(raw string) bool {
	target, err := url.Parse(raw)
	return err == nil && isHTTPURL(target)
}
