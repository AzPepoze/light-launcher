package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"
)

// TestRunWithoutShellServesUntilCancelled starts the real run loop on a random port and
// checks that a second launch with a game path reaches it.
func TestRunWithoutShellServesUntilCancelled(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "launcher.lock")
	ui := fstest.MapFS{"index.html": {Data: []byte("<!doctype html>ui")}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- run(ctx, runConfig{UI: ui, NoShell: true, LockPath: lockPath})
	}()

	info := waitForLock(t, lockPath)
	if err := callInstance(info, http.MethodGet, "/api/platform", nil); err != nil {
		t.Fatalf("platform call: %v", err)
	}
	if !forwardToRunning(lockPath, "/games/demo.exe") {
		t.Fatal("second launch with a game path was not forwarded")
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("run: %v", err)
	}
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatalf("lock file left behind: %v", err)
	}
}

// waitForLock polls until run has written the lock file, then returns its contents.
func waitForLock(t *testing.T, lockPath string) instanceInfo {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if info, err := readInstance(lockPath); err == nil {
			return info
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("lock file was not written")
	return instanceInfo{}
}
