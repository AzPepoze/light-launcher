package server

import (
	"encoding/json"
	"errors"
	"net/http"
)

// envelope is the reply shape of the invoke and open calls.
type envelope struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// decodeBody reads one JSON value of at most limit bytes. On failure it writes the reply and returns false.
func decodeBody(w http.ResponseWriter, r *http.Request, limit int64, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	err := json.NewDecoder(r.Body).Decode(dst)
	if err == nil {
		return true
	}
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		writeJSON(w, http.StatusRequestEntityTooLarge, envelope{Error: "request body too large"})
	} else {
		writeJSON(w, http.StatusBadRequest, envelope{Error: "bad request body"})
	}
	return false
}
