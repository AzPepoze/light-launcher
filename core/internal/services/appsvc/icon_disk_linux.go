package appsvc

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// iconCacheKeyVersion changes the disk key when the cached format changes. It matches the Electron cache.
const iconCacheKeyVersion = "v2"

type cachedIconFormat struct {
	ext  string
	mime string
}

var cachedIconFormats = []cachedIconFormat{
	{ext: "png", mime: "image/png"},
	{ext: "ico", mime: "image/x-icon"},
}

func iconCacheDirectory() string {
	return filepath.Join(os.TempDir(), "light-launcher-icon-cache")
}

// iconCacheBase returns the path without extension. The key hashes the same text as the Electron cache.
func iconCacheBase(exePath string, stamp fileStamp) string {
	key := fmt.Sprintf("%s|%s|%d|%s", exePath, strconv.FormatFloat(stamp.mtimeMs, 'f', -1, 64), stamp.size, iconCacheKeyVersion)
	sum := sha1.Sum([]byte(key))
	return filepath.Join(iconCacheDirectory(), hex.EncodeToString(sum[:]))
}

// readCachedIcon returns the cached icon. ok is false on a miss; an empty icon means "known to have none".
func readCachedIcon(exePath string, stamp fileStamp) (icon string, ok bool) {
	base := iconCacheBase(exePath, stamp)
	for _, format := range cachedIconFormats {
		data, err := os.ReadFile(base + "." + format.ext)
		if err != nil || len(data) == 0 {
			continue
		}
		return dataURL(format.mime, data), true
	}
	if _, err := os.Stat(base + ".miss"); err == nil {
		return "", true
	}
	return "", false
}

// writeCachedIcon stores the icon bytes. Empty data is not stored.
func writeCachedIcon(exePath string, stamp fileStamp, data []byte, format string) error {
	if len(data) == 0 {
		return nil
	}
	if err := os.MkdirAll(iconCacheDirectory(), 0o755); err != nil {
		return err
	}
	target := iconCacheBase(exePath, stamp) + "." + format
	tempPath := fmt.Sprintf("%s.%d.%d.tmp", target, os.Getpid(), time.Now().UnixMilli())
	if err := os.WriteFile(tempPath, data, 0o644); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	if err := os.Rename(tempPath, target); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	return nil
}

// markCachedMiss records that the executable has no extractable icon, so later runs skip extraction.
func markCachedMiss(exePath string, stamp fileStamp) error {
	if err := os.MkdirAll(iconCacheDirectory(), 0o755); err != nil {
		return err
	}
	return os.WriteFile(iconCacheBase(exePath, stamp)+".miss", nil, 0o644)
}
