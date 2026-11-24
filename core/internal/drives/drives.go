package drives

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"light-launcher/core/internal/logger"
)

const (
	mountInfoPath = "/proc/self/mountinfo"
	rootMount     = "/"
	// C: is the prefix's own drive and Z: is the Linux root, so drive letters are D: to Y:.
	firstLetter = 'd'
	lastLetter  = 'y'
)

var errNoFreeLetter = errors.New("no free drive letter in the prefix")

// EnsureGameDrive maps the disk that holds each path to a free drive letter in the prefix.
// Disks on the root filesystem are already visible as Z:, so they are skipped.
func EnsureGameDrive(prefixPath string, paths ...string) {
	mountinfo, err := os.ReadFile(mountInfoPath)
	if err != nil {
		logger.Warn("Drives", "Cannot read %s: %v", mountInfoPath, err)
		return
	}
	dosdevices := filepath.Join(prefixPath, "dosdevices")
	for _, path := range paths {
		mapPathDisk(dosdevices, string(mountinfo), path)
	}
}

// mapPathDisk links the disk of one path into dosdevices, unless it is already mapped.
func mapPathDisk(dosdevices string, mountinfo string, path string) {
	if path == "" {
		return
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		logger.Warn("Drives", "Cannot resolve %s: %v", path, err)
		return
	}
	mount := mountPointFor(mountinfo, resolved)
	if mount == "" || mount == rootMount {
		return
	}
	name, err := linkMount(dosdevices, mount)
	if err != nil {
		logger.Warn("Drives", "Cannot map %s to a drive letter: %v", mount, err)
		return
	}
	logger.Info("Drives", "Game disk %s is at %s", mount, name)
}

// linkMount returns the drive name for mount, creating a link when none exists yet.
func linkMount(dosdevices string, mount string) (string, error) {
	if err := os.MkdirAll(dosdevices, 0755); err != nil {
		return "", err
	}
	existing, used, err := scanLetters(dosdevices)
	if err != nil {
		return "", err
	}
	for _, drive := range existing {
		if drive.Target == mount {
			return drive.Letter, nil
		}
	}

	free, ok := freeLetter(used)
	if !ok {
		return "", errNoFreeLetter
	}
	name := string(free) + ":"
	if err := os.Symlink(mount, filepath.Join(dosdevices, name)); err != nil {
		return "", err
	}
	return name, nil
}

// freeLetter returns the first letter from D: to Y: that is not used.
func freeLetter(used map[byte]bool) (byte, bool) {
	for letter := byte(firstLetter); letter <= lastLetter; letter++ {
		if !used[letter] {
			return letter, true
		}
	}
	return 0, false
}

// driveLetter returns the lower-case letter of a name such as "D:", or false for other names.
func driveLetter(name string) (byte, bool) {
	if len(name) != 2 || name[1] != ':' {
		return 0, false
	}
	letter := strings.ToLower(name[:1])[0]
	return letter, letter >= 'a' && letter <= 'z'
}
