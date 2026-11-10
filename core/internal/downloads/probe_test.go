package downloads

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

var probeBody = bytes.Repeat([]byte("0123456789"), 1000)

func TestProbeReadsRangeSupportAndName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", `"v1"`)
		w.Header().Set("Content-Disposition", `attachment; filename="Game-1.0.zip"`)
		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(probeBody))
	}))
	defer srv.Close()

	got, err := Probe(context.Background(), srv.Client(), srv.URL+"/file", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Ranges || got.Size != int64(len(probeBody)) || got.ETag != `"v1"` {
		t.Fatalf("unexpected probe: %+v", got)
	}
	if name := filenameFromDisposition(got.ContentDisposition); name != "Game-1.0.zip" {
		t.Fatalf("name from disposition = %q", name)
	}
}

func TestProbeWithoutRangeSupport(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.Itoa(len(probeBody)))
		w.Write(probeBody)
	}))
	defer srv.Close()

	got, err := Probe(context.Background(), srv.Client(), srv.URL+"/file", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Ranges {
		t.Fatal("server without ranges reported ranges")
	}
	if got.Size != int64(len(probeBody)) {
		t.Fatalf("size = %d, want %d", got.Size, len(probeBody))
	}
}

func TestProbeRejectsWebPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("<html>captcha</html>"))
	}))
	defer srv.Close()

	_, err := Probe(context.Background(), srv.Client(), srv.URL+"/page", nil)
	var webPage *WebPageError
	if !errors.As(err, &webPage) {
		t.Fatalf("want WebPageError, got %v", err)
	}
}

func TestProbeFollowsRedirectAndDropsCookieOnOtherSite(t *testing.T) {
	var mu sync.Mutex
	sawCookie := ""
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		sawCookie += r.Header.Get("Cookie")
		mu.Unlock()
		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(probeBody))
	}))
	defer target.Close()

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, strings.Replace(target.URL, "127.0.0.1", "localhost", 1)+"/final.bin", http.StatusFound)
	}))
	defer origin.Close()

	headers := map[string]string{"Cookie": "session=secret"}
	got, err := Probe(context.Background(), origin.Client(), origin.URL+"/start", headers)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(got.URL, "/final.bin") {
		t.Fatalf("final URL = %q, want the redirect target", got.URL)
	}
	mu.Lock()
	defer mu.Unlock()
	if sawCookie != "" {
		t.Fatalf("cookie leaked to another site: %q", sawCookie)
	}
}

func TestSizeOfPrefersContentRangeTotal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Range", "bytes 0-0/4096")
		w.Header().Set("Content-Length", "1")
		w.WriteHeader(http.StatusPartialContent)
		w.Write([]byte{0})
	}))
	defer srv.Close()

	res, err := srv.Client().Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if size := sizeOf(res, 0); size != 4096 {
		t.Fatalf("sizeOf = %d, want 4096", size)
	}
}
