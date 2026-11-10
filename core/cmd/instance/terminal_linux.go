package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"light-launcher/core/internal/logger"
)

func findTerminal() string {
	if t := os.Getenv("TERMINAL"); t != "" {
		parts := strings.Fields(t)
		if len(parts) > 0 {
			if p, err := exec.LookPath(parts[0]); err == nil {
				return p
			}
		}
	}
	terms := []string{"kitty", "alacritty", "ghostty", "foot", "wezterm", "gnome-terminal", "konsole", "xfce4-terminal", "xterm", "x-terminal-emulator"}
	for _, t := range terms {
		if p, err := exec.LookPath(t); err == nil {
			return p
		}
	}
	return ""
}

// startLogTerminal opens a terminal window to display game logs
func startLogTerminal(logPath string, gamePID int) {
	if logPath == "" {
		logger.Warn("Terminal", "Cannot open terminal: logPath is empty")
		return
	}

	term := findTerminal()
	if term == "" {
		logger.Warn("Terminal", "No compatible terminal emulator found on system")
		return
	}

	termBase := filepath.Base(term)
	filterExpr := "setpriority|vk_xwayland_wait_ready|vk_wsi_force_swapchain"

	// Properly single-quote and escape logPath for bash
	safeLogPath := strings.ReplaceAll(logPath, "'", "'\\''")

	tailScript := fmt.Sprintf(
		"sleep 0.2; if [ -f '%s' ]; then cat '%s' | grep -E -v '%s'; fi; tail --pid %d -f '%s' 2>/dev/null | grep -E -v '%s'; echo; echo '======================================='; echo 'Process finished. Press Enter to close...'; read",
		safeLogPath, safeLogPath, filterExpr, gamePID, safeLogPath, filterExpr,
	)

	var cmd *exec.Cmd
	switch {
	case strings.Contains(termBase, "kitty"):
		cmd = exec.Command(term, "-T", "LightLauncher Logs", "bash", "-c", tailScript)
	case strings.Contains(termBase, "alacritty"):
		cmd = exec.Command(term, "-T", "LightLauncher Logs", "-e", "bash", "-c", tailScript)
	case strings.Contains(termBase, "ghostty") || strings.Contains(termBase, "foot") || strings.Contains(termBase, "wezterm"):
		cmd = exec.Command(term, "-e", "bash", "-c", tailScript)
	case strings.Contains(termBase, "gnome-terminal"):
		cmd = exec.Command(term, "--title=LightLauncher Logs", "--", "bash", "-c", tailScript)
	case strings.Contains(termBase, "konsole"):
		cmd = exec.Command(term, "-p", "tabtitle=LightLauncher Logs", "-e", "bash", "-c", tailScript)
	case strings.Contains(termBase, "xfce4-terminal"):
		cmd = exec.Command(term, "--title=LightLauncher Logs", "-e", fmt.Sprintf("bash -c %q", tailScript))
	default:
		cmd = exec.Command(term, "-e", "bash", "-c", tailScript)
	}

	if err := cmd.Start(); err != nil {
		logger.Error("Terminal", "Failed to launch terminal (%s): %v", termBase, err)
	} else {
		logger.Info("Terminal", "Spawned log viewer in %s (PID: %d)", termBase, cmd.Process.Pid)
		_ = cmd.Process.Release()
	}
}
