package appsvc

import (
	"os"
	"strconv"
	"strings"
	"syscall"
)

// totalMemoryBytes reads MemTotal from /proc/meminfo, then falls back to sysinfo.
func totalMemoryBytes() (uint64, error) {
	if data, err := os.ReadFile("/proc/meminfo"); err == nil {
		if kb, ok := parseMemTotalKB(string(data)); ok {
			return kb * 1024, nil
		}
	}
	var info syscall.Sysinfo_t
	if err := syscall.Sysinfo(&info); err != nil {
		return 0, err
	}
	return uint64(info.Totalram) * uint64(info.Unit), nil
}

func parseMemTotalKB(meminfo string) (uint64, bool) {
	for _, line := range strings.Split(meminfo, "\n") {
		if !strings.HasPrefix(line, "MemTotal:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0, false
		}
		kb, err := strconv.ParseUint(fields[1], 10, 64)
		return kb, err == nil
	}
	return 0, false
}
