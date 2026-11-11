package assist

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

const (
	eventReady    = "ready"
	eventDownload = "download"
	eventClosed   = "closed"
	eventError    = "error"
)

// Event is one line of the assist NDJSON output.
type Event struct {
	Event    string             `json:"event"`
	URL      string             `json:"url,omitempty"`
	Filename *string            `json:"filename,omitempty"`
	Size     *int64             `json:"size,omitempty"`
	Headers  *map[string]string `json:"headers,omitempty"`
	Message  string             `json:"message,omitempty"`
}

// Ready returns the event that reports the window is shown.
func Ready() Event {
	return Event{Event: eventReady}
}

// Closed returns the event that reports the window closed without a download.
func Closed() Event {
	return Event{Event: eventClosed}
}

// Error returns the fatal event for a message.
func Error(message string) Event {
	return Event{Event: eventError, Message: message}
}

// Download returns a download event. An empty header set is omitted from the line.
func Download(rawURL, filename string, size int64, headers map[string]string) Event {
	return newDownload(rawURL, filename, size, headers, false)
}

// DownloadWithHeaders returns a download event that always writes the header set,
// even when it is empty.
func DownloadWithHeaders(rawURL, filename string, size int64, headers map[string]string) Event {
	return newDownload(rawURL, filename, size, headers, true)
}

func newDownload(rawURL, filename string, size int64, headers map[string]string, keepEmptyHeaders bool) Event {
	if size < 0 {
		size = -1
	}
	event := Event{Event: eventDownload, URL: rawURL, Filename: &filename, Size: &size}
	if len(headers) > 0 || keepEmptyHeaders {
		if headers == nil {
			headers = map[string]string{}
		}
		event.Headers = &headers
	}
	return event
}

// Writer writes one JSON object per line and flushes every line. It keeps the
// first error and writes nothing after it. Safe for concurrent use.
type Writer struct {
	mu  sync.Mutex
	out *bufio.Writer
	err error
}

// NewWriter returns a writer that emits to out.
func NewWriter(out io.Writer) *Writer {
	return &Writer{out: bufio.NewWriter(out)}
}

// Write emits one event as a single flushed line.
func (w *Writer) Write(event any) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.err != nil {
		return w.err
	}
	data, err := json.Marshal(event)
	if err != nil {
		w.err = fmt.Errorf("encode event: %w", err)
		return w.err
	}
	data = append(data, '\n')
	if _, err := w.out.Write(data); err != nil {
		w.err = fmt.Errorf("write event: %w", err)
		return w.err
	}
	if err := w.out.Flush(); err != nil {
		w.err = fmt.Errorf("flush event: %w", err)
		return w.err
	}
	return nil
}

// Err returns the first write error, or nil.
func (w *Writer) Err() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.err
}
