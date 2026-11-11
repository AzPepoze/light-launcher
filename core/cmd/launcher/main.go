// Command light-launcher serves the launcher UI API and opens its window.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/plugins"
	"light-launcher/core/internal/services/appsvc"
)

const (
	exitOK      = 0
	exitFailure = 1
	exitUsage   = 2

	// defaultDevPort is the port the Vite dev proxy expects when --dev-token is given.
	defaultDevPort = 9246
)

// version is the launcher version. The release build sets it with -ldflags "-X main.version=...".
var version = "dev"

var validShells = map[string]bool{"auto": true, "qt": true, "webview2": true, "browser": true}

type options struct {
	shell    string
	noShell  bool
	uiDir    string
	editLsfg bool
	gamePath string
	port     int
	devToken string
}

func main() {
	plugins.HostVersion = hostVersion(version)
	if len(os.Args) > 1 && os.Args[1] == "plugin" {
		os.Exit(plugins.RunCLI(os.Args[1:], os.Stdout))
	}
	os.Exit(launch(os.Args[1:]))
}

// launch hands the game to a running launcher, or starts a new one. It returns the exit code.
func launch(args []string) int {
	opts, err := parseOptions(args)
	if errors.Is(err, flag.ErrHelp) {
		return exitOK
	}
	if err != nil {
		logger.Warn("main", "%v", err)
		return exitUsage
	}

	lockPath := lockFilePath()
	if forwardToRunning(lockPath, opts.gamePath) {
		return exitOK
	}
	ui, err := resolveUI(opts.uiDir)
	if err != nil {
		logger.Error("main", "%v", err)
		return exitFailure
	}

	// The game path and the lsfg editor flag are read by the UI on start, through the app service.
	appsvc.SetInitialArgs("", opts.gamePath, opts.editLsfg)
	port, token := apiEndpoint(opts)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	err = run(ctx, runConfig{
		UI:       ui,
		Shell:    resolveShell(opts.shell),
		NoShell:  opts.noShell,
		LockPath: lockPath,
		Port:     port,
		Token:    token,
	})
	if err != nil {
		logger.Error("main", "%v", err)
		return exitFailure
	}
	return exitOK
}

// apiEndpoint returns the port and token for the API. A fixed token is only used with --dev-token,
// and then the port defaults to the one the Vite dev proxy expects.
func apiEndpoint(opts options) (int, string) {
	if opts.devToken == "" {
		return opts.port, ""
	}
	if opts.port == 0 {
		return defaultDevPort, opts.devToken
	}
	return opts.port, opts.devToken
}

func parseOptions(args []string) (options, error) {
	var opts options
	flags := flag.NewFlagSet("light-launcher", flag.ContinueOnError)
	flags.StringVar(&opts.shell, "shell", "auto", "window type: auto, qt, webview2 or browser")
	flags.BoolVar(&opts.noShell, "no-shell", false, "serve the UI and print the URL, without a window")
	flags.StringVar(&opts.uiDir, "ui", "", "directory with the built UI (default: embedded UI, else ui next to the executable)")
	flags.BoolVar(&opts.editLsfg, "edit-lsfg", false, "open the lsfg-vk editor on start")
	flags.IntVar(&opts.port, "port", 0, "loopback port for the API (default: a free port)")
	flags.StringVar(&opts.devToken, "dev-token", "", "fixed API token for the Vite dev proxy; sets port 9246 unless --port is given")
	if err := flags.Parse(args); err != nil {
		return opts, err
	}
	if !validShells[opts.shell] {
		return opts, fmt.Errorf("unknown shell %q", opts.shell)
	}
	if opts.port < 0 || opts.port > 65535 {
		return opts, fmt.Errorf("port %d is out of range", opts.port)
	}
	if flags.NArg() > 0 {
		gamePath, err := filepath.Abs(flags.Arg(0))
		if err != nil {
			return opts, err
		}
		opts.gamePath = gamePath
	}
	return opts, nil
}

// resolveUI picks the UI source and checks that it has an index.html.
func resolveUI(dir string) (fs.FS, error) {
	ui, err := uiSource(dir)
	if err != nil {
		return nil, err
	}
	if _, err := fs.Stat(ui, "index.html"); err != nil {
		return nil, fmt.Errorf("ui not found: %w", err)
	}
	return ui, nil
}

// uiSource prefers the --ui folder, then the embedded UI, then ui next to the executable.
func uiSource(dir string) (fs.FS, error) {
	if dir != "" {
		return os.DirFS(dir), nil
	}
	if ui, ok := embeddedUI(); ok {
		return ui, nil
	}
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	return os.DirFS(filepath.Join(filepath.Dir(exe), "ui")), nil
}
