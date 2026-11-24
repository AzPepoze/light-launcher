//go:build !linux

package drives

// diskUsage reports no sizes outside Linux, where the mount table is not read.
func diskUsage(string) (uint64, uint64) {
	return 0, 0
}

// deviceLabel reports no labels outside Linux.
func deviceLabel(string) string {
	return ""
}
