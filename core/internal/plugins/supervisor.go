package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"light-launcher/core/internal/logger"
)

const (
	defaultCallTimeout = 60 * time.Second
	initializeTimeout  = 10 * time.Second
	shutdownTimeout    = 3 * time.Second
	exitGrace          = 3 * time.Second // the SDK waits up to 2 s for running requests before it answers shutdown
	defaultIdleTimeout = 5 * time.Minute
)

// ErrPluginDisabled is returned for calls to a plugin that was turned off, or auto-disabled after crashes.
var ErrPluginDisabled = errors.New("plugin is disabled")

// supervisorConfig is what one supervisor needs. Zero timing fields take the defaults from withDefaults.
type supervisorConfig struct {
	manifest    Manifest
	dir         string // installed version folder
	dataDir     string
	hostVersion string
	env         *pluginEnv
	onDisable   func(reason string)
	onIdle      func()
	heartbeat   time.Duration
	backoff     time.Duration
	crashWindow time.Duration
	crashLimit  int
	idleTimeout *time.Duration // 0 disables and nil takes defaultIdleTimeout
	now         func() time.Time
	until       func(deadline time.Time) <-chan time.Time // fires at an absolute time, so a late call cannot move the deadline
}

// Supervisor starts one plugin on first use, keeps it alive, and stops it on request.
type Supervisor struct {
	mu sync.Mutex
	// startMu serializes starts, including the initialize handshake. It is separate from mu
	// because the plugin may call host services during the handshake, and those calls record
	// activity on this supervisor; holding mu there would deadlock the handshake.
	startMu    sync.Mutex
	cfg        supervisorConfig
	conn       *connection
	crashes    []time.Time
	nextStart  time.Time
	disabled   bool
	idle       bool
	lastActive time.Time
	inFlight   int
	jobs       map[string]bool
	activityCh chan struct{}
}

// connection is one live plugin process and its JSON-RPC peer.
type connection struct {
	child        *childProcess
	peer         *Peer
	exited       chan struct{} // closed after the exit has been recorded
	stopping     bool          // set by Shutdown or the idle watcher, so the exit is not counted as a crash
	stopOnce     sync.Once
	stopComplete chan struct{} // closed after the shutdown RPC and process exit
}

type initializeParams struct {
	HostVersion        string      `json:"hostVersion"`
	APIVersion         string      `json:"apiVersion"`
	Platform           string      `json:"platform"`
	GrantedPermissions Permissions `json:"grantedPermissions"`
	DataDir            string      `json:"dataDir"`
}

type initializeResult struct {
	PluginVersion string                     `json:"pluginVersion"`
	Capabilities  map[string]json.RawMessage `json:"capabilities"`
}

func newSupervisor(cfg supervisorConfig) *Supervisor {
	sup := &Supervisor{cfg: cfg.withDefaults(), jobs: map[string]bool{}, activityCh: make(chan struct{}, 1)}
	sup.lastActive = sup.cfg.now()
	if sup.cfg.env != nil {
		sup.cfg.env.setActivityHooks(sup.beginActivity, sup.endActivity, sup.activity, sup.progress)
	}
	return sup
}

// Call runs one plugin method. It starts the plugin when needed and applies a default timeout.
func (s *Supervisor) Call(ctx context.Context, method string, params, out any) error {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultCallTimeout)
		defer cancel()
	}
	s.beginActivity()
	defer s.endActivity()
	conn, err := s.connect(ctx)
	if err != nil {
		return err
	}
	return conn.peer.Call(ctx, method, params, out)
}

// Running reports whether the plugin process is up.
func (s *Supervisor) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn != nil
}

// Start launches the plugin and completes the initialize handshake.
func (s *Supervisor) Start(ctx context.Context) error {
	_, err := s.connect(ctx)
	return err
}

// Disabled reports whether the supervisor refuses to start the plugin.
func (s *Supervisor) Disabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.disabled
}

// Idle reports whether the plugin was stopped because it was idle.
func (s *Supervisor) Idle() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.idle
}

// LastActive returns the time of the last call or notification.
func (s *Supervisor) LastActive() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastActive
}

// connect returns the live connection, starting the plugin when there is none.
func (s *Supervisor) connect(ctx context.Context) (*connection, error) {
	for {
		s.mu.Lock()
		current := s.conn
		stopping := current != nil && current.stopping
		s.mu.Unlock()
		if current == nil {
			break
		}
		if stopping {
			// A stop is in progress. Wait for it, then start a fresh process.
			select {
			case <-current.exited:
				continue
			case <-ctx.Done():
				return nil, contextError(ctx.Err())
			}
		}
		if current.alive() {
			return current, nil
		}
		// The exit must be recorded before a new start, so the backoff applies.
		<-current.exited
	}
	conn, reason, err := s.startSerial(ctx)
	if reason != "" {
		s.notifyDisabled(reason)
	}
	return conn, err
}

// startSerial starts the plugin when no live connection exists. It returns a disable reason
// when the start counts as a crash. Only the short state checks and updates hold mu; the
// backoff sleep and the handshake run under startMu alone, so host services the plugin calls
// during initialize can still reach this supervisor.
func (s *Supervisor) startSerial(ctx context.Context) (*connection, string, error) {
	s.startMu.Lock()
	defer s.startMu.Unlock()

	s.mu.Lock()
	if s.disabled {
		s.mu.Unlock()
		return nil, "", ErrPluginDisabled
	}
	if s.conn != nil && s.conn.alive() {
		conn := s.conn
		s.mu.Unlock()
		return conn, "", nil
	}
	nextStart := s.nextStart
	s.mu.Unlock()

	if err := sleepContext(ctx, time.Until(nextStart)); err != nil {
		return nil, "", contextError(err)
	}
	conn, err := s.start(ctx)
	if err != nil {
		return nil, s.recordCrash(time.Now()), err
	}

	s.mu.Lock()
	s.conn = conn
	s.idle = false
	s.lastActive = s.cfg.now()
	s.mu.Unlock()
	return conn, "", nil
}

