package appsvc

import "sync"

// maxMemoryIcons bounds the memory cache. The oldest entry is dropped first.
const maxMemoryIcons = 1000

type memoryIcon struct {
	stamp fileStamp
	icon  string
}

type memoryIconCache struct {
	mu      sync.Mutex
	limit   int
	entries map[string]memoryIcon
	order   []string
}

func newMemoryIconCache(limit int) *memoryIconCache {
	return &memoryIconCache{limit: limit, entries: map[string]memoryIcon{}}
}

// get returns the cached icon only when the file has not changed since it was stored.
func (c *memoryIconCache) get(path string, stamp fileStamp) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[path]
	if !ok || entry.stamp != stamp {
		return "", false
	}
	return entry.icon, true
}

func (c *memoryIconCache) remember(path string, stamp fileStamp, icon string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= c.limit && len(c.order) > 0 {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.entries, oldest)
	}
	if _, exists := c.entries[path]; !exists {
		c.order = append(c.order, path)
	}
	c.entries[path] = memoryIcon{stamp: stamp, icon: icon}
}
