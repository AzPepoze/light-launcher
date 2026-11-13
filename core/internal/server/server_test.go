package server

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/platform"
	"light-launcher/core/internal/uploads"
)

const (
	testToken = "test-token"
	testPort  = 4321
	testHost  = "127.0.0.1:4321"
)

// fakeMedia maps game ids to cached cover and screenshot files.
type fakeMedia struct {
	covers      map[int64]string
	screenshots map[int64]map[int]string
}

func (m *fakeMedia) CoverFile(id int64) (string, bool) {
	path, ok := m.covers[id]
	return path, ok
}

func (m *fakeMedia) ScreenshotFile(id int64, index int) (string, bool) {
	path, ok := m.screenshots[id][index]
	return path, ok
}

// fixture is a server with an in-memory UI and a record of opened game paths.
type fixture struct {
	handler   http.Handler
	router    *api.Router
	hub       *api.Hub
	media     *fakeMedia
	uploadDir string
	opened    []string
}

func newFixture(t *testing.T, port int) *fixture {
	t.Helper()
	fx := &fixture{
		router: api.NewRouter(),
		hub:    api.NewHub(),
		media:  &fakeMedia{covers: map[int64]string{}, screenshots: map[int64]map[int]string{}},
	}
	fx.uploadDir = filepath.Join(t.TempDir(), "uploads")
	uploadStore, err := uploads.Open(fx.uploadDir)
	if err != nil {
		t.Fatal(err)
	}
	ui := fstest.MapFS{
		"index.html": {Data: []byte("<!doctype html><title>ui</title>")},
		"app.js":     {Data: []byte("console.log('app')")},
	}
	fx.handler = New(Config{
		UI:      ui,
		Router:  fx.router,
		Hub:     fx.hub,
		Media:   fx.media,
		Uploads: uploadStore,
		Token:   testToken,
		Port:    port,
		OnOpen: func(gamePath string) error {
			fx.opened = append(fx.opened, gamePath)
			return nil
		},
	})
	return fx
}

// request builds a request that passes every check. Tests break one check at a time.
func (fx *fixture) request(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Host = testHost
	req.RemoteAddr = "127.0.0.1:50000"
	req.Header.Set("X-LL-Token", testToken)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func (fx *fixture) serve(req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	fx.handler.ServeHTTP(rec, req)
	return rec
}

func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) envelope {
	t.Helper()
	var env envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode envelope %q: %v", rec.Body.String(), err)
	}
	return env
}

