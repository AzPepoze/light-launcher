// Package browserassist runs the embedded browser window for a plugin's manual browser step.
// The user logs in, waits and presses the download button by hand. The shell only watches for the
// download the page starts; nothing is clicked, typed, solved or hidden.
package browserassist

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/plugins"
	"light-launcher/core/internal/shellbin"
)

const (
	defaultMessage = "Finish the page; the download is captured automatically"
	windowTitle    = "LightLauncher"

	defaultTimeout = 2 * time.Minute
	maxTimeout     = 10 * time.Minute
)

var (
	// ErrNoShell means the shell binary is not next to the launcher.
	ErrNoShell = errors.New("browser shell is not installed")
	// ErrClosedByUser means the user closed the window without a download.
	ErrClosedByUser = errors.New("browser window was closed without a download")
	// ErrTimeout means the window did not produce a download in time.
	ErrTimeout = errors.New("browser assist timed out")
	// ErrBusy means another browser window is already open for a plugin.
	ErrBusy = errors.New("another browser assist is already running")
	// ErrDisabled means the browserAssistEnabled host setting is off.
	ErrDisabled = errors.New("browser assist is turned off")
	// ErrShuttingDown means the launcher is stopping and will not open a window.
	ErrShuttingDown = errors.New("browser assist is shutting down")
	// ErrInvalidURL means the page is not an http or https address.
	ErrInvalidURL = errors.New("browser assist needs an http or https address")

	errEnded = errors.New("browser shell output ended")
)

// ShellError is a failure the shell reported, or a shell that stopped without a result.
type ShellError struct {
	Message string
}

func (e *ShellError) Error() string {
	if e.Message == "" {
		return "browser shell failed"
	}
	return "browser shell: " + e.Message
}

// Options wires the service to the launcher. Only ProfilesDir is required.
type Options struct {
	ProfilesDir string
	Shell       func() (string, error) // default: shellbin.Path
	Starter     Starter                // default: start the shell binary
	Enabled     func() bool            // default: always on
}

// Service starts one embedded browser window at a time and returns what it captures.
type Service struct {
	profilesDir string
	shell       func() (string, error)
	starter     Starter
	enabled     func() bool

	mu     sync.Mutex
	busy   bool
	closed bool
	cancel context.CancelFunc
	done   chan struct{}
}

var _ plugins.BrowserAssist = (*Service)(nil)

// New prepares the service. It starts no process.
func New(opts Options) (*Service, error) {
	if opts.ProfilesDir == "" {
		return nil, errors.New("browser assist needs a profile folder")
	}
	service := &Service{
		profilesDir: opts.ProfilesDir,
		shell:       opts.Shell,
		starter:     opts.Starter,
		enabled:     opts.Enabled,
	}
	if service.shell == nil {
		service.shell = shellbin.Path
	}
	if service.starter == nil {
		service.starter = execStarter{}
	}
	return service, nil
}

// Assist opens the page in the shell window and waits for its download event.
func (s *Service) Assist(ctx context.Context, request plugins.BrowserAssistRequest) (plugins.BrowserCapture, error) {
	if !s.isEnabled() {
		return plugins.BrowserCapture{}, ErrDisabled
	}
	session, err := s.begin(ctx)
	if err != nil {
		return plugins.BrowserCapture{}, err
	}
	defer s.end(session)
	return s.run(session.ctx, request)
}

// Close stops the running window, if any, and waits for its process to be gone.
func (s *Service) Close() {
	s.mu.Lock()
	s.closed = true
	cancel := s.cancel
	done := s.done
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
}

func (s *Service) isEnabled() bool {
	return s.enabled == nil || s.enabled()
}

type session struct {
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

func (s *Service) begin(ctx context.Context) (*session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, ErrShuttingDown
	}
	if s.busy {
		return nil, ErrBusy
	}
	sessionCtx, cancel := context.WithCancel(ctx)
	s.busy = true
	s.cancel = cancel
	s.done = make(chan struct{})
	return &session{ctx: sessionCtx, cancel: cancel, done: s.done}, nil
}

func (s *Service) end(session *session) {
	session.cancel()
	s.mu.Lock()
	s.busy = false
	s.cancel = nil
	s.done = nil
	s.mu.Unlock()
	close(session.done)
}

