package server

import (
	"crypto/subtle"
	"net"
	"net/http"
	"strings"

	"light-launcher/core/internal/logger"
)

const tokenHeader = "X-LL-Token"

// guard runs the checks every request passes before a route handles it.
func (s *server) guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setSecurityHeaders(w, s.csp)
		isAPI := strings.HasPrefix(r.URL.Path, "/api/")
		if isAPI {
			w.Header().Set("Cache-Control", "no-store")
		}
		if !isLoopback(r.RemoteAddr) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if r.Host != s.host {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" && origin != s.origin {
			logger.Warn("auth", "rejected origin %q", origin)
			http.Error(w, "forbidden origin", http.StatusForbidden)
			return
		}
		if isAPI && !s.hasValidToken(r) {
			logger.Warn("auth", "rejected request without a valid token on %s", r.URL.Path)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// hasValidToken compares in constant time so the token cannot be guessed byte by byte.
func (s *server) hasValidToken(r *http.Request) bool {
	given := []byte(r.Header.Get(tokenHeader))
	return subtle.ConstantTimeCompare(given, []byte(s.cfg.Token)) == 1
}

func isLoopback(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func setSecurityHeaders(w http.ResponseWriter, csp string) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", csp)
}
