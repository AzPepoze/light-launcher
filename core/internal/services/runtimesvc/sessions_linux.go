//go:build linux

package runtimesvc

import (
	"os"
	"os/exec"
	"strconv"
)

// runningInstancePids finds instance processes with pgrep. The short name is a fallback, as in the Electron app.
func runningInstancePids() []int {
	output, err := exec.Command("pgrep", "light-launcher-instance").Output()
	if err != nil {
		output, _ = exec.Command("pgrep", "light-launcher-instan").Output()
	}
	return parsePids(string(output))
}

func readProcessCmdline(pid int) (string, error) {
	content, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/cmdline")
	return string(content), err
}