func TestGuardRejectsBadRequests(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*http.Request)
		want   int
	}{
		{"missing token", func(r *http.Request) { r.Header.Del(tokenHeader) }, http.StatusUnauthorized},
		{"wrong token", func(r *http.Request) { r.Header.Set(tokenHeader, "nope") }, http.StatusUnauthorized},
		{"bad origin", func(r *http.Request) { r.Header.Set("Origin", "http://evil.example") }, http.StatusForbidden},
		{"bad host", func(r *http.Request) { r.Host = "evil.example" }, http.StatusForbidden},
		{"remote peer", func(r *http.Request) { r.RemoteAddr = "192.0.2.10:1234" }, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fx := newFixture(t, testPort)
			req := fx.request(http.MethodPost, "/api/invoke", `{"method":"Ping"}`)
			tc.mutate(req)
			if rec := fx.serve(req); rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestGuardAcceptsServerOrigin(t *testing.T) {
	fx := newFixture(t, testPort)
	fx.router.Handle("Ping", api.NoInput(func(context.Context) (string, error) { return "pong", nil }))
	req := fx.request(http.MethodPost, "/api/invoke", `{"method":"Ping"}`)
	req.Header.Set("Origin", "http://"+testHost)
	if rec := fx.serve(req); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestInvokeSuccessAndHandlerError(t *testing.T) {
	fx := newFixture(t, testPort)
	fx.router.Handle("Echo", api.Typed(func(_ context.Context, in map[string]string) (map[string]string, error) {
		return in, nil
	}))
	fx.router.Handle("Fail", api.NoInput(func(context.Context) (string, error) {
		return "", errors.New("disk full")
	}))

	rec := fx.serve(fx.request(http.MethodPost, "/api/invoke", `{"method":"Echo","payload":{"name":"x"}}`))
	env := decodeEnvelope(t, rec)
	data, _ := env.Data.(map[string]any)
	if rec.Code != http.StatusOK || !env.Success || data["name"] != "x" {
		t.Fatalf("echo: status %d, envelope %+v", rec.Code, env)
	}

	env = decodeEnvelope(t, fx.serve(fx.request(http.MethodPost, "/api/invoke", `{"method":"Fail"}`)))
	if env.Success || env.Error != "disk full" {
		t.Fatalf("fail: envelope %+v", env)
	}
}

func TestInvokeUsesTheMethodTimeout(t *testing.T) {
	fx := newFixture(t, testPort)
	var plain, long time.Duration
	fx.router.Handle("Plain", api.NoInput(func(ctx context.Context) (string, error) {
		deadline, _ := ctx.Deadline()
		plain = time.Until(deadline)
		return "", nil
	}))
	fx.router.HandleWithTimeout("Long", 11*time.Minute, api.NoInput(func(ctx context.Context) (string, error) {
		deadline, _ := ctx.Deadline()
		long = time.Until(deadline)
		return "", nil
	}))

	fx.serve(fx.request(http.MethodPost, "/api/invoke", `{"method":"Plain"}`))
	fx.serve(fx.request(http.MethodPost, "/api/invoke", `{"method":"Long"}`))
	if plain > invokeTimeout || plain < invokeTimeout-time.Minute {
		t.Errorf("plain method deadline in %v, want about %v", plain, invokeTimeout)
	}
	if long < 10*time.Minute {
		t.Errorf("long method deadline in %v, want about 11m", long)
	}
}

func TestInvokeUnknownMethod(t *testing.T) {
	fx := newFixture(t, testPort)
	env := decodeEnvelope(t, fx.serve(fx.request(http.MethodPost, "/api/invoke", `{"method":"Nope"}`)))
	if env.Success || !strings.Contains(env.Error, "unknown method") || env.Code != "method_not_found" {
		t.Fatalf("envelope %+v", env)
	}
}

func TestInvokeBadBody(t *testing.T) {
	fx := newFixture(t, testPort)
	rec := fx.serve(fx.request(http.MethodPost, "/api/invoke", `{"method":`))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestInvokeBodyLimit(t *testing.T) {
	fx := newFixture(t, testPort)
	body := `{"method":"Ping","payload":"` + strings.Repeat("a", maxInvokeBody) + `"}`
	rec := fx.serve(fx.request(http.MethodPost, "/api/invoke", body))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", rec.Code)
	}
}

func TestOpenPassesGamePath(t *testing.T) {
	fx := newFixture(t, testPort)
	rec := fx.serve(fx.request(http.MethodPost, "/api/open", `{"path":"/games/demo.exe"}`))
	if rec.Code != http.StatusOK || !decodeEnvelope(t, rec).Success {
		t.Fatalf("status %d, body %s", rec.Code, rec.Body.String())
	}
	if len(fx.opened) != 1 || fx.opened[0] != "/games/demo.exe" {
		t.Fatalf("opened = %v", fx.opened)
	}
}

func TestOpenRequiresPath(t *testing.T) {
	fx := newFixture(t, testPort)
	rec := fx.serve(fx.request(http.MethodPost, "/api/open", `{}`))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestPlatformReportsOS(t *testing.T) {
	fx := newFixture(t, testPort)
	rec := fx.serve(fx.request(http.MethodGet, "/api/platform", ""))
	var info platformInfo
	if err := json.Unmarshal(rec.Body.Bytes(), &info); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if rec.Code != http.StatusOK || info.OS != platform.OS() {
		t.Fatalf("status %d, info %+v", rec.Code, info)
	}
}

func TestUnknownAPIPathIsNotFound(t *testing.T) {
	fx := newFixture(t, testPort)
	rec := fx.serve(fx.request(http.MethodGet, "/api/nothing", ""))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestAPIResponsesAreNotCached(t *testing.T) {
	fx := newFixture(t, testPort)
	rec := fx.serve(fx.request(http.MethodGet, "/api/platform", ""))
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q", got)
	}
}

// newLiveFixture serves the fixture on a real loopback port, so Host and SSE behave as in production.
func newLiveFixture(t *testing.T) (*fixture, string) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	fx := newFixture(t, port)
	ts := httptest.NewUnstartedServer(fx.handler)
	ts.Listener.Close()
	ts.Listener = listener
	ts.Start()
	t.Cleanup(ts.Close)
	return fx, ts.URL
}
