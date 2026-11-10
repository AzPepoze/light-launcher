package server

import "net/http"

const maxOpenBody = 64 << 10

type openRequest struct {
	Path string `json:"path"`
}

// handleOpen takes a game path from a second launch and hands it to the launcher.
func (s *server) handleOpen(w http.ResponseWriter, r *http.Request) {
	var req openRequest
	if !decodeBody(w, r, maxOpenBody, &req) {
		return
	}
	if req.Path == "" {
		writeJSON(w, http.StatusBadRequest, envelope{Error: "missing path"})
		return
	}
	if err := s.cfg.OnOpen(req.Path); err != nil {
		writeJSON(w, http.StatusOK, envelope{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, envelope{Success: true})
}
