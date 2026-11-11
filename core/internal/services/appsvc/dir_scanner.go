package appsvc

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"light-launcher/core/internal/config"
)

// unlimitedDepth scans every subfolder.
const unlimitedDepth = -1

// scanOptions controls scanDirectoryFiles. A nil Extensions list accepts every file.
type scanOptions struct {
	MaxDepth        int
	Extensions      []string
	ExcludePatterns []string
}

// scannedFile is a file found by a scan, with the size and modification time taken from the directory entry.
type scannedFile struct {
	Path    string
	Size    int64
	ModTime time.Time
}

// scanDirectoryFiles lists files under dir that match the options. Unreadable folders are skipped.
func scanDirectoryFiles(dir string, options scanOptions) []scannedFile {
	root := config.CleanPath(dir)
	if root == "" {
		return []scannedFile{}
	}
	if _, err := os.Stat(root); err != nil {
		return []scannedFile{}
	}
	scanner := newDirScanner(options)
	scanner.walk(root, 0)
	return scanner.results
}

type dirScanner struct {
	maxDepth    int
	excludes    []*regexp.Regexp
	extensions  []string
	filterByExt bool
	results     []scannedFile
}

func newDirScanner(options scanOptions) *dirScanner {
	scanner := &dirScanner{
		maxDepth:    options.MaxDepth,
		excludes:    compileWildcardPatterns(options.ExcludePatterns),
		filterByExt: options.Extensions != nil,
		results:     []scannedFile{},
	}
	for _, ext := range options.Extensions {
		scanner.extensions = append(scanner.extensions, strings.ToLower(ext))
	}
	return scanner
}

func (s *dirScanner) walk(dir string, depth int) {
	if s.maxDepth != unlimitedDepth && depth > s.maxDepth {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		name := entry.Name()
		if s.isExcluded(name) {
			continue
		}
		fullPath := filepath.Join(dir, name)
		if entry.IsDir() {
			s.walk(fullPath, depth+1)
		} else if entry.Type().IsRegular() && s.matchesExtension(name) {
			s.addFile(fullPath, entry)
		}
	}
}

// addFile records a matching file. A file that vanished during the walk is skipped.
func (s *dirScanner) addFile(fullPath string, entry os.DirEntry) {
	info, err := entry.Info()
	if err != nil {
		return
	}
	s.results = append(s.results, scannedFile{Path: fullPath, Size: info.Size(), ModTime: info.ModTime()})
}

func (s *dirScanner) isExcluded(name string) bool {
	for _, regex := range s.excludes {
		if regex.MatchString(name) {
			return true
		}
	}
	return false
}

func (s *dirScanner) matchesExtension(name string) bool {
	if !s.filterByExt {
		return true
	}
	lowerName := strings.ToLower(name)
	for _, ext := range s.extensions {
		if strings.HasSuffix(lowerName, ext) {
			return true
		}
	}
	return false
}
