package runtimesvc

import (
	"runtime"
	"testing"
)

// skipUnlessLinux skips tests of Linux-only features (Proton, /proc session scan) on other systems.
func skipUnlessLinux(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only feature")
	}
}
