package server

import (
	"net/http"

	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/platform"
)

type platformInfo struct {
	OS       string            `json:"os"`
	Features platform.Features `json:"features"`
}

// handleReady logs the first time the UI reports that it has loaded.
func (s *server) handleReady(w http.ResponseWriter, _ *http.Request) {
	s.readyOnce.Do(func() {
		logger.Info("ready", "page loaded")
	})
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) handlePlatform(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, platformInfo{OS: platform.OS(), Features: platform.Current()})
}
