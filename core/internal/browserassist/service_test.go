package browserassist

import (
	"context"
	"errors"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"light-launcher/core/internal/plugins"
)

type fakeProcess struct {
	mu        sync.Mutex
	output    string
	pos       int
	hold      bool
	waitErr   error
	closed    chan struct{}
	closeOnce sync.Once
	killed    atomic.Bool
}

func newFakeProcess(output string, hold bool, waitErr error) *fakeProcess {
	return &fakeProcess{output: output, hold: hold, waitErr: waitErr, closed: make(chan struct{})}
}

func (p *fakeProcess) Stdout() io.Reader { return p }

func (p *fakeProcess) Read(buf []byte) (int, error) {
	p.mu.Lock()
	if p.pos < len(p.output) {
		n := copy(buf, p.output[p.pos:])
		p.pos += n
		p.mu.Unlock()
		return n, nil
	}
	hold := p.hold
	p.mu.Unlock()
	if hold {
		<-p.closed
	}
	return 0, io.EOF
}

func (p *fakeProcess) Kill() error {
	p.killed.Store(true)
	p.closeOnce.Do(func() { close(p.closed) })
	return nil
}

func (p *fakeProcess) Wait() error { return p.waitErr }

type fakeStarter struct {
	mu        sync.Mutex
	proc      *fakeProcess
	startErr  error
	bin       string
	args      []string
	starts    int
	started   chan struct{}
	startOnce sync.Once
}

func newFakeStarter(proc *fakeProcess) *fakeStarter {
	return &fakeStarter{proc: proc, started: make(chan struct{})}
}

func (s *fakeStarter) Start(bin string, args []string) (Process, error) {
	s.mu.Lock()
	s.bin = bin
	s.args = slices.Clone(args)
	s.starts++
	proc := s.proc
	startErr := s.startErr
	s.mu.Unlock()
	s.startOnce.Do(func() { close(s.started) })
	if startErr != nil {
		return nil, startErr
	}
	return proc, nil
}

func (s *fakeStarter) call() (string, []string, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.bin, slices.Clone(s.args), s.starts
}

type recordingNotifier struct {
	mu      sync.Mutex
	waiting []string
}

func (n *recordingNotifier) Waiting(message string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.waiting = append(n.waiting, message)
}

func (n *recordingNotifier) Progress(int, string) {}

func (n *recordingNotifier) messages() []string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return slices.Clone(n.waiting)
}

