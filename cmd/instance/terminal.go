package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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
	terms := []string{"kitty", "alacritty", "gnome-terminal", "konsole", "xfce4-terminal", "xterm"}
	for _, t := range terms {
		if p, err := exec.LookPath(t); err == nil {
			return p
		}
	}
	return ""
}

// startLogTerminal opens a terminal window to display game logs/
func startLogTerminal(logPath string, gamePID int) {
	term := findTerminal()
	if term == "" {
		return
	}

	filterExpr := "setpriority|vk_xwayland_wait_ready|vk_wsi_force_swapchain"
	tailCmd := fmt.Sprintf(
		"sleep 0.2; cat %s | grep -E -v '%s'; tail --pid %d -f %s | grep -E -v '%s'; echo; echo '---------------------------------------'; echo 'Process finished. Press Enter to close...'; read",
		logPath, filterExpr, gamePID, logPath, filterExpr,
	)

	var cmd *exec.Cmd
	if strings.Contains(term, "kitty") || strings.Contains(term, "alacritty") {
		cmd = exec.Command(term, "--", "bash", "-c", tailCmd)
	} else {
		cmd = exec.Command(term, "-e", "bash", "-c", tailCmd)
	}

	cmd.Start()
}
