package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strings"

	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/webengine"
)

// Assist mode is for a human at the page. The shell only observes the download the page starts:
// it never clicks, types, fills forms, solves challenges, spoofs fingerprints or hides Chromium.
const (
	exitOK        = 0
	exitFailure   = 1
	exitBadArgs   = 2
	assistTitle   = "LightLauncher"
	assistWidth   = 1000
	assistHeight  = 720
	profileName   = "light-launcher-assist"
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
	if err := os.MkdirAll(opts.profile, 0o700); err != nil {
		log.Error(assistLogName, "cannot create profile folder: %v", err)
		writeError(os.Stdout, fmt.Sprintf("cannot create profile folder: %v", err))
		return exitFailure
	}

	// Only the program name goes to Qt. The shell flags are already parsed.
	qt.NewQApplication([]string{os.Args[0]})
	session := newAssistSession(opts, newEventWriter(os.Stdout))
	code := session.exec()
	session.dispose()
	return code
}

// assistSession owns the assist window and its persistent profile. Callbacks run on the Qt thread.
type assistSession struct {
	opts     assistOptions
	events   *eventWriter
	jar      *cookieJar
	profile  *webengine.QWebEngineProfile
	page     *webengine.QWebEnginePage
	window   *qt.QWidget
	finished bool
	code     int
}

func (s *assistSession) exec() int {
	s.window.Show()
	if err := s.events.write(readyEvent()); err != nil {
		log.Error(assistLogName, "cannot write ready event: %v", err)
		s.window.Close()
		return exitFailure
	}
	s.page.Load(qt.NewQUrl3(s.opts.target.String()))
	log.Info(assistLogName, "opened %s", origin(s.opts.target))

	qt.QApplication_Exec()
	runtime.KeepAlive(s)
	if !s.finished {
		s.finished = true
		s.code = exitFailure
		if err := s.events.write(errorEvent("window closed before an event")); err != nil {
			log.Error(assistLogName, "cannot write error event: %v", err)
		}
	}
	return s.code
}

// onDownload captures the request, cancels it before Chromium writes anything, and ends the mode.
func (s *assistSession) onDownload(download *webengine.QWebEngineDownloadRequest) {
	if s.finished {
		download.Cancel()
		return
	}
	rawURL := download.Url().ToString()
	if !isCaptureDownload(rawURL) {
		download.Cancel()
		log.Info(assistLogName, "ignored download on scheme %q", download.Url().Scheme())
		return
	}
	s.finished = true
	event := downloadEvent(rawURL, download.SuggestedFileName(), download.TotalBytes(),
		downloadHeaders(s.jar.header(rawURL), s.pageURL(), s.profile.HttpUserAgent()))
	download.Cancel()
	s.code = exitOK
	if err := s.events.write(event); err != nil {
		log.Error(assistLogName, "cannot write download event: %v", err)
		s.code = exitFailure
	}
	s.window.Close()
	qt.QCoreApplication_Quit()
}

func (s *assistSession) onWindowClosed(super func(event *qt.QCloseEvent), event *qt.QCloseEvent) {
	super(event)
	if s.finished {
		return
	}
	s.finished = true
	s.code = exitOK
	if err := s.events.write(closedEvent()); err != nil {
		log.Error(assistLogName, "cannot write closed event: %v", err)
		s.code = exitFailure
	}
	qt.QCoreApplication_Quit()
}

func (s *assistSession) pageURL() string {
	if current := s.page.Url().ToString(); current != "" {
		return current
	}
	return s.opts.target.String()
}

func (s *assistSession) dispose() {
	if s.page != nil {
		s.page.Delete()
	}
	if s.window != nil {
		s.window.Delete()
	}
	if s.profile != nil {
		s.profile.Delete()
	}
}
