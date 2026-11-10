package plugins

import (
	"context"
	"strings"
	"sync"
	"time"
)

const (
	hostRequestInterval = 250 * time.Millisecond
	maxCachedEntries    = 256
)

// hostMatches reports whether a permission entry covers host.
// "example.com" matches only that host. "*.example.com" matches subdomains, not the bare domain.
func hostMatches(entry, host string) bool {
	entry = strings.ToLower(entry)
	host = strings.ToLower(host)
	if suffix, ok := strings.CutPrefix(entry, "*."); ok {
		return strings.HasSuffix(host, "."+suffix)
	}
	return entry == host
}

// allowsHost reports whether the manifest network list covers host.
func (p Permissions) allowsHost(host string) bool {
	for _, entry := range p.Network {
		if hostMatches(entry, host) {
			return true
		}
	}
	return false
}

// hostLimiter spaces the requests a plugin makes to one host. It is shared by all plugins.
type hostLimiter struct {
	mu       sync.Mutex
	interval time.Duration
	next     map[string]time.Time
}

func newHostLimiter(interval time.Duration) *hostLimiter {
	return &hostLimiter{interval: interval, next: map[string]time.Time{}}
}

// wait reserves the next slot for host and sleeps until it comes, or until ctx ends.
func (l *hostLimiter) wait(ctx context.Context, host string) error {
	delay := l.reserve(host, time.Now())
	if delay <= 0 {
		return nil
	}
	return sleepContext(ctx, delay)
}

func (l *hostLimiter) reserve(host string, now time.Time) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	slot := now
	if next := l.next[host]; next.After(now) {
		slot = next
	}
	l.next[host] = slot.Add(l.interval)
	return slot.Sub(now)
}

// etagCache remembers the ETag of each cache key, so a plugin can get "not modified" answers.
type etagCache struct {
	mu    sync.Mutex
	etags map[string]string
}

func newETagCache() *etagCache {
	return &etagCache{etags: map[string]string{}}
}

func (c *etagCache) get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	etag, ok := c.etags[key]
	return etag, ok
}

// put stores an ETag. The cache is cleared when it is full, which keeps memory bounded.
func (c *etagCache) put(key, etag string) {
	if etag == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.etags) >= maxCachedEntries {
		c.etags = map[string]string{}
	}
	c.etags[key] = etag
}
