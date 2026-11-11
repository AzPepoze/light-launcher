package server

import (
	"net/http"
	"strconv"
)

// handleCover serves the cached cover of one game. The file path comes from the library, never from the request.
func (s *server) handleCover(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	path, ok := s.cfg.Covers.CoverFile(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, path)
}
