package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

var fixtureBinary string

// TestMain builds the fixture plugin once for all tests in this package.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "plugins-fixture-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	name := "fixtureplugin"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	fixtureBinary = filepath.Join(dir, name)
	build := exec.Command("go", "build", "-o", fixtureBinary, "./testdata/fixtureplugin")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build fixture plugin: %v\n%s", err, out)
		_ = os.RemoveAll(dir)
		os.Exit(1)
	}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// disableRecorder remembers the reasons a supervisor reported it was disabled.
type disableRecorder struct {
	mu      sync.Mutex
	reasons []string
}

func (d *disableRecorder) record(reason string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.reasons = append(d.reasons, reason)
}

func (d *disableRecorder) count() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.reasons)
}

// copyExecutable copies src to dst and makes dst executable.
func copyExecutable(t *testing.T, src, dst string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, data, 0o755); err != nil {
		t.Fatal(err)
	}
}

// newFixtureSupervisor starts a supervisor for a copy of the fixture plugin.
func newFixtureSupervisor(t *testing.T, perms Permissions) (*Supervisor, *pluginEnv, *disableRecorder) {
	t.Helper()
	dir := t.TempDir()
	manifest := fixtureManifest(perms)
	copyExecutable(t, fixtureBinary, filepath.Join(dir, filepath.FromSlash(manifest.Entry[PlatformKey()])))

	deps := &serviceDeps{
		secrets:   newFileSecrets(t.TempDir()),
		transport: http.DefaultTransport.(*http.Transport).Clone(),
		limiter:   newHostLimiter(0),
		cache:     newETagCache(),
	}
	env := newPluginEnv(manifest, t.TempDir(), deps, newLineRing(100))
	recorder := &disableRecorder{}
	sup := newSupervisor(supervisorConfig{
		manifest:    manifest,
		dir:         dir,
		dataDir:     env.dataDir,
		hostVersion: "2.0.0",
		env:         env,
		onDisable:   recorder.record,
		backoff:     10 * time.Millisecond,
		crashLimit:  3,
	})
	t.Cleanup(sup.Shutdown)
	return sup, env, recorder
}

func fixtureManifest(perms Permissions) Manifest {
	entry := "bin/" + PlatformKey() + "/fixture"
	if runtime.GOOS == "windows" {
		entry += ".exe"
	}
	return Manifest{
		ID:          "fixture",
		Name:        "Fixture",
		Version:     "0.1.0",
		APIVersion:  SupportedAPIVersion,
		MinHost:     "2.0.0",
		Entry:       map[string]string{PlatformKey(): entry},
		Contributes: Contributes{Sources: []SourceContribution{{ID: "fixture", Label: "Fixture", Capabilities: []string{"browse"}}}},
		Permissions: perms,
	}
}

func callFixture(sup *Supervisor, method string, params any) (json.RawMessage, error) {
	var out json.RawMessage
	err := sup.Call(context.Background(), method, params, &out)
	return out, err
}

func TestHandshakeAndSuccessfulCall(t *testing.T) {
	sup, _, _ := newFixtureSupervisor(t, Permissions{})
	out, err := callFixture(sup, "source.parseRef", map[string]string{"url": "https://example.test/g/42"})
	if err != nil {
		t.Fatalf("source.parseRef: %v", err)
	}
	var ref map[string]string
	if err := json.Unmarshal(out, &ref); err != nil || ref["id"] != "42" {
		t.Fatalf("ref = %s, %v", out, err)
	}
	if !sup.Running() {
		t.Fatalf("plugin should be running after a call")
	}
}

func TestPluginNotificationsReachTheLog(t *testing.T) {
	sup, env, _ := newFixtureSupervisor(t, Permissions{})
	if _, err := callFixture(sup, "source.echo", map[string]string{"x": "1"}); err != nil {
		t.Fatalf("source.echo: %v", err)
	}
	lines := env.lines.tail(0)
	if !containsLine(lines, "[info] echo called") {
		t.Fatalf("log lines = %v", lines)
	}
	if !containsLine(lines, "fixture plugin started") {
		t.Fatalf("stderr was not captured: %v", lines)
	}
}

func TestHostServiceDenialAndAllowance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	defer server.Close()
	sup, _, _ := newFixtureSupervisor(t, Permissions{Network: []string{"127.0.0.1"}})

	denied := probeService(t, sup, "http.request", map[string]string{"url": "https://denied.example/page"})
	if denied.Error == nil || denied.Error.Code != CodePermissionDenied {
		t.Fatalf("denied host answer = %+v, want code %d", denied, CodePermissionDenied)
	}

	allowed := probeService(t, sup, "http.request", map[string]string{"url": server.URL})
	if allowed.Error != nil {
		t.Fatalf("allowed host failed: %+v", allowed.Error)
	}
}

type probeAnswer struct {
	Result json.RawMessage `json:"result"`
	Error  *RPCError       `json:"error"`
}

func probeService(t *testing.T, sup *Supervisor, service string, args any) probeAnswer {
	t.Helper()
	raw, err := callFixture(sup, "source.probe", map[string]any{"service": service, "args": args})
	if err != nil {
		t.Fatalf("source.probe: %v", err)
	}
	var answer probeAnswer
	if err := json.Unmarshal(raw, &answer); err != nil {
		t.Fatalf("probe answer %s: %v", raw, err)
	}
	return answer
}

func TestCrashRestartsWithBackoffThenAutoDisables(t *testing.T) {
	sup, _, recorder := newFixtureSupervisor(t, Permissions{})
	for crash := 1; crash <= 3; crash++ {
		if _, err := callFixture(sup, "source.parseRef", map[string]string{}); err != nil {
			t.Fatalf("crash %d: call before crash failed: %v", crash, err)
		}
		if _, err := callFixture(sup, "source.crash", nil); err == nil {
			t.Fatalf("crash %d: the crashing call should fail", crash)
		}
		waitUntil(t, func() bool { return sup.crashCount() == crash })
	}
	waitUntil(t, func() bool { return recorder.count() == 1 })

	if _, err := callFixture(sup, "source.parseRef", map[string]string{}); !errors.Is(err, ErrPluginDisabled) {
		t.Fatalf("after the crash limit err = %v, want ErrPluginDisabled", err)
	}
	if !sup.Disabled() {
		t.Fatalf("supervisor should report disabled")
	}
}

func TestShutdownStopsPluginAndNextCallRestartsIt(t *testing.T) {
	sup, _, _ := newFixtureSupervisor(t, Permissions{})
	if _, err := callFixture(sup, "source.parseRef", map[string]string{}); err != nil {
		t.Fatal(err)
	}
	sup.Shutdown()
	if sup.Running() {
		t.Fatalf("plugin still running after Shutdown")
	}
	if _, err := callFixture(sup, "source.parseRef", map[string]string{}); err != nil {
		t.Fatalf("call after Shutdown: %v", err)
	}
	if sup.crashCount() != 0 {
		t.Fatalf("a requested shutdown must not count as a crash")
	}
}

func (s *Supervisor) crashCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.crashes)
}

func containsLine(lines []string, want string) bool {
	for _, line := range lines {
		if strings.Contains(line, want) {
			return true
		}
	}
	return false
}

// waitUntil polls cond for up to two seconds and fails the test if it never holds.
func waitUntil(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("condition not met within 2s")
}
