package drives

import (
	"path/filepath"
	"slices"
	"strings"
)

// Partition is a disk mounted from a real device, which can be bound as a drive.
type Partition struct {
	Name       string `json:"name"`
	Device     string `json:"device"`
	Mount      string `json:"mount"`
	FsType     string `json:"fsType"`
	TotalBytes uint64 `json:"totalBytes"`
	FreeBytes  uint64 `json:"freeBytes"`
}

// ListPartitions returns the user disks found in the mount table, one entry per device.
func ListPartitions() []Partition {
	return partitionsFromMountInfo(readMountInfo())
}

func partitionsFromMountInfo(mountinfo string) []Partition {
	partitions := []Partition{}
	seen := make(map[string]bool)
	for _, line := range strings.Split(mountinfo, "\n") {
		partition, ok := parsePartition(line)
		if !ok || seen[partition.Device] {
			continue
		}
		seen[partition.Device] = true
		partitions = append(partitions, partition)
	}
	return partitions
}

// parsePartition reads one mountinfo line. Fields after "-" hold the file system type and the source device.
func parsePartition(line string) (Partition, bool) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return Partition{}, false
	}
	separator := slices.Index(fields, "-")
	if separator < 0 || len(fields) < separator+3 {
		return Partition{}, false
	}
	mount := mountEscapes.Replace(fields[4])
	device := fields[separator+2]
	if !isUserDisk(device, mount) {
		return Partition{}, false
	}
	total, free := diskUsage(mount)
	return Partition{
		Name:       partitionName(device, mount),
		Device:     device,
		Mount:      mount,
		FsType:     fields[separator+1],
		TotalBytes: total,
		FreeBytes:  free,
	}, true
}

// isUserDisk keeps real disks and skips the system mounts and snap loop devices.
func isUserDisk(device string, mount string) bool {
	if !strings.HasPrefix(device, "/dev/") || strings.HasPrefix(device, "/dev/loop") {
		return false
	}
	return mount != rootMount && !strings.HasPrefix(mount, "/boot") && !strings.HasPrefix(mount, "/efi")
}

// partitionName prefers the disk label and falls back to the mount folder name.
func partitionName(device string, mount string) string {
	if label := deviceLabel(device); label != "" {
		return label
	}
	return filepath.Base(mount)
}
