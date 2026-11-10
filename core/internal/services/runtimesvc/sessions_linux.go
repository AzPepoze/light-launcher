//go:build linux

package runtimesvc

import (
	"os"
	"os/exec"
	"strconv"
)

// instancePattern matches the instance binary at the start of a command line. Plain pgrep names
// are cut to 15 characters by the kernel, so the full command line is matched instead.
const instancePattern = `(^|/)light-launcher-instance( |$)`

// runningInstancePids finds instance processes with pgrep.
func runningInstancePids() []int {
	output, _ := exec.Command("pgrep", "-f", instancePattern).Output()
	return parsePids(string(output))
}

func readProcessCmdline(pid int) (string, error) {
	content, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/cmdline")
	return string(content), err
}
