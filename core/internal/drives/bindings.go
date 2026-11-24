package drives

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// systemLetter is the prefix's C: drive. Wine points it at wineSystemTarget by default.
const systemLetter = "c:"

const wineSystemTarget = "../drive_c"

// rootLetter is the Linux root that Wine maps as Z:. It is never changed.
const rootLetter = "z:"

var (
	ErrLetterTaken   = errors.New("that drive letter is already used")
	ErrInvalidLetter = errors.New("the drive letter must be C: or D: to Y:")
	ErrLockedLetter  = errors.New("Z: cannot be changed")
	ErrNotFolder     = errors.New("the path is not an existing folder or disk")
)

// PrefixDrive is one drive letter in a prefix and the folder or disk it points to.
type PrefixDrive struct {
	Letter string `json:"letter"`
	Target string `json:"target"`
	Kind   string `json:"kind"`
	Exists bool   `json:"exists"`
}

// ListBindings returns the drive letters in a prefix, sorted by letter.
func ListBindings(prefixPath string) ([]PrefixDrive, error) {
	drives, _, err := scanLetters(filepath.Join(prefixPath, "dosdevices"))
	if errors.Is(err, fs.ErrNotExist) {
		return []PrefixDrive{}, nil
	}
	return drives, err
}

// scanLetters returns the drive links in dosdevices and the letters they use.
func scanLetters(dosdevices string) ([]PrefixDrive, map[byte]bool, error) {
	entries, err := os.ReadDir(dosdevices)
	if err != nil {
		return nil, nil, err
	}
	mountinfo := readMountInfo()
	used := make(map[byte]bool)
	drives := []PrefixDrive{}
	for _, entry := range entries {
		letter, ok := driveLetter(entry.Name())
		if !ok {
			continue
		}
		used[letter] = true
		target, err := os.Readlink(filepath.Join(dosdevices, entry.Name()))
		if err != nil {
			continue
		}
		drives = append(drives, describeDrive(dosdevices, entry.Name(), target, mountinfo))
	}
	return drives, used, nil
}

func describeDrive(dosdevices string, name string, target string, mountinfo string) PrefixDrive {
	letter := strings.ToLower(name)
	return PrefixDrive{
		Letter: letter,
		Target: target,
		Kind:   driveKind(letter, target, mountinfo),
		Exists: pathExists(linkDestination(dosdevices, target)),
	}
}

// driveKind names what a drive points to: the system drive, the Linux root, a disk, or a folder.
func driveKind(letter string, target string, mountinfo string) string {
	switch {
	case letter == systemLetter && target == wineSystemTarget:
		return "system"
	case target == rootMount:
		return "root"
	case mountPointFor(mountinfo, target) == target:
		return "partition"
	default:
		return "folder"
	}
}

// normalizeLetter returns the lower-case name such as "d:" for a letter the setup may use.
func normalizeLetter(input string) (string, error) {
	name := strings.ToLower(strings.TrimSpace(input))
	if len(name) == 1 {
		name += ":"
	}
	if name == rootLetter {
		return "", ErrLockedLetter
	}
	letter, ok := driveLetter(name)
	if !ok || (letter != 'c' && (letter < firstLetter || letter > lastLetter)) {
		return "", ErrInvalidLetter
	}
	return name, nil
}

// resolveFolder returns the real path of an existing folder or disk, following symlinks.
func resolveFolder(folder string) (string, error) {
	resolved, err := filepath.EvalSymlinks(folder)
	if err != nil {
		return "", ErrNotFolder
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		return "", ErrNotFolder
	}
	return resolved, nil
}

// linkDestination makes a relative link target absolute, as Wine writes some of them relative to dosdevices.
func linkDestination(dosdevices string, target string) string {
	if filepath.IsAbs(target) {
		return target
	}
	return filepath.Join(dosdevices, target)
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readMountInfo() string {
	data, err := os.ReadFile(mountInfoPath)
	if err != nil {
		return ""
	}
	return string(data)
}
