package library

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

// Finds the launcher inside an install folder, ported from the old tree.ts heuristics.

// badExecutable matches installers, uninstallers, crash handlers and redistributables.
var badExecutable = regexp.MustCompile(`(?i)(unins|uninstall|crash|crashhandler|report|redist|vcredist|vc_redist|dxsetup|directx|dotnet|prereq|eula|readme|license|notification|_helper|helper\.exe)`)

var junkNames = map[string]bool{".DS_Store": true, "Thumbs.db": true, "__MACOSX": true}

type executableCandidate struct {
	path  string
	depth int
	size  int64
	isExe bool
}

// executableSuffixes are the launcher file types for the current platform.
func executableSuffixes() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{".exe"}
	case "darwin":
		return nil
	default:
		// .exe stays in the list so Windows-only builds can still run through wine
		return []string{".appimage", ".sh", ".x86_64", ".run", ".exe"}
	}
}

// FindExecutables lists launchers under dir, best first: shallowest, native before .exe, then largest.
func FindExecutables(dir string) []string {
	var found []executableCandidate
	_ = filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || path == dir {
			return nil
		}
		if junkNames[entry.Name()] {
			return skipEntry(entry)
		}
		if entry.IsDir() {
			return visitDir(&found, dir, path, entry)
		}
		if candidate, ok := fileCandidate(dir, path, entry); ok {
			found = append(found, candidate)
		}
		return nil
	})
	sort.SliceStable(found, func(i, j int) bool {
		a, b := found[i], found[j]
		if a.isExe != b.isExe {
			return !a.isExe
		}
		if a.depth != b.depth {
			return a.depth < b.depth
		}
		return a.size > b.size
	})
	paths := make([]string, len(found))
	for i, candidate := range found {
		paths[i] = candidate.path
	}
	return paths
}

// DetectExecutable returns the best launcher in dir, or "" when there is none.
func DetectExecutable(dir string) string {
	paths := FindExecutables(dir)
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

func skipEntry(entry fs.DirEntry) error {
	if entry.IsDir() {
		return filepath.SkipDir
	}
	return nil
}

// visitDir treats a macOS .app bundle as one launcher and descends into any other folder.
func visitDir(found *[]executableCandidate, root, path string, entry fs.DirEntry) error {
	name := entry.Name()
	if runtime.GOOS == "darwin" && strings.HasSuffix(strings.ToLower(name), ".app") && !badExecutable.MatchString(name) {
		*found = append(*found, executableCandidate{path: path, depth: depthOf(root, path), size: dirSize(path)})
		return filepath.SkipDir
	}
	return nil
}

func fileCandidate(root, path string, entry fs.DirEntry) (executableCandidate, bool) {
	if !entry.Type().IsRegular() || badExecutable.MatchString(entry.Name()) {
		return executableCandidate{}, false
	}
	lower := strings.ToLower(entry.Name())
	if !hasAnySuffix(lower, executableSuffixes()) {
		return executableCandidate{}, false
	}
	info, err := entry.Info()
	if err != nil {
		return executableCandidate{}, false
	}
	return executableCandidate{
		path:  path,
		depth: depthOf(root, path),
		size:  info.Size(),
		isExe: strings.HasSuffix(lower, ".exe"),
	}, true
}

func hasAnySuffix(name string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

// depthOf counts path segments below root; a file directly in root has depth 1.
func depthOf(root, path string) int {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return 0
	}
	return strings.Count(rel, string(filepath.Separator)) + 1
}

// dirSize sums the sizes of all regular files under dir.
func dirSize(dir string) int64 {
	var total int64
	_ = filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || !entry.Type().IsRegular() {
			return nil
		}
		if info, statErr := os.Stat(path); statErr == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}
