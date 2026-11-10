package plugins

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	maxPackageBytes = 256 << 20
	maxRedirects    = 5
	downloadTimeout = 5 * time.Minute
)

// newDownloadClient returns the client for package downloads. It only follows https redirects.
func newDownloadClient() *http.Client {
	return &http.Client{
		Timeout: downloadTimeout,
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return errors.New("too many redirects")
			}
			if request.URL.Scheme != "https" {
				return errors.New("package redirect must use https")
			}
			return nil
		},
	}
}

// isURL reports whether a package source is a web address rather than a local path.
func isURL(source string) bool {
	return strings.Contains(source, "://")
}

// fetchPackage downloads an https package into dir and returns the file path.
func fetchPackage(ctx context.Context, client *http.Client, rawURL, dir string) (string, error) {
	if !strings.HasPrefix(rawURL, "https://") {
		return "", errors.New("package URL must start with https://")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("download package: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download package: HTTP %d", response.StatusCode)
	}
	return writeDownload(response.Body, filepath.Join(dir, "package.llplugin"))
}

func writeDownload(body io.Reader, target string) (string, error) {
	file, err := os.Create(target)
	if err != nil {
		return "", err
	}
	written, err := io.Copy(file, io.LimitReader(body, maxPackageBytes+1))
	closeErr := file.Close()
	if err != nil {
		return "", fmt.Errorf("download package: %w", err)
	}
	if written > maxPackageBytes {
		return "", fmt.Errorf("package is larger than %d bytes", maxPackageBytes)
	}
	if closeErr != nil {
		return "", closeErr
	}
	return target, nil
}
