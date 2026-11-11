package downloads

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/logger"
)

const logCategory = "downloads"

// Installer turns a finished archive into an installed game.
type Installer interface {
	Install(ctx context.Context, job InstallJob) error
}

// InstallJob is one archive to install.
type InstallJob struct {
	DownloadID int64
	GameID     int64
	Archive    string
	Subfolder  string // folder under the games folder; empty means the folder itself
}

// Options configures NewManager.
type Options struct {
	Store     Store
	Events    api.Events   // optional
	Dir       string       // folder that finished files are saved to
	GamesDir  string       // optional; the games folder, needed for requests with Root "games"
	Installer Installer    // optional; without it, downloads that ask for an install are only saved
	Client    *http.Client // optional; a default client is used when nil
	Settings  Settings     // optional; the zero value means DefaultSettings
}

// Manager queues downloads, runs them within the limits in Settings, and reports progress.
type Manager struct {
	store     Store
	events    api.Events
	dir       string
	gamesDir  string
	installer Installer
	client    *http.Client
	limiter   *RateLimiter
	hosts     *hostGates

	ctx      context.Context
	cancel   context.CancelFunc
	jobs     sync.WaitGroup // running transfers
	installs sync.WaitGroup // running installers

	mu       sync.Mutex
	settings Settings
	active   map[int64]*job
	closed   bool
}

var _ contract.Downloads = (*Manager)(nil)

// NewManager builds a manager. Call Restore to pick up downloads left by an earlier run.
func NewManager(opts Options) (*Manager, error) {
	if opts.Store == nil {
		return nil, errors.New("downloads: a store is required")
	}
	if opts.Dir == "" {
		return nil, errors.New("downloads: a folder is required")
	}
	settings := DefaultSettings()
	if opts.Settings != (Settings{}) {
		settings = opts.Settings.normalized()
	}
	client := opts.Client
	if client == nil {
		client = defaultClient()
	}
	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{
		store:     opts.Store,
		events:    opts.Events,
		dir:       opts.Dir,
		gamesDir:  opts.GamesDir,
		installer: opts.Installer,
		client:    client,
		limiter:   &RateLimiter{},
		hosts:     newHostGates(settings.HostConnections),
		ctx:       ctx,
		cancel:    cancel,
		settings:  settings,
		active:    map[int64]*job{},
	}
	m.limiter.SetRate(bytesPerSecond(settings))
	return m, nil
}

func defaultClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = 32
	return &http.Client{Transport: transport}
}

// Enqueue adds a download and starts it when a slot is free. It implements contract.Downloads.
func (m *Manager) Enqueue(ctx context.Context, req contract.DownloadRequest) (int64, error) {
	if err := checkLink(req.URL); err != nil {
		return 0, err
	}
	headers := requestHeaders(req.URL, req.Headers)
	probe, err := Probe(ctx, m.client, req.URL, headers)
	var webPage *WebPageError
	if errors.As(err, &webPage) {
		return 0, err
	}
	if err != nil {
		logger.Warn(logCategory, "could not probe %s: %v", hostOf(req.URL), plainError(err))
		probe = ProbeResult{URL: req.URL}
	}
	headers = stripForeignCredentials(headers, req.URL, probe.URL)
	if err := ValidateRequest(req); err != nil {
		return 0, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, errors.New("downloads are shutting down")
	}
	dir, err := m.downloadFolder(req)
	if err != nil {
		return 0, err
	}
	dest, err := m.freeDestLocked(ctx, dir, downloadName(req.Filename, probe))
	if err != nil {
		return 0, err
	}
	id, err := m.store.Insert(ctx, Download{
		GameID:           req.GameID,
		URL:              probe.URL,
		Filename:         filepath.Base(dest),
		Dest:             dest,
		Size:             probe.Size,
		Connections:      connectionsFor(probe, m.settings.Connections),
		State:            StateQueued,
		Headers:          headers,
		Install:          req.Install,
		InstallSubfolder: req.InstallSubfolder,
		DeleteArchive:    req.DeleteArchive,
	})
	if err != nil {
		return 0, err
	}
	logger.Info(logCategory, "queued download %d from %s", id, hostOf(req.URL))
	m.pumpLocked()
	return id, nil
}

// Restore resumes what an earlier run left behind. Downloads that were running go back to the queue.
func (m *Manager) Restore(ctx context.Context) error {
	rows, err := m.store.List(ctx)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, d := range rows {
		if d.State != StateDownloading {
			continue
		}
		if fileExists(d.Dest + ".part") {
			err = m.updateLocked(ctx, d.ID, statePatch(StateQueued, ""))
		} else {
			err = m.updateLocked(ctx, d.ID, statePatch(StateError, "partial file is missing"))
		}
		if err != nil {
			return err
		}
	}
	m.pumpLocked()
	return nil
}

// Shutdown stops running transfers, keeps their progress, and waits for transfers and installs to end.
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	m.closed = true
	for _, j := range m.active {
		j.requestStop(stopShutdown)
	}
	m.mu.Unlock()
	if err := waitGroup(ctx, &m.jobs); err != nil {
		return err
	}
	m.cancel()
	return waitGroup(ctx, &m.installs)
}

// List returns every download, newest first.
func (m *Manager) List(ctx context.Context) ([]View, error) {
	rows, err := m.store.List(ctx)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	out := make([]View, 0, len(rows))
	for _, d := range rows {
		out = append(out, m.viewLocked(d, now))
	}
	return out, nil
}

// Get returns one download.
func (m *Manager) Get(ctx context.Context, id int64) (View, error) {
	d, err := m.store.Get(ctx, id)
	if err != nil {
		return View{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.viewLocked(d, time.Now()), nil
}
