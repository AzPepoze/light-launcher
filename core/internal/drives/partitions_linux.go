//go:build linux

package drives

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

const byLabelDir = "/dev/disk/by-label"

// labelEscapes undoes the \xNN escapes udev uses in label link names.
var labelEscapes = strings.NewReplacer(`\x20`, " ")

// diskUsage returns the size and free space of the file system mounted at mount.
func diskUsage(mount string) (uint64, uint64) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(mount, &stat); err != nil {
		return 0, 0
	}
	blockSize := uint64(stat.Bsize)
	return uint64(stat.Blocks) * blockSize, uint64(stat.Bavail) * blockSize
}

// deviceLabel returns the file system label of device, or "" when it has none.
func deviceLabel(device string) string {
	entries, err := os.ReadDir(byLabelDir)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		resolved, err := filepath.EvalSymlinks(filepath.Join(byLabelDir, entry.Name()))
		if err == nil && resolved == device {
			return labelEscapes.Replace(entry.Name())
		}
	}
	return ""
}
