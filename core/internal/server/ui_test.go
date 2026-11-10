package server

import (
	"net/http"
	"strings"
	"testing"
)

func TestUIServesFilesAndFallsBackToIndex(t *testing.T) {
	fx := newFixture(t, testPort)
	cases := []struct {
		target string
		want   string
	}{
		{"/app.js", "console.log('app')"},
		{"/", "<title>ui</title>"},
		{"/games/42", "<title>ui</title>"},
	}
	for _, tc := range cases {
		rec := fx.serve(fx.request(http.MethodGet, tc.target, ""))
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), tc.want) {
			t.Fatalf("%s: status %d, body %q", tc.target, rec.Code, rec.Body.String())
		}
	}
}

func TestUIHasSecurityHeaders(t *testing.T) {
	fx := newFixture(t, testPort)
	rec := fx.serve(fx.request(http.MethodGet, "/", ""))
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q", got)
	}
	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "default-src 'self'") || !strings.Contains(csp, "ws://127.0.0.1:4321") {
		t.Fatalf("Content-Security-Policy = %q", csp)
	}
}
