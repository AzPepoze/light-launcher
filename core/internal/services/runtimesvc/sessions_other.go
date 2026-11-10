//go:build !linux

package runtimesvc

import "errors"

// runningInstancePids finds no sessions outside Linux, where pgrep and /proc are not used.
func runningInstancePids() []int {
	return nil
}

func readProcessCmdline(pid int) (string, error) {
	return "", errors.New("process command lines are not read on this platform")
}
