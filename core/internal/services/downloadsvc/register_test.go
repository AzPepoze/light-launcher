package downloadsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/downloads"
)

func newRouter(t *testing.T) *api.Router {
	t.Helper()
	store := downloads.NewMemoryStore()
	m, err := downloads.NewManager(downloads.Options{Store: store, Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		m.Shutdown(ctx)
	})
	r := api.NewRouter()
	Register(r, nil, m)
	return r
}

func dispatch(t *testing.T, r *api.Router, method, payload string) any {
	t.Helper()
	out, err := r.Dispatch(context.Background(), method, json.RawMessage(payload))
	if err != nil {
		t.Fatalf("%s: %v", method, err)
	}
	return out
}

func TestRegisterAddsAllDownloadMethods(t *testing.T) {
	r := newRouter(t)
	want := []string{
		"DownloadsList", "DownloadEnqueue", "DownloadPause", "DownloadResume", "DownloadCancel",
		"DownloadRetry", "DownloadRemove", "DownloadSetSpeedLimit", "DownloadsGetSettings", "DownloadsSetSettings",
	}
	got := r.Methods()
	for _, name := range want {
		if !slices.Contains(got, name) {
			t.Errorf("method %s is not registered", name)
		}
	}
	if len(got) != len(want) {
		t.Errorf("registered %d methods, want %d", len(got), len(want))
	}
}

func TestEnqueueThenListShowsFinishedDownload(t *testing.T) {
	body := bytes.Repeat([]byte("g"), 50_000)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Accept-Ranges", "bytes")
		http.ServeContent(w, r, "game.zip", time.Time{}, bytes.NewReader(body))
	}))
	defer srv.Close()

	r := newRouter(t)
	view := dispatch(t, r, "DownloadEnqueue", `{"url":"`+srv.URL+`/game.zip"}`).(downloads.View)
	if view.State == "" || view.ID == 0 {
		t.Fatalf("enqueue returned %+v", view)
	}

	deadline := time.Now().Add(10 * time.Second)
	for {
		list := dispatch(t, r, "DownloadsList", "null").([]downloads.View)
		if len(list) == 1 && list[0].State == downloads.StateDone {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("download did not finish: %+v", list)
		}
		time.Sleep(10 * time.Millisecond)
	}

	raw, err := json.Marshal(dispatch(t, r, "DownloadsList", "null"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"state":"done"`) || strings.Contains(string(raw), "headers") {
		t.Fatalf("unexpected JSON shape: %s", raw)
	}
}

func TestActionOnUnknownDownloadReturnsError(t *testing.T) {
	r := newRouter(t)
	if _, err := r.Dispatch(context.Background(), "DownloadPause", json.RawMessage(`{"id":99}`)); err == nil {
		t.Fatal("pausing an unknown download must fail")
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	r := newRouter(t)
	got := dispatch(t, r, "DownloadsSetSettings", `{"connections":4,"maxActive":2,"hostConnections":6,"retries":1}`).(downloads.Settings)
	if got.Connections != 4 || got.MaxActive != 2 {
		t.Fatalf("set settings = %+v", got)
	}
	read := dispatch(t, r, "DownloadsGetSettings", "null").(downloads.Settings)
	if read != got {
		t.Fatalf("get settings = %+v, want %+v", read, got)
	}
	limited := dispatch(t, r, "DownloadSetSpeedLimit", `{"speedLimitKBps":128}`).(downloads.Settings)
	if limited.SpeedLimitKBps != 128 {
		t.Fatalf("speed limit = %d", limited.SpeedLimitKBps)
	}
}

func TestRemoveReturnsOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "a.zip", time.Time{}, bytes.NewReader([]byte("abc")))
	}))
	defer srv.Close()
	r := newRouter(t)
	dispatch(t, r, "DownloadEnqueue", `{"url":"`+srv.URL+`/a.zip"}`)
	ok := dispatch(t, r, "DownloadRemove", `{"id":1,"deleteFile":true}`).(okResponse)
	if !ok.OK {
		t.Fatal("remove should report ok")
	}
}
