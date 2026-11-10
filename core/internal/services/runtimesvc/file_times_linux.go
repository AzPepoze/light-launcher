//go:build linux

package runtimesvc

import "golang.org/x/sys/unix"

// fileTimes returns birth, change and modify times in milliseconds. Birth is 0 when the filesystem does not report it.
func fileTimes(target string) (birthMs, changeMs, modifyMs float64, err error) {
	var stat unix.Statx_t
	mask := unix.STATX_BTIME | unix.STATX_CTIME | unix.STATX_MTIME
	if err := unix.Statx(unix.AT_FDCWD, target, 0, mask, &stat); err != nil {
		return 0, 0, 0, err
	}
	if stat.Mask&unix.STATX_BTIME != 0 {
		birthMs = timestampMs(stat.Btime)
	}
	return birthMs, timestampMs(stat.Ctime), timestampMs(stat.Mtime), nil
}

func timestampMs(timestamp unix.StatxTimestamp) float64 {
	return float64(timestamp.Sec)*1000 + float64(timestamp.Nsec)/1e6
}
