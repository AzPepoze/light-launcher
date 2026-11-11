package main

import (
	"net/url"
	"strings"
)

// isExternal reports whether a URL is outside the launcher origin.
func isExternal(target *url.URL, launcherOrigin string) bool {
	return urlOrigin(target) != launcherOrigin
}

func isWebURL(target *url.URL) bool {
	return target != nil && (target.Scheme == "http" || target.Scheme == "https") && target.Host != ""
}

// urlOrigin returns scheme and authority (host and port) of a parsed URL.
func urlOrigin(target *url.URL) string {
	if target == nil {
		return ""
	}
	return strings.ToLower(target.Scheme + "://" + target.Host)
}

// isCaptureDownload reports whether a download URL is one assist mode reports. The page may also
// start blob: or data: downloads; those stay inside WebKit and are ignored.
func isCaptureDownload(raw string) bool {
	target, err := url.Parse(raw)
	if err != nil || !isWebURL(target) {
		return false
	}
	return true
}
