package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestForwardToRunningWithoutLock(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "launcher.lock")
	if forwardToRunning(lockPath, "/games/demo.exe") {
		t.Fatal("forwarded without a lock file")
	}
}

func TestForwardToRunningIgnoresStaleLock(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	lockPath := filepath.Join(t.TempDir(), "launcher.lock")
	if err := writeInstance(lockPath, instanceInfo{Port: port, Token: "stale"}); err != nil {
		t.Fatal(err)
	}
	if forwardToRunning(lockPath, "/games/demo.exe") {
		t.Fatal("forwarded to a launcher that is not running")
	}
}

func TestCallInstanceSendsTokenAndPath(t *testing.T) {
	var gotToken, gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.Header.Get("X-LL-Token")
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	info := instanceInfo{Port: ts.Listener.Addr().(*net.TCPAddr).Port, Token: "abc"}
	if err := callInstance(info, http.MethodGet, "/api/platform", nil); err != nil {
		t.Fatalf("call: %v", err)
	}
	if gotToken != "abc" || gotPath != "/api/platform" {
		t.Fatalf("token %q, path %q", gotToken, gotPath)
	}
}

func TestWriteInstanceRoundTrip(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "nested", "launcher.lock")
	if err := writeInstance(lockPath, instanceInfo{Port: 9, Token: "t"}); err != nil {
		t.Fatal(err)
	}
	info, err := readInstance(lockPath)
	if err != nil || info.Port != 9 || info.Token != "t" {
		t.Fatalf("read back %+v, %v", info, err)
	}
}
