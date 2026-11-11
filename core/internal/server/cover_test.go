package server

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestCoverServesCachedFile(t *testing.T) {
	fx := newFixture(t, testPort)
	coverPath := filepath.Join(t.TempDir(), "cover.jpg")
	if err := os.WriteFile(coverPath, []byte("cover-bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	fx.covers[7] = coverPath

	rec := fx.serve(fx.request(http.MethodGet, "/api/cover/7", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %q", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); body != "cover-bytes" {
		t.Fatalf("body = %q, want the cover file", body)
	}
}

func TestCoverIsNotFoundWithoutFile(t *testing.T) {
	fx := newFixture(t, testPort)
	for _, target := range []string{"/api/cover/7", "/api/cover/not-a-number"} {
		rec := fx.serve(fx.request(http.MethodGet, target, ""))
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s: status = %d, want 404", target, rec.Code)
		}
	}
}

func TestCoverNeedsToken(t *testing.T) {
	fx := newFixture(t, testPort)
	coverPath := filepath.Join(t.TempDir(), "cover.jpg")
	if err := os.WriteFile(coverPath, []byte("cover-bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	fx.covers[7] = coverPath

	req := fx.request(http.MethodGet, "/api/cover/7", "")
	req.Header.Del("X-LL-Token")
	if rec := fx.serve(req); rec.Code != http.StatusUnauthorized {
		t.Fatalf("status without token = %d, want 401", rec.Code)
	}
}
