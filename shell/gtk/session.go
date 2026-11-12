//go:build linux

package main

import (
	"net/url"

	"light-launcher/shell/common/assist"
	"light-launcher/shell/common/bridge"
	"light-launcher/shell/common/urlx"
)

// shellSession holds the state shared by the GTK callbacks. All callbacks run on the main thread.
type shellSession struct {
	assist         bool
	launcherOrigin string
	events         *assist.Writer
	geometry       windowGeometry
	geometrySet    bool
	finished       bool
	code           int
}

func newNormalSession(opts options) *shellSession {
	return &shellSession{launcherOrigin: urlx.Origin(opts.url), code: exitOK}
}

func newAssistSession(events *assist.Writer) *shellSession {
	return &shellSession{assist: true, events: events, code: exitOK}
}

func (s *shellSession) onReady() {
	if !s.assist {
		return
	}
	if err := s.events.Write(assist.Ready()); err != nil {
		log.Error("assist", "cannot write ready event: %v", err)
		s.finished = true
		s.code = exitFailure
		quitShell()
	}
}

func (s *shellSession) onDownload(rawURL, filename string, size int64,
	cookie, referer, userAgent string) {
	if s.finished || !s.assist || !urlx.IsWebAddress(rawURL) {
		return
	}
	s.finished = true
	event := assist.Download(rawURL, assist.DownloadFileName(filename, rawURL), size,
		assist.DownloadHeaders(cookie, referer, userAgent))
	s.code = exitOK
	if err := s.events.Write(event); err != nil {
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
		if err := s.events.Write(assist.Closed()); err != nil {
			log.Error("assist", "cannot write closed event: %v", err)
			s.code = exitFailure
		}
	}
	quitShell()
}

func (s *shellSession) onGeometry(x, y, width, height int, positioned bool) {
	if s.assist || width <= 0 || height <= 0 {
		return
	}
	s.geometry = windowGeometry{
		Version:    geometryVersion,
		X:          x,
		Y:          y,
		Width:      width,
		Height:     height,
		Positioned: positioned,
	}
	s.geometrySet = true
}

func (s *shellSession) onBridgeMessage(body, pageURI string) {
	if !bridge.Allows(pageURI, s.launcherOrigin) {
		log.Warn("bridge", "ignored a message from another origin")
		return
	}
	request, ok := bridge.ParseRequest(body)
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
	script, err := bridge.ResolveScript(id, pickedPath)
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
	if err != nil || !urlx.IsExternal(target, s.launcherOrigin) {
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
	return urlx.IsWebAddress(rawURL)
}

func (s *shellSession) captureDownload(rawURL string) bool {
	return s.assist && urlx.IsWebAddress(rawURL)
}
