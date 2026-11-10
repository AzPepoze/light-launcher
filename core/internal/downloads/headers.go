package downloads

import (
	"net/http"
	"net/url"
	"strings"
)

// userAgent is the browser string hosts expect from a download client.
const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

// refererByHost lists hosts whose CDNs refuse requests without a Referer.
var refererByHost = map[string]string{
	"f95zone.to":      "https://f95zone.to/",
	"itch.io":         "https://itch.io/",
	"itch.zone":       "https://itch.io/",
	"mediafire.com":   "https://www.mediafire.com/",
	"mega.nz":         "https://mega.nz/",
	"pixeldrain.com":  "https://pixeldrain.com/",
	"gofile.io":       "https://gofile.io/",
	"workupload.com":  "https://workupload.com/",
	"datanodes.to":    "https://datanodes.to/",
	"krakenfiles.com": "https://krakenfiles.com/",
	"buzzheavier.com": "https://buzzheavier.com/",
	"vikingfile.com":  "https://vikingfile.com/",
}

func hostOf(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Hostname())
}

// defaultReferer finds the referer for a host or one of its parent domains.
func defaultReferer(host string) string {
	for host != "" {
		if referer, ok := refererByHost[host]; ok {
			return referer
		}
		dot := strings.IndexByte(host, '.')
		if dot < 0 {
			break
		}
		host = host[dot+1:]
	}
	return ""
}

// requestHeaders builds the headers every request starts with. Caller headers win.
func requestHeaders(rawURL string, extra map[string]string) map[string]string {
	defaults := map[string]string{
		"User-Agent":      userAgent,
		"Accept-Encoding": "identity",
	}
	if referer := defaultReferer(hostOf(rawURL)); referer != "" {
		defaults["Referer"] = referer
	}
	return mergeHeaders(defaults, extra)
}

// mergeHeaders returns a copy of base with extra applied on top, using canonical header names.
func mergeHeaders(base, extra map[string]string) map[string]string {
	out := make(map[string]string, len(base)+len(extra))
	for name, value := range base {
		out[name] = value
	}
	for name, value := range extra {
		out[http.CanonicalHeaderKey(name)] = value
	}
	return out
}

// sameSiteHost matches a host with its parent domain, such as cdn.example.com and example.com.
func sameSiteHost(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	return a == b || strings.HasSuffix(a, "."+b) || strings.HasSuffix(b, "."+a)
}

// stripForeignCredentials removes the cookie and referer when a download points at another site.
func stripForeignCredentials(headers map[string]string, originURL, targetURL string) map[string]string {
	if sameSiteHost(hostOf(targetURL), hostOf(originURL)) {
		return headers
	}
	out := make(map[string]string, len(headers))
	for name, value := range headers {
		if name == "Cookie" || name == "Referer" {
			continue
		}
		out[name] = value
	}
	return out
}
