package appsvc

import (
	"context"
	"os"
	"path/filepath"

	"light-launcher/core/internal/logger"
)

// memoryIcons is the first icon cache, checked before the disk cache.
var memoryIcons = newMemoryIconCache(maxMemoryIcons)

// fileStamp identifies one version of a file the way Node's stat does: mtime in milliseconds, plus size.
type fileStamp struct {
	mtimeMs float64
	size    int64
}

func newFileStamp(info os.FileInfo) fileStamp {
	modified := info.ModTime()
	mtimeMs := float64(modified.Unix())*1000 + float64(modified.Nanosecond())/1e6
	return fileStamp{mtimeMs: mtimeMs, size: info.Size()}
}

// extractExeIcon returns the icon of an executable as a data URL, or "" when it has none.
// It checks the memory cache, then the disk cache, and extracts the icon only on a miss.
func extractExeIcon(ctx context.Context, exePath string) (string, error) {
	info, err := os.Stat(exePath)
	if err != nil || !info.Mode().IsRegular() {
		return "", nil
	}
	name := filepath.Base(exePath)
	stamp := newFileStamp(info)

	if icon, ok := memoryIcons.get(exePath, stamp); ok {
		logger.Debug("Icon", "hit memory %s", name)
		return icon, nil
	}
	if icon, ok := readCachedIcon(exePath, stamp); ok {
		if icon != "" {
			memoryIcons.remember(exePath, stamp, icon)
			logger.Debug("Icon", "hit disk %s", name)
		} else {
			logger.Debug("Icon", "hit disk (no icon) %s", name)
		}
		return icon, nil
	}

	logger.Debug("Icon", "miss %s", name)
	return extractAndCacheIcon(ctx, exePath, info, name)
}

func extractAndCacheIcon(ctx context.Context, exePath string, info os.FileInfo, name string) (string, error) {
	tempDir, err := os.MkdirTemp("", "light-launcher-icon-")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	stamp := newFileStamp(info)
	icoPath := extractIcoFile(ctx, exePath, tempDir)
	if icoPath == "" {
		memoryIcons.remember(exePath, stamp, "")
		if err := markCachedMiss(exePath, stamp); err != nil {
			logger.Debug("Icon", "cache write failed: %v", err)
		}
		logger.Debug("Icon", "no icon %s", name)
		return "", nil
	}

	icon, err := buildAndCacheIcon(ctx, exePath, stamp, icoPath, tempDir)
	if err != nil {
		return "", err
	}
	logger.Debug("Icon", "extracted %s (%d bytes)", name, len(icon))
	return icon, nil
}

// buildAndCacheIcon prefers the PNG of the largest ICO frame and keeps the raw ICO when that fails.
func buildAndCacheIcon(ctx context.Context, exePath string, stamp fileStamp, icoPath, tempDir string) (string, error) {
	ico, err := os.ReadFile(icoPath)
	if err != nil {
		return "", err
	}
	data, mime, format := iconPayload(extractIconPNG(ctx, ico, icoPath, tempDir), ico)
	icon := dataURL(mime, data)
	memoryIcons.remember(exePath, stamp, icon)
	if err := writeCachedIcon(exePath, stamp, data, format); err != nil {
		logger.Debug("Icon", "cache write failed: %v", err)
	}
	return icon, nil
}

func iconPayload(png, ico []byte) (data []byte, mime, format string) {
	if png != nil {
		return png, "image/png", "png"
	}
	return ico, "image/x-icon", "ico"
}
