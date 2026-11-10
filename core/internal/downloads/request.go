package downloads

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	copyBufferSize = 32 * 1024
	stallTimeout   = 30 * time.Second
)

var errStalled = errors.New("download stalled: no data for 30 seconds")

// request is one GET. A ranged request wants bytes [from, end); an unranged one streams the whole body.
type request struct {
	url     string
	headers map[string]string
	ranged  bool
	from    int64
	end     int64
}

// fetch streams a response body into file. onWrite receives the byte count of every write.
func (t *transferer) fetch(ctx context.Context, rq request, file io.WriterAt, onWrite func(n int64)) error {
	reqCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	stall := time.AfterFunc(stallTimeout, func() { cancel(errStalled) })
	defer stall.Stop()

	res, err := t.send(reqCtx, rq)
	if err != nil {
		return failureOf(reqCtx, err)
	}
	defer res.Body.Close()
	if err := checkStatus(res, rq.ranged); err != nil {
		return err
	}
	err = t.copyBody(reqCtx, res.Body, rq, file, onWrite, func() { stall.Reset(stallTimeout) })
	return failureOf(reqCtx, err)
}

func (t *transferer) send(ctx context.Context, rq request) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rq.url, nil)
	if err != nil {
		return nil, err
	}
	for name, value := range rq.headers {
		req.Header.Set(name, value)
	}
	if rq.ranged {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", rq.from, rq.end-1))
	}
	return t.client.Do(req)
}

// failureOf reports a stall instead of the generic cancel error that the body read returns.
func failureOf(ctx context.Context, err error) error {
	if err != nil && errors.Is(context.Cause(ctx), errStalled) {
		return errStalled
	}
	return err
}

func checkStatus(res *http.Response, ranged bool) error {
	switch {
	case res.StatusCode == http.StatusTooManyRequests || res.StatusCode == http.StatusServiceUnavailable:
		return &rateLimitError{retryAfter: parseRetryAfter(res.Header.Get("Retry-After"), time.Now())}
	case ranged && res.StatusCode != http.StatusPartialContent:
		return fmt.Errorf("range request failed: HTTP %d", res.StatusCode)
	case !ranged && res.StatusCode != http.StatusOK:
		return fmt.Errorf("HTTP %d", res.StatusCode)
	}
	return nil
}

// copyBody writes the body at its offsets, within the speed limit. A ranged copy stops at its end.
func (t *transferer) copyBody(ctx context.Context, body io.Reader, rq request, file io.WriterAt, onWrite func(int64), onActivity func()) error {
	buf := make([]byte, copyBufferSize)
	pos := rq.from
	for {
		if rq.ranged && pos >= rq.end {
			return nil
		}
		chunk := buf
		if rq.ranged {
			chunk = buf[:min(int64(len(buf)), rq.end-pos)]
		}
		n, readErr := body.Read(chunk)
		if n > 0 {
			if err := t.limiter.Wait(ctx, n); err != nil {
				return err
			}
			if _, err := file.WriteAt(chunk[:n], pos); err != nil {
				return err
			}
			pos += int64(n)
			onWrite(int64(n))
			onActivity()
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	if rq.ranged && pos < rq.end {
		return fmt.Errorf("short read (%d of %d bytes)", pos-rq.from, rq.end-rq.from)
	}
	return nil
}
