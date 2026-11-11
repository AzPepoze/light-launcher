package main

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

type shellEvent struct {
	Event    string            `json:"event"`
	URL      string            `json:"url,omitempty"`
	Filename *string           `json:"filename,omitempty"`
	Size     *int64            `json:"size,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
	Message  string            `json:"message,omitempty"`
}

// eventWriter writes one flushed JSON line per event.
type eventWriter struct {
	mu  sync.Mutex
	out *bufio.Writer
}

func newEventWriter(w io.Writer) *eventWriter {
	return &eventWriter{out: bufio.NewWriter(w)}
}

func (w *eventWriter) write(event shellEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode event: %w", err)
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, err := w.out.Write(data); err != nil {
		return fmt.Errorf("write event: %w", err)
	}
	if err := w.out.WriteByte('\n'); err != nil {
		return fmt.Errorf("write event: %w", err)
	}
	if err := w.out.Flush(); err != nil {
		return fmt.Errorf("flush event: %w", err)
	}
	return nil
}

func readyEvent() shellEvent {
	return shellEvent{Event: eventReady}
}

func closedEvent() shellEvent {
	return shellEvent{Event: eventClosed}
}

func errorEvent(message string) shellEvent {
	return shellEvent{Event: eventError, Message: message}
}

func downloadEvent(rawURL, filename string, total int64, headers map[string]string) shellEvent {
	size := total
	if size < 0 {
		size = -1
	}
	return shellEvent{
		Event:    eventDownload,
		URL:      rawURL,
		Filename: &filename,
		Size:     &size,
		Headers:  headers,
	}
}

func writeError(out io.Writer, message string) {
	if err := newEventWriter(out).write(errorEvent(message)); err != nil {
		log.Error("assist", "cannot write error event: %v", err)
	}
}
