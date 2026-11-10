package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/logger"
)

// heartbeatEvery keeps idle streams alive through proxies. Tests shorten it.
var heartbeatEvery = 15 * time.Second

// handleEvents streams hub events as server-sent events until the client goes away.
func (s *server) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	events, unsubscribe := s.cfg.Hub.Subscribe()
	defer unsubscribe()

	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ticker := time.NewTicker(heartbeatEvery)
	defer ticker.Stop()
	for {
		var err error
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			_, err = fmt.Fprint(w, ": heartbeat\n\n")
		case event := <-events:
			err = writeEvent(w, event)
		}
		if err != nil {
			return
		}
		flusher.Flush()
	}
}

// writeEvent writes one event. An event that cannot be encoded is skipped, not fatal.
func writeEvent(w io.Writer, event api.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		logger.Warn("events", "cannot encode %q event: %v", event.Channel, err)
		return nil
	}
	_, err = fmt.Fprintf(w, "data: %s\n\n", data)
	return err
}
