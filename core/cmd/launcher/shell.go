package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"light-launcher/core/internal/logger"
)

const (
	shellBinaryName  = "light-launcher-shell"
	shellExitTimeout = 3 * time.Second
)

// resolveShell turns "auto" into the window type for this OS.
func resolveShell(name string) string {
	if name != "auto" {
		return name
	}
	switch runtime.GOOS {
	case "linux":
		return "qt"
	case "windows":
		return "webview2"
	}
	return "browser"
}

// openWindow runs the native shell until it exits, or the default browser when the shell is missing.
func openWindow(ctx context.Context, shell, url string) error {
	if shell == "browser" {
		return openInBrowser(ctx, url)
	}
	bin, err := shellPath()
	if err != nil {
		logger.Warn("shell", "%v, opening the default browser instead", err)
		return openInBrowser(ctx, url)
	}
	return runShell(ctx, bin, url)
}

// shellPath returns the sibling shell binary and fails if it is not next to the launcher.
func shellPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	name := shellBinaryName
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	path := filepath.Join(filepath.Dir(exe), name)
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("shell binary %s not found next to the launcher", name)
	}
	return path, nil
}

// runShell waits for the shell to exit. A stop request from ctx is not an error.
func runShell(ctx context.Context, bin, url string) error {
	logger.Info("shell", "starting %s", filepath.Base(bin))
	cmd := exec.CommandContext(ctx, bin, url)
	cmd.Cancel = closeProcess(cmd)
	cmd.WaitDelay = shellExitTimeout
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		return nil
	}
	return err
}

// closeProcess asks the shell to close with an interrupt. Windows has no such signal, so it kills.
func closeProcess(cmd *exec.Cmd) func() error {
	if runtime.GOOS == "windows" {
		return func() error { return cmd.Process.Kill() }
	}
	return func() error { return cmd.Process.Signal(os.Interrupt) }
}

func openInBrowser(ctx context.Context, url string) error {
	logger.Info("shell", "opening the default browser")
	if err := openBrowser(url); err != nil {
		return err
	}
	<-ctx.Done()
	return nil
}
