package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	assistTitle   = "LightLauncher"
	assistUsage   = "usage: light-launcher-shell --assist <url> --profile <dir> [--title <text>] [--message <text>]"
	assistLogName = "assist"
)

type assistOptions struct {
	target  *url.URL
	profile string
	title   string
	message string
}

// wantsAssist reports whether the command line selects assist mode. Both flag spellings are
// accepted so the standard flag package can parse the rest.
func wantsAssist(args []string) bool {
	for _, arg := range args {
		if arg == "-assist" || arg == "--assist" ||
			strings.HasPrefix(arg, "-assist=") || strings.HasPrefix(arg, "--assist=") {
			return true
		}
	}
	return false
}

func parseAssistOptions(args []string) (assistOptions, error) {
	flags := flag.NewFlagSet("light-launcher-shell --assist", flag.ContinueOnError)
	target := flags.String("assist", "", "page URL")
	profile := flags.String("profile", "", "persistent browser profile folder")
	title := flags.String("title", assistTitle, "window title")
	message := flags.String("message", "", "banner text above the page")
	if err := flags.Parse(args); err != nil {
		return assistOptions{}, err
	}
	if len(flags.Args()) != 0 {
		return assistOptions{}, fmt.Errorf("expected no extra arguments, got %d", len(flags.Args()))
	}
	if *profile == "" {
		return assistOptions{}, errors.New("--profile is required")
	}
	page, err := parseTarget(*target)
	if err != nil {
		return assistOptions{}, err
	}
	if *title == "" {
		*title = assistTitle
	}
	return assistOptions{target: page, profile: *profile, title: *title, message: *message}, nil
}

func runAssist(args []string) int {
	opts, err := parseAssistOptions(args)
	if err != nil {
		log.Error("args", "%v", err)
		log.Error("args", assistUsage)
		return exitBadArgs
	}
	profile, err := prepareProfile(opts.profile)
	if err != nil {
		log.Error(assistLogName, "%v", err)
		writeError(os.Stdout, err.Error())
		return exitFailure
	}

	session := newAssistSession(newEventWriter(os.Stdout))
	log.Info(assistLogName, "opened %s", urlOrigin(opts.target))
	config := gtkConfig{
		url:        opts.target.String(),
		title:      opts.title,
		profileDir: profile,
		message:    opts.message,
		assist:     true,
		devtools:   developerTools(),
		disableGPU: gpuDisabled(),
	}
	if _, err := runGTK(session, config); err != nil {
		log.Error(assistLogName, "%v", err)
		if !session.finished {
			session.finished = true
			session.code = exitFailure
			writeError(os.Stdout, err.Error())
		}
		return exitFailure
	}
	if !session.finished {
		session.finished = true
		session.code = exitFailure
		writeError(os.Stdout, "window closed before an event")
	}
	return session.code
}

// prepareProfile creates the persistent profile folder and returns its absolute path.
func prepareProfile(dir string) (string, error) {
	absolute, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve profile folder: %w", err)
	}
	if err := os.MkdirAll(absolute, 0o700); err != nil {
		return "", fmt.Errorf("create profile folder: %w", err)
	}
	return absolute, nil
}
