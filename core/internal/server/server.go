// Package server is the HTTP side of the launcher: the UI API, the event stream and the built UI.
package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"sync"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/uploads"
)

// Config wires the server to the services. Port is the loopback port the listener uses.
type Config struct {
	UI      fs.FS
	Router  *api.Router
	Hub     *api.Hub
	Covers  CoverSource
	Uploads *uploads.Store
	Token   string
	Port    int
	// OnOpen receives a game path sent by a second launch.
	OnOpen func(gamePath string) error
}

// CoverSource finds the cached cover file of a game.
type CoverSource interface {
	CoverFile(id int64) (path string, ok bool)
}

type server struct {
	cfg       Config
	host      string
	origin    string
	csp       string
	readyOnce sync.Once
}

// New returns the handler for the launcher's HTTP server. It panics on a config that cannot work.
func New(cfg Config) http.Handler {
	if cfg.UI == nil || cfg.Router == nil || cfg.Hub == nil || cfg.Covers == nil || cfg.Uploads == nil || cfg.OnOpen == nil || cfg.Token == "" {
		panic("server: UI, Router, Hub, Covers, Uploads, OnOpen and Token are required")
	}
	host := fmt.Sprintf("127.0.0.1:%d", cfg.Port)
	s := &server{
		cfg:    cfg,
		host:   host,
		origin: "http://" + host,
		csp:    contentSecurityPolicy(cfg.Port),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/invoke", s.handleInvoke)
	mux.HandleFunc("GET /api/events", s.handleEvents)
	mux.HandleFunc("POST /api/ready", s.handleReady)
	mux.HandleFunc("GET /api/platform", s.handlePlatform)
	mux.HandleFunc("POST /api/open", s.handleOpen)
	mux.HandleFunc("GET /api/cover/{id}", s.handleCover)
	mux.HandleFunc("POST /api/plugin-upload", s.handleUpload)
	// Unknown API paths must not fall through to the UI's index.html.
	mux.Handle("/api/", http.NotFoundHandler())
	mux.Handle("/", serveUI(cfg.UI))
	return s.guard(mux)
}
