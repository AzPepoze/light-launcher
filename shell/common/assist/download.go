package assist

import (
	"net/url"
	"path"
	"strings"
)

// CookiePair is one cookie the browser holds for a download address.
type CookiePair struct {
	Name  string
	Value string
}

// CookieHeader builds the Cookie header. Cookies without a name are skipped.
func CookieHeader(cookies []CookiePair) string {
	parts := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie.Name == "" {
			continue
		}
		parts = append(parts, cookie.Name+"="+cookie.Value)
	}
	return strings.Join(parts, "; ")
}

// DownloadHeaders keeps only the headers the shell knows the browser would send.
func DownloadHeaders(cookie, referer, userAgent string) map[string]string {
	headers := make(map[string]string, 3)
	if cookie != "" {
		headers["Cookie"] = cookie
	}
	if referer != "" {
		headers["Referer"] = referer
	}
	if userAgent != "" {
		headers["User-Agent"] = userAgent
	}
	return headers
}

// ResultFileName returns the file name of a download result path, on either separator.
func ResultFileName(resultPath string) string {
	if index := strings.LastIndexAny(resultPath, `\/`); index >= 0 {
		return resultPath[index+1:]
	}
	return resultPath
}

// DownloadFileName returns the name the server suggested, or the last part of the
// URL when the response headers were not ready before the download was cancelled.
func DownloadFileName(suggested, rawURL string) string {
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
