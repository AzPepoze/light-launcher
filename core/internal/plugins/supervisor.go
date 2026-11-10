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
	heartbeat   time.Duration
	backoff     time.Duration
	crashWindow time.Duration
	crashLimit  int
}

// Supervisor starts one plugin on first use, keeps it alive, and stops it on request.
type Supervisor struct {
	mu        sync.Mutex
	cfg       supervisorConfig
	conn      *connection
	crashes   []time.Time
	nextStart time.Time
	disabled  bool
}

// connection is one live plugin process and its JSON-RPC peer.
type connection struct {
	child    *childProcess
	peer     *Peer
	exited   chan struct{} // closed after the exit has been recorded
	stopping bool          // set by Shutdown, so the exit is not counted as a crash
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
	return &Supervisor{cfg: cfg.withDefaults()}
}

// Call runs one plugin method. It starts the plugin when needed and applies a default timeout.
func (s *Supervisor) Call(ctx context.Context, method string, params, out any) error {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultCallTimeout)
		defer cancel()
	}
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

// Disabled reports whether the supervisor refuses to start the plugin.
func (s *Supervisor) Disabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.disabled
}

// connect returns the live connection, starting the plugin when there is none.
func (s *Supervisor) connect(ctx context.Context) (*connection, error) {
	s.mu.Lock()
	current := s.conn
	s.mu.Unlock()
	if current != nil && current.alive() {
		return current, nil
	}
	if current != nil {
		// The exit must be recorded before a new start, so the backoff applies.
		<-current.exited
	}
	conn, reason, err := s.startLocked(ctx)
	if reason != "" {
		s.notifyDisabled(reason)
	}
	return conn, err
}

// startLocked holds the lock while it starts the plugin. It returns a disable reason when the start counts as a crash.
func (s *Supervisor) startLocked(ctx context.Context) (*connection, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.disabled {
		return nil, "", ErrPluginDisabled
	}
	if s.conn != nil && s.conn.alive() {
		return s.conn, "", nil
	}
	if err := sleepContext(ctx, time.Until(s.nextStart)); err != nil {
		return nil, "", contextError(err)
	}
	conn, err := s.start(ctx)
	if err != nil {
		return nil, s.recordCrashLocked(time.Now()), err
	}
	s.conn = conn
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
	conn := &connection{child: child, exited: make(chan struct{})}
	conn.peer = NewPeer(child.stdout, child.stdin, s.cfg.env.call, s.cfg.env.notify)
	s.cfg.env.setPeer(conn.peer)
	go s.watch(conn)
	if err := s.handshake(ctx, conn); err != nil {
		conn.child.kill()
		return nil, err
	}
	logger.Info(s.cfg.env.category(), "started version %s", s.cfg.manifest.Version)
	go s.heartbeat(conn)
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
	s.mu.Lock()
	conn := s.conn
	s.conn = nil
	if conn != nil {
		conn.stopping = true
	}
	s.mu.Unlock()
	if conn == nil {
		return
	}
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
}

func (c *connection) alive() bool {
	select {
	case <-c.peer.Done():
		return false
	default:
		return true
	}
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
	return c
}
