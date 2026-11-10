//go:build e2e && !windows

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// urlLine matches the line the launcher logs when it serves the UI. Group 1 is the port, group 2 the token.
var urlLine = regexp.MustCompile(`open http://127\.0\.0\.1:(\d+)/#t=([0-9a-f]+)`)

// readMethods are the calls the UI makes on start. Each must succeed with an empty payload.
var readMethods = []string{
	"GetPlatform", "GetAppSettings", "GetAllGames", "LibraryList",
	"PluginsList", "DownloadsList", "GetSystemInfo",
}

// e2eLauncher is one running launcher process under test.
type e2eLauncher struct {
	cmd    *exec.Cmd
	log    *bytes.Buffer
	exited chan error
	base   string
	token  string
}

func TestLauncherServesAPIEndToEnd(t *testing.T) {
	binary := buildLauncher(t)
	root := t.TempDir()
	ui := filepath.Join(root, "ui")
	if err := os.MkdirAll(ui, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ui, "index.html"), []byte("<!doctype html><title>ui</title>"), 0o644); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(root, "config")

	launcher := startLauncher(t, binary, ui, config, filepath.Join(root, "home"))
	for _, method := range readMethods {
		reply := invoke(t, launcher.base, launcher.token, method)
		if !reply.Success {
			t.Errorf("%s failed: %s", method, reply.Error)
		}
	}
	res := postInvoke(t, launcher.base, "", "GetPlatform")
	res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("request without a token got %d, want 401", res.StatusCode)
	}

	stopLauncher(t, launcher)
	if _, err := os.Stat(filepath.Join(config, "light-launcher", "launcher.lock")); !os.IsNotExist(err) {
		t.Errorf("lock file left behind after exit: %v", err)
	}
}

// buildLauncher compiles this package into a temp folder.
func buildLauncher(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "light-launcher")
	build := exec.Command("go", "build", "-o", binary, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build launcher: %v\n%s", err, out)
	}
	return binary
}

// startLauncher runs the launcher without a window and waits for it to log its URL.
func startLauncher(t *testing.T, binary, ui, config, home string) *e2eLauncher {
	t.Helper()
	cmd := exec.Command(binary, "--no-shell", "--ui", ui)
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+config, "HOME="+home)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	launcher := &e2eLauncher{cmd: cmd, log: &bytes.Buffer{}, exited: make(chan error, 1)}
	t.Cleanup(func() { _ = cmd.Process.Kill() })

	found := make(chan [2]string, 1)
	logDone := make(chan struct{})
	go func() {
		readLog(stderr, launcher.log, found)
		close(logDone)
	}()
	// Wait reads the pipe's end, so it runs only after the log is fully read.
	go func() {
		<-logDone
		launcher.exited <- cmd.Wait()
	}()

	select {
	case match := <-found:
		launcher.base = "http://127.0.0.1:" + match[0]
		launcher.token = match[1]
	case <-time.After(20 * time.Second):
		t.Fatalf("launcher did not log its URL\n%s", launcher.log.String())
	}
	return launcher
}

// readLog copies the launcher log into buf and sends the port and token from the URL line, once.
func readLog(src io.Reader, buf *bytes.Buffer, found chan<- [2]string) {
	scanner := bufio.NewScanner(src)
	sent := false
	for scanner.Scan() {
		line := scanner.Text()
		buf.WriteString(line + "\n")
		if match := urlLine.FindStringSubmatch(line); match != nil && !sent {
			found <- [2]string{match[1], match[2]}
			sent = true
		}
	}
}

// stopLauncher sends SIGINT and checks that the launcher exits with status 0.
func stopLauncher(t *testing.T, launcher *e2eLauncher) {
	t.Helper()
	if err := launcher.cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("send interrupt: %v", err)
	}
	select {
	case err := <-launcher.exited:
		if err != nil {
			t.Fatalf("launcher exited with error after interrupt: %v\n%s", err, launcher.log.String())
		}
	case <-time.After(15 * time.Second):
		_ = launcher.cmd.Process.Kill()
		t.Fatalf("launcher did not stop after interrupt\n%s", launcher.log.String())
	}
}

// invokeReply is the envelope the invoke endpoint returns.
type invokeReply struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`
}

// invoke calls one method with an empty payload and decodes the reply.
func invoke(t *testing.T, base, token, method string) invokeReply {
	t.Helper()
	res := postInvoke(t, base, token, method)
	defer res.Body.Close()
	var reply invokeReply
	if err := json.NewDecoder(res.Body).Decode(&reply); err != nil {
		t.Fatalf("%s: bad reply: %v", method, err)
	}
	return reply
}

// postInvoke sends one call to the invoke endpoint. The caller closes the body.
func postInvoke(t *testing.T, base, token, method string) *http.Response {
	t.Helper()
	payload := fmt.Sprintf(`{"method":%q,"payload":null}`, method)
	req, err := http.NewRequest(http.MethodPost, base+"/api/invoke", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("X-LL-Token", token)
	}
	res, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("%s: %v", method, err)
	}
	return res
}
