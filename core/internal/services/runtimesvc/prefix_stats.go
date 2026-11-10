package runtimesvc

import (
	"io/fs"
	"os"
	"path/filepath"
)

// PrefixStats is the size and creation time of one prefix. Unknown values are null.
type PrefixStats struct {
	Name      string   `json:"name"`
	CreatedAt *float64 `json:"createdAt"`
	SizeBytes *int64   `json:"sizeBytes"`
}

func getPrefixStats(name string) PrefixStats {
	stats := PrefixStats{Name: name}
	prefixPath := filepath.Join(prefixBaseFolder(), name)
	if !pathExists(prefixPath) {
		return stats
	}
	stats.CreatedAt = prefixCreatedAt(prefixPath)
	stats.SizeBytes = directorySize(prefixPath)
	return stats
}

// prefixCreatedAt returns the birth time in milliseconds, or nil when it is not reliable.
// A birth time equal to the change or modify time means the filesystem only reports a copy of one of them.
func prefixCreatedAt(prefixPath string) *float64 {
	birthMs, changeMs, modifyMs, err := fileTimes(prefixPath)
	if err != nil {
		return nil
	}
	hasValidBirth := birthMs >= 1 && birthMs != changeMs && birthMs != modifyMs
	if !hasValidBirth {
		return nil
	}
	return &birthMs
}

// directorySize sums file sizes under a folder. Symbolic links count as their own size and are not followed.
func directorySize(directory string) *int64 {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil
	}

	var total int64
	for _, entry := range entries {
		fullPath := filepath.Join(directory, entry.Name())
		if entry.IsDir() {
			if subSize := directorySize(fullPath); subSize != nil {
				total += *subSize
			}
			continue
		}
		if !isFileOrLink(entry.Type()) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		total += info.Size()
	}
	return &total
}

func isFileOrLink(mode fs.FileMode) bool {
	return mode.IsRegular() || mode&fs.ModeSymlink != 0
}
