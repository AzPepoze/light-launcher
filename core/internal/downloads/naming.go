package downloads

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"

	"light-launcher/core/internal/contract"
)

// checkLink accepts only web links. The error never includes the link, since it may carry a token.
func checkLink(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errors.New("the download link must start with http:// or https://")
	}
	return nil
}

// connectionsFor is the connection count a row starts with: one unless the server supports ranges.
func connectionsFor(probe ProbeResult, configured int) int {
	if !probe.Ranges || probe.Size <= 0 {
		return 1
	}
	return ClampConnections(configured)
}

// downloadName picks the saved file name: the server's name when it is useful, then the requested one, then the URL's.
func downloadName(requested string, probe ProbeResult) string {
	name := usefulFilename(filenameFromDisposition(probe.ContentDisposition))
	if name == "" {
		name = usefulFilename(requested)
	}
	if name == "" {
		name = usefulFilename(path.Base(urlPath(probe.URL)))
	}
	if name == "" {
		name = fallbackFilename
	}
	return sanitizeFilename(ensureFilenameExtension(name, probe.ContentType))
}

func urlPath(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Path
}

// downloadFolder returns the absolute folder a request saves into, creating any folders it needs.
func (m *Manager) downloadFolder(req contract.DownloadRequest) (string, error) {
	root, err := m.rootFolder(req.Root)
	if err != nil {
		return "", err
	}
	return resolveFolder(root, req.Subfolder)
}

// rootFolder maps a root name to its folder. An empty name means the downloads folder.
func (m *Manager) rootFolder(root string) (string, error) {
	switch root {
	case "", contract.RootDownloads:
		return m.dir, nil
	case contract.RootGames:
		if m.gamesDir == "" {
			return "", badFolder("the games folder is not set")
		}
		return m.gamesDir, nil
	}
	return "", badFolder(fmt.Sprintf("unknown root %q", root))
}

// freeDestLocked returns a path in dir that no row and no file uses yet.
func (m *Manager) freeDestLocked(ctx context.Context, dir, name string) (string, error) {
	rows, err := m.store.List(ctx)
	if err != nil {
		return "", err
	}
	taken := make(map[string]bool, len(rows))
	for _, d := range rows {
		taken[d.Dest] = true
	}
	ext := filepath.Ext(name)
	stem := name[:len(name)-len(ext)]
	for n := 1; ; n++ {
		candidate := filepath.Join(dir, name)
		if n > 1 {
			candidate = filepath.Join(dir, fmt.Sprintf("%s (%d)%s", stem, n, ext))
		}
		if !taken[candidate] && !fileExists(candidate) {
			return candidate, nil
		}
	}
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// waitGroup waits for wg, or gives up when ctx ends first.
func waitGroup(ctx context.Context, wg *sync.WaitGroup) error {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// plainError drops the request URL that net/http adds to errors, since it may carry a token.
func plainError(err error) string {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return urlErr.Err.Error()
	}
	return err.Error()
}
