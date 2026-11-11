package main

import (
	"net/url"
	"path"
)

// shellSession holds the state shared by the GTK callbacks. All callbacks run on the main thread.
type shellSession struct {
	assist         bool
	launcherOrigin string
	events         *eventWriter
	finished       bool
	code           int
}

func newNormalSession(opts options) *shellSession {
	return &shellSession{launcherOrigin: urlOrigin(opts.url), code: exitOK}
}

func newAssistSession(events *eventWriter) *shellSession {
	return &shellSession{assist: true, events: events, code: exitOK}
}

func (s *shellSession) onReady() {
	if !s.assist {
		return
	}
	if err := s.events.write(readyEvent()); err != nil {
		log.Error("assist", "cannot write ready event: %v", err)
		s.finished = true
		s.code = exitFailure
		quitShell()
	}
}

func (s *shellSession) onDownload(rawURL, filename string, size int64,
	cookie, referer, userAgent string) {
	if s.finished || !s.assist || !isCaptureDownload(rawURL) {
		return
	}
	s.finished = true
	event := downloadEvent(rawURL, downloadFilename(filename, rawURL), size,
		downloadHeaders(cookie, referer, userAgent))
	s.code = exitOK
	if err := s.events.write(event); err != nil {
		log.Error("assist", "cannot write download event: %v", err)
		s.code = exitFailure
	}
	quitShell()
}

func (s *shellSession) onClosed() {
	if s.finished {
		return
	}
	s.finished = true
	s.code = exitOK
	if s.assist {
		if err := s.events.write(closedEvent()); err != nil {
			log.Error("assist", "cannot write closed event: %v", err)
			s.code = exitFailure
		}
	}
	quitShell()
}

func (s *shellSession) onBridgeMessage(body, pageURI string) {
	if !allowsBridge(pageURI, s.launcherOrigin) {
		log.Warn("bridge", "ignored a message from another origin")
		return
	}
	request, ok := parseBridgeRequest(body)
	if !ok {
		return
	}
	switch request.Method {
	case "pickFolder":
		openPicker(request.ID, true)
	case "pickFile":
		openPicker(request.ID, false)
	default:
		log.Warn("bridge", "unknown method %q", request.Method)
	}
}

func (s *shellSession) onPickerResult(id int, pickedPath string) {
	script, err := resolveScript(id, pickedPath)
	if err != nil {
		log.Error("bridge", "cannot build reply: %v", err)
		return
	}
	evalJS(script)
}

// linkClicked opens an external link in the default browser. It reports whether the navigation
// must be stopped, even when no browser accepted the address.
func (s *shellSession) linkClicked(rawURL string) bool {
	target, err := url.Parse(rawURL)
	if err != nil || !isExternal(target, s.launcherOrigin) {
		return false
	}
	if err := openURI(rawURL); err != nil {
		log.Warn("links", "could not open external link on host %s", target.Host)
	} else {
		log.Info("links", "opened external link on host %s", target.Host)
	}
	return true
}

func (s *shellSession) webURL(rawURL string) bool {
	target, err := url.Parse(rawURL)
	return err == nil && isWebURL(target)
}

func (s *shellSession) captureDownload(rawURL string) bool {
	return s.assist && isCaptureDownload(rawURL)
}

// downloadFilename uses the name the server suggested, or the last part of the URL when the
// response headers were not ready before the download was cancelled.
func downloadFilename(suggested, rawURL string) string {
	if suggested != "" {
		return suggested
	}
	target, err := url.Parse(rawURL)
	if err != nil {
		return "download"
	}
	name := path.Base(target.Path)
	if name == "." || name == "/" || name == "" {
		return "download"
	}
	return name
}

// downloadHeaders adds only the headers the shell knows the browser would send.
func downloadHeaders(cookieHeader, referer, userAgent string) map[string]string {
	headers := make(map[string]string, 3)
	if cookieHeader != "" {
		headers["Cookie"] = cookieHeader
	}
	if referer != "" {
		headers["Referer"] = referer
	}
	if userAgent != "" {
		headers["User-Agent"] = userAgent
	}
	return headers
}
