package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"light-launcher/shell/common/assist"
	"light-launcher/shell/common/urlx"
)

const (
	assistFlag        = "--assist"
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
	opts := assistOptions{title: assist.Title}
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
		opts.title = assist.Title
	}
	target, err := urlx.ParseTarget(rawURL)
	if err != nil {
		return assistOptions{}, err
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
