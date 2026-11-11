package main

import (
	"net"
	"net/url"
	"sort"
	"strings"
	"sync"
)

type cookie struct {
	Name   string
	Value  string
	Domain string
	Path   string
	Secure bool
}

// cookieJar keeps the profile's cookies in memory so a download header can be built without Qt.
type cookieJar struct {
	mu      sync.Mutex
	cookies []cookie
}

func newCookieJar() *cookieJar {
	return &cookieJar{}
}

func (j *cookieJar) add(c cookie) {
	if c.Name == "" {
		return
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	for i := range j.cookies {
		if j.cookies[i].sameIdentifier(c) {
			j.cookies[i] = c
			return
		}
	}
	j.cookies = append(j.cookies, c)
}

func (j *cookieJar) remove(c cookie) {
	j.mu.Lock()
	defer j.mu.Unlock()
	for i := range j.cookies {
		if j.cookies[i].sameIdentifier(c) {
			j.cookies = append(j.cookies[:i], j.cookies[i+1:]...)
			return
		}
	}
}

// header returns the Cookie request header for a download URL. Longer paths go first, as browsers do.
func (j *cookieJar) header(rawURL string) string {
	target, err := url.Parse(rawURL)
	if err != nil || (target.Scheme != "http" && target.Scheme != "https") || target.Host == "" {
		return ""
	}
	requestPath := target.EscapedPath()
	if requestPath == "" {
		requestPath = "/"
	}
	host := target.Hostname()

	j.mu.Lock()
	matches := make([]cookie, 0, len(j.cookies))
	for _, c := range j.cookies {
		if c.Secure && target.Scheme != "https" {
			continue
		}
		if !domainMatches(c.Domain, host) || !pathMatches(c.Path, requestPath) {
			continue
		}
		matches = append(matches, c)
	}
	j.mu.Unlock()

	sort.SliceStable(matches, func(a, b int) bool {
		return len(normalizePath(matches[a].Path)) > len(normalizePath(matches[b].Path))
	})
	var header strings.Builder
	for _, c := range matches {
		if header.Len() > 0 {
			header.WriteString("; ")
		}
		header.WriteString(c.Name)
		header.WriteByte('=')
		header.WriteString(c.Value)
	}
	return header.String()
}

func (c cookie) sameIdentifier(other cookie) bool {
	return c.Name == other.Name &&
		normalizeDomain(c.Domain) == normalizeDomain(other.Domain) &&
		normalizePath(c.Path) == normalizePath(other.Path)
}

func normalizeDomain(domain string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimPrefix(domain, ".")), ".")
}

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	return path
}

func domainMatches(domain, host string) bool {
	domain = normalizeDomain(domain)
	host = normalizeDomain(host)
	if domain == "" {
		return false
	}
	if host == domain {
		return true
	}
	if net.ParseIP(host) != nil || net.ParseIP(domain) != nil {
		return false
	}
	return strings.HasSuffix(host, "."+domain)
}

func pathMatches(cookiePath, requestPath string) bool {
	cookiePath = normalizePath(cookiePath)
	if requestPath == cookiePath {
		return true
	}
	if !strings.HasPrefix(requestPath, cookiePath) {
		return false
	}
	return strings.HasSuffix(cookiePath, "/") || requestPath[len(cookiePath)] == '/'
}