func newTestService(t *testing.T, proc *fakeProcess) (*Service, *fakeStarter, string) {
	t.Helper()
	profiles := t.TempDir()
	starter := newFakeStarter(proc)
	service, err := New(Options{
		ProfilesDir: profiles,
		Shell:       func() (string, error) { return "/fake/light-launcher-shell", nil },
		Starter:     starter,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(service.Close)
	return service, starter, profiles
}

func testRequest(notify plugins.BrowserNotifier) plugins.BrowserAssistRequest {
	return plugins.BrowserAssistRequest{
		Plugin:  "f95",
		URL:     "https://example.com/thread/1",
		Message: "",
		Token:   "token",
		Timeout: time.Second,
		Notify:  notify,
	}
}

func TestAssistCapturesDownload(t *testing.T) {
	proc := newFakeProcess(`{"event":"ready"}
{"event":"download","url":"https://files.example/game.zip","filename":"game.zip","size":123,"headers":{"Cookie":"a=b","Referer":"https://example.com/","User-Agent":"test"}}
`, false, nil)
	service, starter, profiles := newTestService(t, proc)
	notify := &recordingNotifier{}

	capture, err := service.Assist(context.Background(), testRequest(notify))
	if err != nil {
		t.Fatalf("Assist: %v", err)
	}
	if capture.URL != "https://files.example/game.zip" || capture.Filename != "game.zip" || capture.Size != 123 {
		t.Fatalf("capture = %+v", capture)
	}
	wantHeaders := map[string]string{"Cookie": "a=b", "Referer": "https://example.com/", "User-Agent": "test"}
	if !maps.Equal(capture.Headers, wantHeaders) {
		t.Fatalf("headers = %v, want %v", capture.Headers, wantHeaders)
	}
	if got := notify.messages(); !slices.Equal(got, []string{defaultMessage}) {
		t.Fatalf("waiting messages = %v", got)
	}
	if !proc.killed.Load() {
		t.Fatal("shell process was not killed")
	}
	profileDir := filepath.Join(profiles, "f95")
	if info, err := os.Stat(profileDir); err != nil || !info.IsDir() {
		t.Fatalf("profile dir %s: %v", profileDir, err)
	}
	bin, args, starts := starter.call()
	if starts != 1 || bin != "/fake/light-launcher-shell" {
		t.Fatalf("starts = %d, bin = %q", starts, bin)
	}
	wantArgs := []string{
		"--assist", "https://example.com/thread/1",
		"--profile", profileDir,
		"--title", windowTitle,
		"--message", defaultMessage,
	}
	if !slices.Equal(args, wantArgs) {
		t.Fatalf("args = %v, want %v", args, wantArgs)
	}
}

func TestAssistPassesRequestMessage(t *testing.T) {
	proc := newFakeProcess(`{"event":"ready"}
{"event":"download","url":"https://files.example/game.zip","filename":"game.zip","size":-1}
`, false, nil)
	service, starter, _ := newTestService(t, proc)
	notify := &recordingNotifier{}
	request := testRequest(notify)
	request.Message = "Press the download button"

	capture, err := service.Assist(context.Background(), request)
	if err != nil {
		t.Fatalf("Assist: %v", err)
	}
	if capture.Size != -1 {
		t.Fatalf("size = %d, want -1", capture.Size)
	}
	if got := notify.messages(); !slices.Equal(got, []string{request.Message}) {
		t.Fatalf("waiting messages = %v", got)
	}
	_, args, _ := starter.call()
	if !slices.Contains(args, request.Message) {
		t.Fatalf("message missing from args %v", args)
	}
}

func TestAssistUnknownSizeIsMinusOne(t *testing.T) {
	proc := newFakeProcess(`{"event":"download","url":"https://files.example/game.zip","filename":"game.zip"}`+"\n", false, nil)
	service, _, _ := newTestService(t, proc)

	capture, err := service.Assist(context.Background(), testRequest(nil))
	if err != nil {
		t.Fatalf("Assist: %v", err)
	}
	if capture.Size != -1 {
		t.Fatalf("size = %d, want -1 when the shell omits it", capture.Size)
	}
}

func TestAssistClosedByUser(t *testing.T) {
	proc := newFakeProcess(`{"event":"closed"}`+"\n", false, nil)
	service, _, _ := newTestService(t, proc)

	_, err := service.Assist(context.Background(), testRequest(nil))
	if !errors.Is(err, ErrClosedByUser) {
		t.Fatalf("Assist error = %v, want ErrClosedByUser", err)
	}
	if !proc.killed.Load() {
		t.Fatal("shell process was not killed")
	}
}

func TestAssistShellErrorEvent(t *testing.T) {
	proc := newFakeProcess(`{"event":"error","message":"WebView2 runtime is missing"}`+"\n", false, nil)
	service, _, _ := newTestService(t, proc)

	_, err := service.Assist(context.Background(), testRequest(nil))
	var shellErr *ShellError
	if !errors.As(err, &shellErr) || shellErr.Message != "WebView2 runtime is missing" {
		t.Fatalf("Assist error = %v, want a ShellError", err)
	}
	if !proc.killed.Load() {
		t.Fatal("shell process was not killed")
	}
}

func TestAssistBadExit(t *testing.T) {
	proc := newFakeProcess("", false, errors.New("exit status 1"))
	service, _, _ := newTestService(t, proc)

	_, err := service.Assist(context.Background(), testRequest(nil))
	var shellErr *ShellError
	if !errors.As(err, &shellErr) || !strings.Contains(shellErr.Message, "exit status 1") {
		t.Fatalf("Assist error = %v, want a ShellError with the exit", err)
	}
	if !proc.killed.Load() {
		t.Fatal("shell process was not killed")
	}
}

func TestAssistTimeoutKillsTheShell(t *testing.T) {
	proc := newFakeProcess(`{"event":"ready"}
`, true, nil)
	service, _, _ := newTestService(t, proc)
	request := testRequest(nil)
	request.Timeout = 80 * time.Millisecond

	_, err := service.Assist(context.Background(), request)
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("Assist error = %v, want ErrTimeout", err)
	}
	if !proc.killed.Load() {
		t.Fatal("shell process was not killed")
	}
}

type cancelNotifier struct {
	cancel context.CancelFunc
	ready  atomic.Bool
}

func (n *cancelNotifier) Waiting(string) {
	n.ready.Store(true)
	n.cancel()
}

func (n *cancelNotifier) Progress(int, string) {}

func TestAssistCancelKillsTheShell(t *testing.T) {
	proc := newFakeProcess(`{"event":"ready"}
`, true, nil)
	service, _, _ := newTestService(t, proc)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	notify := &cancelNotifier{cancel: cancel}

	_, err := service.Assist(ctx, testRequest(notify))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Assist error = %v, want context.Canceled", err)
	}
	if !notify.ready.Load() {
		t.Fatal("ready event was not delivered before the cancel")
	}
	if !proc.killed.Load() {
		t.Fatal("shell process was not killed")
	}
}

