package system

import "os/exec"

func IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
