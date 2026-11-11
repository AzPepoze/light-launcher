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
	path, ok := s.cfg.Media.CoverFile(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, path)
}

// handleScreenshot serves one cached screenshot of a game by its position in the stored list.
func (s *server) handleScreenshot(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	index, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	path, ok := s.cfg.Media.ScreenshotFile(id, index)
	if !ok {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, path)
}
