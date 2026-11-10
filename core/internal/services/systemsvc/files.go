package systemsvc

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// path maps an absolute system path under the service root.
func (s *Service) path(absolute string) string {
	return filepath.Join(s.root, absolute)
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// readText returns the file content, or "" when the file cannot be read.
func readText(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func readTrimmed(path string) string {
	return strings.TrimSpace(readText(path))
}

// readInt parses a small integer file such as a sysfs attribute. It reports false when the file is missing or malformed.
func readInt(path string) (int64, bool) {
	value, err := strconv.ParseInt(readTrimmed(path), 10, 64)
	return value, err == nil
}