// start spawns the program and runs the initialize handshake.
func (s *Supervisor) start(ctx context.Context) (*connection, error) {
	entry, _ := s.cfg.manifest.EntryFor(PlatformKey())
	path := filepath.Join(s.cfg.dir, filepath.FromSlash(entry))
	child, err := startChild(path, s.cfg.dir, func(line string) {
		s.cfg.env.log("info", line)
	})
	if err != nil {
		return nil, fmt.Errorf("start plugin %s: %w", s.cfg.manifest.ID, err)
	}
	conn := &connection{child: child, exited: make(chan struct{}), stopComplete: make(chan struct{})}
	conn.peer = NewPeer(child.stdout, child.stdin, s.cfg.env.call, s.cfg.env.notify)
	s.cfg.env.setPeer(conn.peer)
	go s.watch(conn)
	if err := s.handshake(ctx, conn); err != nil {
		conn.child.kill()
		return nil, err
	}
	logger.Info(s.cfg.env.category(), "started version %s", s.cfg.manifest.Version)
	go s.heartbeat(conn)
	go s.watchIdle(conn)
	return conn, nil
}

func (s *Supervisor) handshake(ctx context.Context, conn *connection) error {
	initCtx, cancel := context.WithTimeout(ctx, initializeTimeout)
	defer cancel()
	params := initializeParams{
		HostVersion:        s.cfg.hostVersion,
		APIVersion:         SupportedAPIVersion,
		Platform:           platformOS(runtime.GOOS),
		GrantedPermissions: grantedPermissions(s.cfg.manifest.Permissions),
		DataDir:            s.cfg.dataDir,
	}
	var result initializeResult
	if err := conn.peer.Call(initCtx, "initialize", params, &result); err != nil {
		return fmt.Errorf("plugin %s did not complete initialize: %w", s.cfg.manifest.ID, err)
	}
	return nil
}

// grantedPermissions returns the grant with empty lists instead of nil, as the schema expects arrays.
func grantedPermissions(p Permissions) Permissions {
	if p.Network == nil {
		p.Network = []string{}
	}
	if p.FS == nil {
		p.FS = []string{}
	}
	return p
}

// Shutdown asks the plugin to exit, then stops its process group after a short wait.
func (s *Supervisor) Shutdown() {
	// Wait for a start in progress, so a plugin that is still handshaking is stopped too.
	s.startMu.Lock()
	s.mu.Lock()
	conn := s.conn
	if conn != nil {
		conn.stopping = true
	}
	s.mu.Unlock()
	s.startMu.Unlock()
	if conn == nil {
		return
	}
	s.stopConn(conn)
}

// stopConn runs the graceful shutdown once and waits until the process is gone.
func (s *Supervisor) stopConn(conn *connection) {
	conn.stopOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = conn.peer.Call(ctx, "shutdown", nil, nil)
		_ = conn.child.stdin.Close()
		select {
		case <-conn.exited:
		case <-time.After(exitGrace):
			conn.child.kill()
			<-conn.exited
		}
		close(conn.stopComplete)
	})
	<-conn.stopComplete
}

func (c *connection) alive() bool {
	select {
	case <-c.peer.Done():
		return false
	default:
		return true
	}
}

// beginActivity marks the plugin as busy and resets its idle timer.
func (s *Supervisor) beginActivity() {
	s.mu.Lock()
	s.inFlight++
	s.idle = false
	s.touchLocked()
	s.mu.Unlock()
}

// endActivity releases one busy mark and resets the idle timer again.
func (s *Supervisor) endActivity() {
	s.mu.Lock()
	if s.inFlight > 0 {
		s.inFlight--
	}
	s.touchLocked()
	s.mu.Unlock()
}

// activity records a notification or other sign of life without changing the busy count.
func (s *Supervisor) activity() {
	s.mu.Lock()
	s.touchLocked()
	s.mu.Unlock()
}

func (s *Supervisor) touchLocked() {
	s.lastActive = s.cfg.now()
	select {
	case s.activityCh <- struct{}{}:
	default:
	}
}

// progress records a progress job. An unfinished job keeps the plugin from sleeping.
func (s *Supervisor) progress(id string, done, total int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.touchLocked()
	if id == "" {
		return
	}
	if total > 0 && done >= total {
		delete(s.jobs, id)
		return
	}
	s.jobs[id] = true
}

func (c supervisorConfig) withDefaults() supervisorConfig {
	if c.heartbeat <= 0 {
		c.heartbeat = defaultHeartbeat
	}
	if c.backoff <= 0 {
		c.backoff = defaultBackoff
	}
	if c.crashWindow <= 0 {
		c.crashWindow = defaultCrashWindow
	}
	if c.crashLimit <= 0 {
		c.crashLimit = defaultCrashLimit
	}
	if c.onDisable == nil {
		c.onDisable = func(string) {}
	}
	if c.onIdle == nil {
		c.onIdle = func() {}
	}
	if c.idleTimeout == nil {
		timeout := defaultIdleTimeout
		c.idleTimeout = &timeout
	}
	if c.now == nil {
		c.now = time.Now
	}
	if c.until == nil {
		c.until = func(deadline time.Time) <-chan time.Time { return time.After(time.Until(deadline)) }
	}
	return c
}
