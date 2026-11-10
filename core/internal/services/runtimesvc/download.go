package runtimesvc

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

const userAgent = "LightLauncher-App"

// releaseClient limits the whole GitHub API request to 15 seconds.
var releaseClient = &http.Client{Timeout: 15 * time.Second}

// downloadClient has no total time limit, so large archives can finish. Only the wait for headers is limited.
var downloadClient = newDownloadClient()

func newDownloadClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = 30 * time.Second
	return &http.Client{Transport: transport}
}

// downloadFile saves url to destination. onProgress gets the bytes so far and the total (0 when unknown).
func downloadFile(url string, destination string, onProgress func(current, total int64)) error {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", userAgent)

	response, err := downloadClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Download failed with status %d", response.StatusCode)
	}

	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	reader := &progressReader{
		reader:     response.Body,
		total:      max(response.ContentLength, 0),
		onProgress: onProgress,
	}
	if _, err := io.Copy(file, reader); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

// progressReader reports how many bytes have been read after every read.
type progressReader struct {
	reader     io.Reader
	total      int64
	current    int64
	onProgress func(current, total int64)
}

func (r *progressReader) Read(buffer []byte) (int, error) {
	count, err := r.reader.Read(buffer)
	r.current += int64(count)
	if count > 0 && r.onProgress != nil {
		r.onProgress(r.current, r.total)
	}
	return count, err
}

// requestError turns an HTTP client error into the message the UI shows.
func requestError(err error) error {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return errors.New("Request timed out connecting to GitHub")
	}
	return fmt.Errorf("Network error: %s", err.Error())
}

func removeIfExists(target string) {
	_ = os.Remove(target)
}