func TestAssistReturnsBusyForASecondWindow(t *testing.T) {
	proc := newFakeProcess(`{"event":"ready"}
`, true, nil)
	service, starter, _ := newTestService(t, proc)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := make(chan error, 1)
	go func() {
		_, err := service.Assist(ctx, testRequest(nil))
		result <- err
	}()
	select {
	case <-starter.started:
	case <-time.After(2 * time.Second):
		t.Fatal("first assist did not start")
	}

	_, err := service.Assist(context.Background(), testRequest(nil))
	if !errors.Is(err, ErrBusy) {
		t.Fatalf("second Assist error = %v, want ErrBusy", err)
	}
	if _, _, starts := starter.call(); starts != 1 {
		t.Fatalf("starter calls = %d, want 1", starts)
	}
	cancel()
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatalf("first Assist error = %v, want context.Canceled", err)
	}
}

func TestAssistRejectsBadPluginID(t *testing.T) {
	tests := []string{"", ".", "..", "../escape", `a\b`, "a/b", "f95:alt", "f95 "}
	for _, pluginID := range tests {
		t.Run(pluginID, func(t *testing.T) {
			service, starter, _ := newTestService(t, newFakeProcess("", false, nil))
			request := testRequest(nil)
			request.Plugin = pluginID
			if _, err := service.Assist(context.Background(), request); !errors.Is(err, plugins.ErrInvalidID) {
				t.Fatalf("Assist(%q) error = %v, want ErrInvalidID", pluginID, err)
			}
			if _, _, starts := starter.call(); starts != 0 {
				t.Fatalf("starter was called %d times", starts)
			}
		})
	}
}

func TestAssistRejectsNonHTTPURL(t *testing.T) {
	tests := []string{"", "file:///tmp/game.zip", "ftp://example.com/game.zip", "https:///no-host", "javascript:alert(1)"}
	for _, pageURL := range tests {
		t.Run(pageURL, func(t *testing.T) {
			service, starter, _ := newTestService(t, newFakeProcess("", false, nil))
			request := testRequest(nil)
			request.URL = pageURL
			if _, err := service.Assist(context.Background(), request); !errors.Is(err, ErrInvalidURL) {
				t.Fatalf("Assist(%q) error = %v, want ErrInvalidURL", pageURL, err)
			}
			if _, _, starts := starter.call(); starts != 0 {
				t.Fatalf("starter was called %d times", starts)
			}
		})
	}
}

func TestAssistRejectsANonHTTPDownload(t *testing.T) {
	script := `{"event":"download","url":"blob:https://example.com/abc","filename":"a.zip","size":1}` + "\n"
	proc := newFakeProcess(script, false, nil)
	service, _, _ := newTestService(t, proc)

	_, err := service.Assist(context.Background(), testRequest(nil))
	var shellErr *ShellError
	if !errors.As(err, &shellErr) {
		t.Fatalf("Assist error = %v, want a ShellError", err)
	}
	if !proc.killed.Load() {
		t.Fatal("shell process was not killed")
	}
}

