package logger

import "os/exec"

// NotifySend shows msg as a desktop notification through notify-send. It returns at once.
// Register it with SetNotifier to get error popups.
func NotifySend(msg string) {
	mu.Lock()
	icon := notificationIcon
	mu.Unlock()

	args := notifyArgs(icon, msg)
	go func() {
		_ = exec.Command("notify-send", args...).Run()
	}()
}

func notifyArgs(icon, msg string) []string {
	args := []string{"-a", "LightLauncher"}
	if icon != "" {
		args = append(args, "-i", icon)
	}
	return append(args, "LightLauncher error", msg)
}
