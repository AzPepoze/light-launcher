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
)

const (
	exitOK      = 0
	exitFailure = 1
	exitUsage   = 2
)

var validShells = map[string]bool{"auto": true, "qt": true, "webview2": true, "browser": true}

// ignoredFlags are accepted so old launch scripts keep working. The launcher does nothing with them.
var ignoredFlags = map[string]bool{"wayland": true, "x11": true, "ozone-platform": true}

type options struct {
	shell    string
	noShell  bool
	uiDir    string
	editLsfg bool
	gamePath string
	ignored  []string
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "plugin" {
		os.Exit(runPluginCommand(os.Args[2:]))
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
	for _, name := range opts.ignored {
		logger.Info("main", "--%s is accepted and ignored", name)
	}
	if opts.editLsfg {
		logger.Warn("main", "--edit-lsfg is not wired yet and is ignored")
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	err = run(ctx, runConfig{
		UI:       ui,
		Shell:    resolveShell(opts.shell),
		NoShell:  opts.noShell,
		GamePath: opts.gamePath,
		LockPath: lockPath,
	})
	if err != nil {
		logger.Error("main", "%v", err)
		return exitFailure
	}
	return exitOK
}

func parseOptions(args []string) (options, error) {
	var opts options
	flags := flag.NewFlagSet("light-launcher", flag.ContinueOnError)
	flags.StringVar(&opts.shell, "shell", "auto", "window type: auto, qt, webview2 or browser")
	flags.BoolVar(&opts.noShell, "no-shell", false, "serve the UI and print the URL, without a window")
	flags.StringVar(&opts.uiDir, "ui", "", "directory with the built UI (default: embedded UI, else ui next to the executable)")
	flags.BoolVar(&opts.editLsfg, "edit-lsfg", false, "open the lsfg-vk editor (not wired yet)")
	flags.Bool("wayland", false, "accepted and ignored")
	flags.Bool("x11", false, "accepted and ignored")
	flags.String("ozone-platform", "", "accepted and ignored")
	if err := flags.Parse(args); err != nil {
		return opts, err
	}
	if !validShells[opts.shell] {
		return opts, fmt.Errorf("unknown shell %q", opts.shell)
	}
	flags.Visit(func(f *flag.Flag) {
		if ignoredFlags[f.Name] {
			opts.ignored = append(opts.ignored, f.Name)
		}
	})
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

func runPluginCommand(args []string) int {
	if PluginCLI == nil {
		logger.Warn("plugin", "plugin commands are not wired yet")
		return exitUsage
	}
	return PluginCLI(args, os.Stdout)
}