func TestAssistMissingShell(t *testing.T) {
	starter := newFakeStarter(newFakeProcess("", false, nil))
	service, err := New(Options{
		ProfilesDir: t.TempDir(),
		Shell:       func() (string, error) { return "", errors.New("not next to the launcher") },
		Starter:     starter,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(service.Close)

	_, err = service.Assist(context.Background(), testRequest(nil))
	if !errors.Is(err, ErrNoShell) {
		t.Fatalf("Assist error = %v, want ErrNoShell", err)
	}
	if _, _, starts := starter.call(); starts != 0 {
		t.Fatalf("starter was called %d times", starts)
	}
}

func TestAssistDisabled(t *testing.T) {
	starter := newFakeStarter(newFakeProcess("", false, nil))
	service, err := New(Options{
		ProfilesDir: t.TempDir(),
		Shell:       func() (string, error) { return "/fake/shell", nil },
		Starter:     starter,
		Enabled:     func() bool { return false },
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(service.Close)

	_, err = service.Assist(context.Background(), testRequest(nil))
	if !errors.Is(err, ErrDisabled) {
		t.Fatalf("Assist error = %v, want ErrDisabled", err)
	}
	if _, _, starts := starter.call(); starts != 0 {
		t.Fatalf("starter was called %d times", starts)
	}
}

func TestAssistStartError(t *testing.T) {
	starter := newFakeStarter(nil)
	starter.startErr = errors.New("cannot run")
	service, err := New(Options{
		ProfilesDir: t.TempDir(),
		Shell:       func() (string, error) { return "/fake/shell", nil },
		Starter:     starter,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(service.Close)

	_, err = service.Assist(context.Background(), testRequest(nil))
	if err == nil || !strings.Contains(err.Error(), "cannot run") {
		t.Fatalf("Assist error = %v, want the start error", err)
	}
}

func TestAssistIgnoresGarbageAndUnknownEvents(t *testing.T) {
	script := "not json\n" +
		"\n" +
		`{"event":"mystery"}` + "\n" +
		`{"event":` + "\n" +
		strings.Repeat("x", maxLineBytes+1) + "\n" +
		`{"event":"download","url":"https://files.example/game.zip","filename":"game.zip","size":7}` + "\n"
	proc := newFakeProcess(script, false, nil)
	service, _, _ := newTestService(t, proc)

	capture, err := service.Assist(context.Background(), testRequest(nil))
	if err != nil {
		t.Fatalf("Assist: %v", err)
	}
	if capture.Filename != "game.zip" || capture.Size != 7 {
		t.Fatalf("capture = %+v", capture)
	}
	if !proc.killed.Load() {
		t.Fatal("shell process was not killed")
	}
}

func TestAssistDropsAPartialLastLine(t *testing.T) {
	proc := newFakeProcess(`{"event":"ready"}`, false, nil)
	service, _, _ := newTestService(t, proc)
	notify := &recordingNotifier{}

	_, err := service.Assist(context.Background(), testRequest(notify))
	var shellErr *ShellError
	if !errors.As(err, &shellErr) {
		t.Fatalf("Assist error = %v, want a ShellError", err)
	}
	if got := notify.messages(); len(got) != 0 {
		t.Fatalf("partial line was handled: %v", got)
	}
	if !proc.killed.Load() {
		t.Fatal("shell process was not killed")
	}
}

func TestAssistCapsTheWholeOutput(t *testing.T) {
	proc := newFakeProcess(strings.Repeat("x", maxStreamBytes+1), false, nil)
	service, _, _ := newTestService(t, proc)

	_, err := service.Assist(context.Background(), testRequest(nil))
	var shellErr *ShellError
	if !errors.As(err, &shellErr) || !strings.Contains(shellErr.Message, "too large") {
		t.Fatalf("Assist error = %v, want a size error", err)
	}
	if !proc.killed.Load() {
		t.Fatal("shell process was not killed")
	}
}

func TestCloseStopsARunningAssist(t *testing.T) {
	proc := newFakeProcess(`{"event":"ready"}
`, true, nil)
	service, starter, _ := newTestService(t, proc)

	result := make(chan error, 1)
	go func() {
		_, err := service.Assist(context.Background(), testRequest(nil))
		result <- err
	}()
	select {
	case <-starter.started:
	case <-time.After(2 * time.Second):
		t.Fatal("assist did not start")
	}

	service.Close()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Assist error = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Assist did not return after Close")
	}
	if !proc.killed.Load() {
		t.Fatal("shell process was not killed")
	}
}

func TestAssistAfterClose(t *testing.T) {
	service, _, _ := newTestService(t, newFakeProcess("", false, nil))
	service.Close()
	if _, err := service.Assist(context.Background(), testRequest(nil)); !errors.Is(err, ErrShuttingDown) {
		t.Fatalf("Assist error = %v, want ErrShuttingDown", err)
	}
}
