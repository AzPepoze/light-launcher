package plugins

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/plugins/updater"
)

const logLinesKept = 500

// ErrInvalidID is returned for plugin ids that could not name a folder.
var ErrInvalidID = errors.New("invalid plugin id")

// HostOptions wires the plugin host to the launcher. Nil services are reported to plugins as unavailable.
type HostOptions struct {
	Layout      Layout
	HostVersion string
	Secrets     contract.Secrets // default: keyring or encrypted file
	Downloads   contract.Downloads
	Library     contract.Library
	Browser     BrowserAssist
	Folders     map[string]string // grant name such as "$GAMES" to its folder

	// UpdateChecker defaults to one that asks GitHub Releases. AllowPrereleases includes release candidates.
	UpdateChecker    UpdateChecker
	AllowPrereleases bool

	// IdleTimeout stops a plugin that no call touched for this long. Nil uses 5 minutes, zero disables it.
	IdleTimeout *time.Duration
}

// Host installs, runs and manages plugins.
type Host struct {
	layout      Layout
	hostVersion string
	install     *installer
	deps        *serviceDeps
	tuning      supervisorConfig
	updates     *updateTracker

	mu          sync.Mutex
	supervisors map[string]*Supervisor

	logsMu sync.Mutex
	logs   map[string]*lineRing
}

// NewHost prepares the plugin host. It does not start any plugin.
func NewHost(opts HostOptions) (*Host, error) {
	if opts.HostVersion == "" {
		return nil, errors.New("plugin host needs the launcher version")
	}
	in, err := newInstaller(opts.Layout, opts.HostVersion)
	if err != nil {
		return nil, fmt.Errorf("plugin state: %w", err)
	}
	secrets := opts.Secrets
	if secrets == nil {
		secrets = newPlatformSecrets(opts.Layout.Root)
	}
	deps := &serviceDeps{
		secrets:   trackedSecrets{inner: secrets, state: in.state},
		downloads: opts.Downloads,
		library:   opts.Library,
		browser:   opts.Browser,
		folders:   opts.Folders,
		transport: http.DefaultTransport.(*http.Transport).Clone(),
		limiter:   newHostLimiter(hostRequestInterval),
		cache:     newETagCache(),
	}
	checker := opts.UpdateChecker
	if checker == nil {
		checker = updater.NewChecker(updater.Options{
			AllowPrereleases: opts.AllowPrereleases,
			UserAgent:        "light-launcher/" + opts.HostVersion,
		})
	}
	return &Host{
		layout:      opts.Layout,
		hostVersion: opts.HostVersion,
		install:     in,
		deps:        deps,
		updates:     newUpdateTracker(checker),
		tuning:      supervisorConfig{idleTimeout: opts.IdleTimeout},
		supervisors: map[string]*Supervisor{},
		logs:        map[string]*lineRing{},
	}, nil
}

// SetEvents sends "plugin-state" and "plugin-progress" events to events. Call it before the host is used.
func (h *Host) SetEvents(events api.Events) {
	h.deps.events = events
}

// Shutdown stops every running plugin and drops a package kept by the last preview.
func (h *Host) Shutdown() {
	h.install.discardPreview()
	h.mu.Lock()
	running := make([]*Supervisor, 0, len(h.supervisors))
	for id, sup := range h.supervisors {
		running = append(running, sup)
		delete(h.supervisors, id)
	}
	h.mu.Unlock()
	for _, sup := range running {
		sup.Shutdown()
	}
}

// requireID rejects ids that could escape the plugin folders.
func requireID(id string) error {
	if !idPattern.MatchString(id) {
		return ErrInvalidID
	}
	return nil
}

func (h *Host) emitState(id string) {
	if h.deps.events == nil {
		return
	}
	h.deps.events.Emit("plugin-state", h.info(id))
}

// linesFor returns the log buffer of a plugin. It survives restarts.
func (h *Host) linesFor(id string) *lineRing {
	h.logsMu.Lock()
	defer h.logsMu.Unlock()
	lines, ok := h.logs[id]
	if !ok {
		lines = newLineRing(logLinesKept)
		h.logs[id] = lines
	}
	return lines
}
