package downloads

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

const macOSJunkDir = "__MACOSX"

// junkFiles are OS metadata files that never belong to a game.
var junkFiles = map[string]bool{".DS_Store": true, "Thumbs.db": true}

// badExecutable matches uninstallers, crash handlers and redistributables that are never the game.
var badExecutable = regexp.MustCompile(`(?i)(unins|uninstall|crash|crashhandler|report|redist|vcredist|vc_redist|dxsetup|directx|dotnet|prereq|eula|readme|license|notification|_helper|helper\.exe)`)

// isJunk reports whether a directory entry is OS metadata.
func isJunk(name string) bool {
	return name == macOSJunkDir || junkFiles[name]
}

// listFiles returns the relative paths of every regular file under dir, skipping junk.
func listFiles(dir string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if isJunk(d.Name()) && p != dir {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type().IsRegular() {
			rel, relErr := filepath.Rel(dir, p)
			if relErr != nil {
				return relErr
			}
			out = append(out, rel)
		}
		return nil
	})
	return out, err
}

// treeSize is the total size of the files under dir, ignoring junk.
func treeSize(dir string) int64 {
	files, _ := listFiles(dir)
	var total int64
	for _, rel := range files {
		if info, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			total += info.Size()
		}
	}
	return total
}

// cleanTree removes macOS metadata folders and OS junk files under dir.
func cleanTree(dir string) error {
	return filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || p == dir {
			return nil
		}
		switch {
		case d.IsDir() && d.Name() == macOSJunkDir:
			if err := os.RemoveAll(p); err != nil {
				return err
			}
			return filepath.SkipDir
		case !d.IsDir() && junkFiles[d.Name()]:
			return os.Remove(p)
		}
		return nil
	})
}

// executable is a plausible launcher found under an install folder.
type executable struct {
	path  string
	rel   string
	size  int64
	depth int
	rank  int // 0 for native launchers, 1 for Windows binaries run through Wine
}

// executableExts are the launcher extensions for this platform. Windows binaries stay launchable on Linux through Wine.
func executableExts() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{".exe"}
	case "darwin":
		return []string{".app"}
	default:
		return []string{".appimage", ".sh", ".x86_64", ".run", ".exe"}
	}
}

// findExecutables lists launchers under dir. The shallowest native launcher comes first.
func findExecutables(dir string) ([]executable, error) {
	exts := executableExts()
	var found []executable
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || p == dir {
			return nil
		}
		if isJunk(d.Name()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 || badExecutable.MatchString(d.Name()) {
			return nil
		}
		lower := strings.ToLower(d.Name())
		if d.IsDir() {
			if strings.HasSuffix(lower, ".app") && hasExt(exts, ".app") {
				found = append(found, newExecutable(dir, p, treeSize(p)))
				return filepath.SkipDir
			}
			return nil
		}
		if hasSuffixIn(lower, exts) {
			info, statErr := d.Info()
			if statErr != nil {
				return nil
			}
			found = append(found, newExecutable(dir, p, info.Size()))
		}
		return nil
	})
	sort.SliceStable(found, func(i, j int) bool {
		a, b := found[i], found[j]
		if a.rank != b.rank {
			return a.rank < b.rank
		}
		if a.depth != b.depth {
			return a.depth < b.depth
		}
		return a.size > b.size
	})
	return found, err
}

func newExecutable(root, p string, size int64) executable {
	rel, _ := filepath.Rel(root, p)
	rank := 0
	if strings.HasSuffix(strings.ToLower(p), ".exe") {
		rank = 1
	}
	return executable{
		path:  p,
		rel:   rel,
		size:  size,
		depth: strings.Count(rel, string(filepath.Separator)) + 1,
		rank:  rank,
	}
}

func hasExt(exts []string, want string) bool {
	for _, ext := range exts {
		if ext == want {
			return true
		}
	}
	return false
}

func hasSuffixIn(name string, exts []string) bool {
	for _, ext := range exts {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}

// detectExecutable returns the best launcher under dir, or "" when there is none.
func detectExecutable(dir string) (string, error) {
	found, err := findExecutables(dir)
	if err != nil || len(found) == 0 {
		return "", err
	}
	return found[0].path, nil
}
