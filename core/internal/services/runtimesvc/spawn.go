package runtimesvc

const setsidBinary = "/usr/bin/setsid"

// setsidCommand wraps a command in setsid when it is installed, so the child keeps running after the launcher exits.
func setsidCommand(command string, args []string) (string, []string) {
	if !pathExists(setsidBinary) {
		return command, args
	}
	return setsidBinary, append([]string{command}, args...)
}
