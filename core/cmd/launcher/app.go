package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/server"
)

const (
	gameOpenChannel = "open-game"
	shutdownTimeout = 3 * time.Second
	headerTimeout   = 10 * time.Second
)

// runConfig is what one launcher run needs. Tests pass in an in-memory UI and no window.
// Port 0 picks a free port. An empty Token makes a random one.
type runConfig struct {
	UI       fs.FS
	Shell    string
	NoShell  bool
	LockPath string
	Port     int
	Token    string
}

// run serves the UI API on a loopback port, opens the window and blocks until the window closes or
// ctx ends.
func run(ctx context.Context, cfg runConfig) error {
	token := cfg.Token
	if token == "" {
		generated, err := newToken()
		if err != nil {
			return err
		}
		token = generated
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", cfg.Port))
	if err != nil {
		return err
	}
	port := listener.Addr().(*net.TCPAddr).Port

	hub := api.NewHub()
	router := api.NewRouter()
	svc, err := wireServices(ctx, router, hub)
	if err != nil {
		listener.Close()
		return err
	}
	defer svc.Close()

	handler := server.New(server.Config{
		UI:     cfg.UI,
		Router: router,
		Hub:    hub,
		Covers: svc.library,
		Token:  token,
		Port:   port,
		OnOpen: func(gamePath string) error {
			openGame(hub, gamePath)
			return nil
		},
	})
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: headerTimeout}
	go serve(srv, listener)
	defer shutdown(srv)

	if err := writeInstance(cfg.LockPath, instanceInfo{Port: port, Token: token}); err != nil {
		return err
	}
	defer removeInstance(cfg.LockPath, port)
	logger.Info("http", "serving the UI on 127.0.0.1:%d", port)

	// The token rides in the URL fragment, so it never reaches the server in a request line.
	// The shell gets the URL through LL_URL, not argv.
	url := fmt.Sprintf("http://127.0.0.1:%d/#t=%s", port, token)
	if cfg.NoShell {
		logger.Info("main", "open %s", url)
		<-ctx.Done()
		return nil
	}
	return openWindow(ctx, cfg.Shell, url)
}

// openGame tells the UI that a second launch sent a game. The first game is read with GetInitialGamePath.
func openGame(hub *api.Hub, gamePath string) {
	logger.Info("game", "open requested: %s", filepath.Base(gamePath))
	hub.Emit(gameOpenChannel, map[string]string{"path": gamePath})
}

func serve(srv *http.Server, listener net.Listener) {
	if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("http", "%v", err)
	}
}

// shutdown lets open requests finish, then closes what is left (for example an event stream).
func shutdown(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		_ = srv.Close()
	}
}

func newToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
