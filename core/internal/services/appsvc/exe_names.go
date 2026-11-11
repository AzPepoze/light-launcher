package appsvc

import (
	"os"
	"sync"
	"time"

	"light-launcher/core/internal/config"
)

// versionCache keeps the version strings of each executable by path.
var versionCache = struct {
	mu      sync.Mutex
	entries map[string]cachedVersion
}{entries: map[string]cachedVersion{}}

// cachedVersion is valid while the file keeps the same size and modification time.
type cachedVersion struct {
	size    int64
	modTime time.Time
	names   versionStrings
}

// cachedVersionStrings returns the version strings of an executable. The file is read only when it is new or changed.
func cachedVersionStrings(file scannedFile) versionStrings {
	key := config.PathKey(file.Path)
	versionCache.mu.Lock()
	entry, ok := versionCache.entries[key]
	versionCache.mu.Unlock()
	if ok && entry.size == file.Size && entry.modTime.Equal(file.ModTime) {
		return entry.names
	}

	names := readExeVersionStrings(file.Path)
	versionCache.mu.Lock()
	versionCache.entries[key] = cachedVersion{size: file.Size, modTime: file.ModTime, names: names}
	versionCache.mu.Unlock()
	return names
}

// readExeVersionStrings reads the version strings of the file at path. A file that cannot be read has none.
func readExeVersionStrings(path string) versionStrings {
	file, err := os.Open(path)
	if err != nil {
		return versionStrings{}
	}
	defer file.Close()
	return readVersionStrings(file)
}
