package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"light-launcher/core/internal/logger"
)

const (
	maxInvokeBody = 1 << 20
	invokeTimeout = 60 * time.Second
)

type invokeRequest struct {
	Method  string          `json:"method"`
	Payload json.RawMessage `json:"payload"`
}

// handleInvoke runs one UI method. Handler errors travel in the envelope with HTTP 200;
// the HTTP status is only for malformed requests.
func (s *server) handleInvoke(w http.ResponseWriter, r *http.Request) {
	var req invokeRequest
	if !decodeBody(w, r, maxInvokeBody, &req) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), invokeTimeout)
	defer cancel()

	data, err := s.cfg.Router.Dispatch(ctx, req.Method, req.Payload)
	if err != nil {
		logger.Warn("api", "%q failed: %v", req.Method, err)
		writeJSON(w, http.StatusOK, envelope{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, envelope{Success: true, Data: data})
}
