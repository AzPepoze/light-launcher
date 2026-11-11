package server

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func writeTestImage(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(name+"-bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestMediaServesCachedCover(t *testing.T) {
	fx := newFixture(t, testPort)
	coverPath := writeTestImage(t, "cover.jpg")
	fx.media.covers[7] = coverPath

	rec := fx.serve(fx.request(http.MethodGet, "/api/media/cover/7", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %q", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); body != "cover.jpg-bytes" {
		t.Errorf("body = %q, want the cover file", body)
	}
}

func TestMediaServesCachedScreenshot(t *testing.T) {
	fx := newFixture(t, testPort)
	fx.media.screenshots[7] = map[int]string{0: writeTestImage(t, "shot.png")}

	rec := fx.serve(fx.request(http.MethodGet, "/api/media/screenshot/7/0", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %q", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); body != "shot.png-bytes" {
		t.Fatalf("body = %q, want the screenshot file", body)
	}
}

func TestMediaIsNotFoundWithoutFile(t *testing.T) {
	fx := newFixture(t, testPort)
	for _, target := range []string{
		"/api/media/cover/7",
		"/api/media/cover/not-a-number",
		"/api/media/screenshot/7/0",
		"/api/media/screenshot/7/not-a-number",
		"/api/media/screenshot/not-a-number/0",
		"/api/media/poster/7",
	} {
		rec := fx.serve(fx.request(http.MethodGet, target, ""))
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s: status = %d, want 404", target, rec.Code)
		}
	}
}

func TestMediaNeedsToken(t *testing.T) {
	fx := newFixture(t, testPort)
	fx.media.covers[7] = writeTestImage(t, "cover.jpg")
	fx.media.screenshots[7] = map[int]string{0: writeTestImage(t, "shot.png")}

	for _, target := range []string{"/api/media/cover/7", "/api/media/screenshot/7/0"} {
		req := fx.request(http.MethodGet, target, "")
		req.Header.Del("X-LL-Token")
		if rec := fx.serve(req); rec.Code != http.StatusUnauthorized {
			t.Errorf("GET %s without token: status = %d, want 401", target, rec.Code)
		}
	}
}
