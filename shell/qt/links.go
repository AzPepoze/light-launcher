package main

import (
	"fmt"

	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/webengine"
)

// routeExternalLinks opens links that leave the launcher origin in the default browser.
func routeExternalLinks(page *webengine.QWebEnginePage, launcherOrigin string) {
	page.OnAcceptNavigationRequest(func(super func(url *qt.QUrl, typeVal webengine.QWebEnginePage__NavigationType, isMainFrame bool) bool, url *qt.QUrl, typeVal webengine.QWebEnginePage__NavigationType, isMainFrame bool) bool {
		if isMainFrame && typeVal == webengine.QWebEnginePage__NavigationTypeLinkClicked && isExternal(url, launcherOrigin) {
			openInBrowser(url)
			return false
		}
		return super(url, typeVal, isMainFrame)
	})

	page.OnCreateWindow(func(super func(typeVal webengine.QWebEnginePage__WebWindowType) *webengine.QWebEnginePage, typeVal webengine.QWebEnginePage__WebWindowType) *webengine.QWebEnginePage {
		// Pages opened by target=_blank or window.open are only used to pick the URL.
		popup := webengine.NewQWebEnginePage2(page.Profile())
		popup.OnAcceptNavigationRequest(func(super func(url *qt.QUrl, typeVal webengine.QWebEnginePage__NavigationType, isMainFrame bool) bool, url *qt.QUrl, typeVal webengine.QWebEnginePage__NavigationType, isMainFrame bool) bool {
			if isWebURL(url) && isExternal(url, launcherOrigin) {
				openInBrowser(url)
				popup.DeleteLater()
			}
			return false
		})
		return popup
	})
}

// isExternal reports whether a URL is outside the launcher origin.
func isExternal(url *qt.QUrl, launcherOrigin string) bool {
	return qUrlOrigin(url) != launcherOrigin
}

func isWebURL(url *qt.QUrl) bool {
	return url.Scheme() == "http" || url.Scheme() == "https"
}

// qUrlOrigin returns scheme and authority (host and port) of a Qt URL.
func qUrlOrigin(url *qt.QUrl) string {
	return fmt.Sprintf("%s://%s", url.Scheme(), url.Authority())
}

func openInBrowser(url *qt.QUrl) {
	if !qt.QDesktopServices_OpenUrl(url) {
		log.Warn("links", "could not open external link on host %s", url.Host())
		return
	}
	log.Info("links", "opened external link on host %s", url.Host())
}
