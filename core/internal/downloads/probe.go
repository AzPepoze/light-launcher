package downloads

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	probeTimeout = 30 * time.Second
	maxRedirects = 8
)

// ProbeResult describes a remote file before the transfer starts.
type ProbeResult struct {
	URL                string // final URL after redirects
	Size               int64
	Ranges             bool // the server accepts byte ranges
	ContentType        string
	ContentDisposition string
	ETag               string
}

// WebPageError means the server answered with a web page instead of the file.
type WebPageError struct {
	URL         string
	Status      int
	ContentType string
}

func (e *WebPageError) Error() string {
	return "the server returned a web page instead of a file: open the link in a browser or try another host"
}

var contentRangeTotal = regexp.MustCompile(`/(\d+)\s*$`)

// reply is a response together with the URL it came from after redirects.
type reply struct {
	res *http.Response
	url string
}

// Probe reads size, range support, name and final URL. It sends a HEAD request first and
// falls back to a one-byte ranged GET, because many hosts omit the headers on HEAD.
func Probe(ctx context.Context, client *http.Client, rawURL string, headers map[string]string) (ProbeResult, error) {
	c := manualRedirects(client)
	base := requestHeaders(rawURL, headers)

	var fromHead *ProbeResult
	var headLength int64
	if head, err := follow(ctx, c, http.MethodHead, rawURL, nil, base); err == nil {
		headLength = contentLength(head.res)
		if isFileReply(head.res) && headLength > 0 {
			result := resultOf(head, headLength)
			fromHead = &result
		}
		head.res.Body.Close()
		if fromHead != nil && fromHead.Ranges {
			return *fromHead, nil
		}
	}

	fallback, err := follow(ctx, c, http.MethodGet, rawURL, map[string]string{"Range": "bytes=0-0"}, base)
	if err != nil {
		if fromHead != nil {
			return *fromHead, nil
		}
		return ProbeResult{}, err
	}
	defer fallback.res.Body.Close()
	return settleProbe(fallback, fromHead, headLength)
}

// settleProbe picks the answer from the ranged GET, keeping a usable HEAD answer when the GET is unusable.
func settleProbe(fallback reply, fromHead *ProbeResult, headLength int64) (ProbeResult, error) {
	status := fallback.res.StatusCode
	if isHTML(fallback.res) {
		if fromHead != nil {
			return *fromHead, nil
		}
		return ProbeResult{}, &WebPageError{URL: fallback.url, Status: status, ContentType: fallback.res.Header.Get("Content-Type")}
	}
	if status >= http.StatusBadRequest {
		if fromHead != nil {
			return *fromHead, nil
		}
		return ProbeResult{}, fmt.Errorf("probe failed: HTTP %d", status)
	}
	return resultOf(fallback, headLength), nil
}

// manualRedirects returns a copy of the client that stops at each redirect, so follow() can strip credentials.
func manualRedirects(client *http.Client) *http.Client {
	c := *client
	c.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	c.Timeout = probeTimeout
	return &c
}

// follow walks redirects by hand so cookies and referers never reach a different site.
func follow(ctx context.Context, client *http.Client, method, start string, extra, base map[string]string) (reply, error) {
	current := start
	for hop := 0; hop <= maxRedirects; hop++ {
		req, err := http.NewRequestWithContext(ctx, method, current, nil)
		if err != nil {
			return reply{}, err
		}
		headers := stripForeignCredentials(mergeHeaders(base, extra), start, current)
		for name, value := range headers {
			req.Header.Set(name, value)
		}
		res, err := client.Do(req)
		if err != nil {
			return reply{}, err
		}
		location := res.Header.Get("Location")
		if !isRedirect(res.StatusCode) || location == "" {
			return reply{res: res, url: current}, nil
		}
		res.Body.Close()
		next, err := req.URL.Parse(location)
		if err != nil {
			return reply{}, err
		}
		current = next.String()
	}
	return reply{}, errors.New("too many redirects")
}

func isRedirect(status int) bool {
	switch status {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	}
	return false
}

// isFileReply reports whether the reply is a successful non-web-page answer.
func isFileReply(res *http.Response) bool {
	return res.StatusCode >= 200 && res.StatusCode < 300 && !isHTML(res)
}

func isHTML(res *http.Response) bool {
	contentType := strings.ToLower(res.Header.Get("Content-Type"))
	return strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/xhtml")
}

func contentLength(res *http.Response) int64 {
	if res.ContentLength > 0 {
		return res.ContentLength
	}
	n, err := strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func resultOf(r reply, fallbackSize int64) ProbeResult {
	h := r.res.Header
	return ProbeResult{
		URL:                r.url,
		Size:               sizeOf(r.res, fallbackSize),
		Ranges:             r.res.StatusCode == http.StatusPartialContent || strings.Contains(strings.ToLower(h.Get("Accept-Ranges")), "bytes"),
		ContentType:        strings.ToLower(h.Get("Content-Type")),
		ContentDisposition: h.Get("Content-Disposition"),
		ETag:               h.Get("ETag"),
	}
}

// sizeOf prefers the total from Content-Range, then Content-Length, then the HEAD size.
func sizeOf(res *http.Response, fallback int64) int64 {
	if m := contentRangeTotal.FindStringSubmatch(res.Header.Get("Content-Range")); m != nil {
		if total, err := strconv.ParseInt(m[1], 10, 64); err == nil && total > 0 {
			return total
		}
	}
	if length := contentLength(res); length > 0 {
		return length
	}
	return fallback
}
