package library

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Cover and screenshot cache. Images are stored as downloaded, without re-encoding.

const (
	maxImageBytes   = 12 << 20
	maxScreenshots  = 8
	imageTimeout    = 30 * time.Second
	userAgentString = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
)

var (
	imageExtByType = map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
		"image/gif":  ".gif",
		"image/avif": ".avif",
	}
	imageExtInURL = regexp.MustCompile(`(?i)\.(jpe?g|png|webp|gif|avif)(?:[?#]|$)`)
	cacheExts     = []string{".jpg", ".png", ".webp", ".gif", ".avif"}
)

type mediaCache struct {
	dir    string
	client *http.Client
}

func newMediaCache(dir string) *mediaCache {
	return &mediaCache{dir: dir, client: &http.Client{Timeout: imageTimeout}}
}

// cacheImage returns the local path of an image, downloading it once. kind is "covers" or "screenshots".
func (m *mediaCache) cacheImage(ctx context.Context, rawURL, kind string) (string, error) {
	if !isHTTPURL(rawURL) {
		return "", fmt.Errorf("not an http url: %s", rawURL)
	}
	if m.dir == "" {
		return "", errors.New("media folder is not set")
	}
	target := filepath.Join(m.dir, kind)
	hash := urlHash(rawURL)
	if cached := findCached(target, hash); cached != "" {
		return cached, nil
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return "", err
	}
	body, contentType, err := m.fetch(ctx, rawURL)
	if err != nil {
		return "", err
	}
	file := filepath.Join(target, hash+imageExt(rawURL, contentType))
	if err := os.WriteFile(file, body, 0o644); err != nil {
		return "", err
	}
	return file, nil
}

// localPath returns a cached file as a path relative to the media folder, or nil when it is not there.
func (m *mediaCache) localPath(path *string) *string {
	if path == nil {
		return nil
	}
	rel, ok := m.relPath(*path)
	if !ok {
		return nil
	}
	return &rel
}

// relPath returns path relative to the media folder, with slashes, or false when path is outside it.
func (m *mediaCache) relPath(path string) (string, bool) {
	if m.dir == "" {
		return "", false
	}
	rel, err := filepath.Rel(m.dir, path)
	if err != nil || rel == "." || rel == ".." || filepath.IsAbs(rel) ||
		strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return filepath.ToSlash(rel), true
}

func (m *mediaCache) fetch(ctx context.Context, rawURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", userAgentString)
	if parsed, err := url.Parse(rawURL); err == nil {
		req.Header.Set("Referer", parsed.Scheme+"://"+parsed.Host+"/")
	}
	res, err := m.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("image request failed: %s", res.Status)
	}
	contentType := res.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, "image/") {
		return nil, "", fmt.Errorf("not an image: %s", contentType)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, maxImageBytes+1))
	if err != nil {
		return nil, "", err
	}
	if len(body) == 0 || len(body) > maxImageBytes {
		return nil, "", errors.New("image is empty or too large")
	}
	return body, contentType, nil
}

// findCached returns the cached file for a URL hash, whatever its extension.
func findCached(dir, hash string) string {
	for _, ext := range cacheExts {
		path := filepath.Join(dir, hash+ext)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func imageExt(rawURL, contentType string) string {
	if ext, ok := imageExtByType[contentType]; ok {
		return ext
	}
	if match := imageExtInURL.FindStringSubmatch(rawURL); match != nil {
		ext := "." + strings.ToLower(match[1])
		if ext == ".jpeg" {
			return ".jpg"
		}
		return ext
	}
	return ".jpg"
}

func isHTTPURL(rawURL string) bool {
	lower := strings.ToLower(rawURL)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

func urlHash(rawURL string) string {
	sum := sha1.Sum([]byte(rawURL))
	return hex.EncodeToString(sum[:])[:20]
}

// AttachCover caches the game cover and stores its local path on the row.
func (l *Library) AttachCover(ctx context.Context, game Game) (Game, error) {
	if game.CoverURL == nil || game.CoverPath != nil {
		return game, nil
	}
	path, err := l.media.cacheImage(ctx, *game.CoverURL, "covers")
	if err != nil {
		return game, err
	}
	return l.Patch(ctx, game.ID, Patch{CoverPath: &path})
}

// CoverFile returns the cached cover of a game, when the file is on disk inside the media folder.
func (l *Library) CoverFile(id int64) (string, bool) {
	game, err := l.Get(context.Background(), id)
	if err != nil || game.CoverPath == nil {
		return "", false
	}
	if _, ok := l.media.relPath(*game.CoverPath); !ok {
		return "", false
	}
	if _, err := os.Stat(*game.CoverPath); err != nil {
		return "", false
	}
	return *game.CoverPath, true
}

// CacheScreenshots downloads up to eight screenshots. Failures are logged and skipped.
func (l *Library) CacheScreenshots(ctx context.Context, game Game) {
	shots := game.Screenshots
	if len(shots) > maxScreenshots {
		shots = shots[:maxScreenshots]
	}
	for _, shot := range shots {
		if _, err := l.media.cacheImage(ctx, shot, "screenshots"); err != nil {
			warnUnlessStopped(ctx, "screenshot %s: %v", shot, err)
		}
	}
}