func (s *Service) run(ctx context.Context, request plugins.BrowserAssistRequest) (plugins.BrowserCapture, error) {
	if err := ctx.Err(); err != nil {
		return plugins.BrowserCapture{}, err
	}
	if _, err := parseAssistURL(request.URL); err != nil {
		return plugins.BrowserCapture{}, err
	}
	profileDir, err := s.profileDir(request.Plugin)
	if err != nil {
		return plugins.BrowserCapture{}, err
	}
	shell, err := s.shell()
	if err != nil {
		return plugins.BrowserCapture{}, fmt.Errorf("%w: %w", ErrNoShell, err)
	}
	if err := os.MkdirAll(profileDir, 0o700); err != nil {
		return plugins.BrowserCapture{}, fmt.Errorf("create browser profile: %w", err)
	}

	message := request.Message
	if strings.TrimSpace(message) == "" {
		message = defaultMessage
	}
	timer := time.NewTimer(assistTimeout(request.Timeout))
	defer timer.Stop()

	proc, err := s.starter.Start(shell, assistArgs(request.URL, profileDir, message))
	if err != nil {
		return plugins.BrowserCapture{}, fmt.Errorf("start browser shell: %w", err)
	}

	var timedOut atomic.Bool
	watch := watchProcess(ctx, proc, timer.C, &timedOut)
	capture, found, result := collect(proc.Stdout(), request.Notify, message)
	watch.stop()
	if err := proc.Kill(); err != nil {
		logger.Debug("browser", "stop browser shell: %v", err)
	}
	waitErr := proc.Wait()

	if found {
		return capture, nil
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return plugins.BrowserCapture{}, ctxErr
	}
	if timedOut.Load() {
		return plugins.BrowserCapture{}, ErrTimeout
	}
	if errors.Is(result, errEnded) {
		if waitErr != nil {
			return plugins.BrowserCapture{}, &ShellError{Message: fmt.Sprintf("the window stopped: %v", waitErr)}
		}
		return plugins.BrowserCapture{}, &ShellError{Message: "the window stopped without a result"}
	}
	return plugins.BrowserCapture{}, result
}

func (s *Service) profileDir(pluginID string) (string, error) {
	if !validPluginID(pluginID) {
		return "", fmt.Errorf("browser assist: %w", plugins.ErrInvalidID)
	}
	return filepath.Join(s.profilesDir, pluginID), nil
}

// validPluginID accepts a plugin id that is one safe folder name.
func validPluginID(id string) bool {
	if id == "" || id == "." || id == ".." || len(id) > 64 {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
		default:
			return false
		}
	}
	return true
}

func parseAssistURL(raw string) (*url.URL, error) {
	target, err := url.Parse(raw)
	if err != nil || (target.Scheme != "http" && target.Scheme != "https") || target.Hostname() == "" {
		return nil, ErrInvalidURL
	}
	return target, nil
}

func assistArgs(pageURL, profileDir, message string) []string {
	return []string{
		"--assist", pageURL,
		"--profile", profileDir,
		"--title", windowTitle,
		"--message", message,
	}
}

func assistTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return defaultTimeout
	}
	return min(timeout, maxTimeout)
}

type shellEvent struct {
	Event    string            `json:"event"`
	URL      string            `json:"url"`
	Filename string            `json:"filename"`
	Size     *int64            `json:"size"`
	Headers  map[string]string `json:"headers"`
	Message  string            `json:"message"`
}

func collect(stdout io.Reader, notify plugins.BrowserNotifier, message string) (plugins.BrowserCapture, bool, error) {
	lines := newLineReader(stdout)
	for {
		raw, err := lines.next()
		if err != nil {
			switch {
			case errors.Is(err, io.EOF):
				return plugins.BrowserCapture{}, false, errEnded
			case errors.Is(err, errLineTooLong):
				logger.Debug("browser", "ignoring an oversized shell line")
				continue
			case errors.Is(err, errStreamLimit):
				return plugins.BrowserCapture{}, false, &ShellError{Message: "the shell output is too large"}
			default:
				return plugins.BrowserCapture{}, false, fmt.Errorf("read browser shell output: %w", err)
			}
		}
		line := bytes.TrimSpace(raw)
		if len(line) == 0 {
			continue
		}
		var event shellEvent
		if err := json.Unmarshal(line, &event); err != nil {
			logger.Debug("browser", "ignoring a malformed shell line: %v", err)
			continue
		}
		switch event.Event {
		case "ready":
			if notify != nil {
				notify.Waiting(message)
			}
		case "download":
			if _, err := parseAssistURL(event.URL); err != nil {
				return plugins.BrowserCapture{}, false, &ShellError{Message: "the shell reported a non-http download address"}
			}
			size := int64(-1)
			if event.Size != nil {
				size = *event.Size
			}
			return plugins.BrowserCapture{
				URL:      event.URL,
				Filename: event.Filename,
				Size:     size,
				Headers:  event.Headers,
			}, true, nil
		case "closed":
			return plugins.BrowserCapture{}, false, ErrClosedByUser
		case "error":
			detail := strings.TrimSpace(event.Message)
			if detail == "" {
				detail = "the shell reported an error"
			}
			return plugins.BrowserCapture{}, false, &ShellError{Message: detail}
		default:
			logger.Debug("browser", "ignoring unknown shell event %q", event.Event)
		}
	}
}

type processWatch struct {
	stopOnce sync.Once
	stopCh   chan struct{}
	done     chan struct{}
}

func watchProcess(ctx context.Context, proc Process, timeout <-chan time.Time, timedOut *atomic.Bool) *processWatch {
	watch := &processWatch{stopCh: make(chan struct{}), done: make(chan struct{})}
	go func() {
		defer close(watch.done)
		select {
		case <-ctx.Done():
			killAndLog(proc)
		case <-timeout:
			timedOut.Store(true)
			killAndLog(proc)
		case <-watch.stopCh:
		}
	}()
	return watch
}

func (w *processWatch) stop() {
	w.stopOnce.Do(func() { close(w.stopCh) })
	<-w.done
}

func killAndLog(proc Process) {
	if err := proc.Kill(); err != nil {
		logger.Debug("browser", "stop browser shell: %v", err)
	}
}
