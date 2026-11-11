package main

import (
	"net/url"
	"path/filepath"

	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/network"
	"github.com/mappu/miqt/qt6/webengine"
)

func newAssistSession(opts assistOptions, events *eventWriter) *assistSession {
	profile := newAssistProfile(opts.profile)
	jar := newCookieJar()
	store := profile.CookieStore()
	// Removals are ignored on purpose: setting a cookie to the same value reports only a removal, and a
	// changed value reports a removal then an add. Keeping the last added value is right for both.
	store.OnCookieAdded(func(cookie *network.QNetworkCookie) {
		jar.add(cookieFromQt(cookie))
	})
	store.LoadAllCookies()

	page := webengine.NewQWebEnginePage2(profile)
	enableSmoothScrolling(page)
	view := webengine.NewQWebEngineView6(page, nil)
	window := qt.NewQWidget2()
	layout := qt.NewQVBoxLayout(window)
	layout.SetContentsMargins(0, 0, 0, 0)
	layout.SetSpacing(0)
	if opts.message != "" {
		banner := qt.NewQLabel5(opts.message, window)
		banner.SetWordWrap(true)
		banner.SetMargin(8)
		banner.SetAlignment(qt.AlignCenter)
		banner.SetFrameShape(qt.QFrame__StyledPanel)
		layout.AddWidget(banner.QWidget)
	}
	layout.AddWidget(view.QWidget)
	window.SetWindowTitle(opts.title)
	window.SetMinimumSize2(800, 600)
	window.Resize(assistWidth, assistHeight)

	session := &assistSession{
		opts:    opts,
		events:  events,
		jar:     jar,
		profile: profile,
		page:    page,
		window:  window,
	}

	window.OnCloseEvent(session.onWindowClosed)
	profile.OnDownloadRequested(session.onDownload)
	page.OnAcceptNavigationRequest(func(super func(url *qt.QUrl, typeVal webengine.QWebEnginePage__NavigationType, isMainFrame bool) bool, url *qt.QUrl, typeVal webengine.QWebEnginePage__NavigationType, isMainFrame bool) bool {
		if !isWebURL(url) {
			log.Warn(assistLogName, "blocked navigation on scheme %q", url.Scheme())
			return false
		}
		return super(url, typeVal, isMainFrame)
	})
	page.OnCreateWindow(func(super func(typeVal webengine.QWebEnginePage__WebWindowType) *webengine.QWebEnginePage, typeVal webengine.QWebEnginePage__WebWindowType) *webengine.QWebEnginePage {
		return page
	})
	return session
}

// newAssistProfile keeps cookies and logins in the profile folder between runs.
func newAssistProfile(dir string) *webengine.QWebEngineProfile {
	profile := webengine.NewQWebEngineProfile2(profileName)
	profile.SetPersistentStoragePath(dir)
	profile.SetCachePath(filepath.Join(dir, "cache"))
	profile.SetPersistentCookiesPolicy(webengine.QWebEngineProfile__ForcePersistentCookies)
	return profile
}

func cookieFromQt(raw *network.QNetworkCookie) cookie {
	return cookie{
		Name:   string(raw.Name()),
		Value:  string(raw.Value()),
		Domain: raw.Domain(),
		Path:   raw.Path(),
		Secure: raw.IsSecure(),
	}
}

// isCaptureDownload reports whether a download URL is one this mode reports. The page may also
// start blob: or data: downloads; those stay inside Chromium and are ignored.
func isCaptureDownload(raw string) bool {
	target, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return (target.Scheme == "http" || target.Scheme == "https") && target.Host != ""
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
