// Package uploads keeps the plugin files the UI sends until they are installed or the launcher starts again.
package uploads

import (
	"os"
	"path/filepath"
	"strings"
)

// Store holds one folder per upload under a single directory.
type Store struct {
	dir string
}

// Open empties dir, which removes uploads left by an earlier run, and returns a store over it.
func Open(dir string) (*Store, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if err := os.RemoveAll(abs); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, err
	}
	return &Store{dir: abs}, nil
}

// NewFolder makes an empty folder for one upload.
func (s *Store) NewFolder() (string, error) {
	return os.MkdirTemp(s.dir, "upload-")
}

// Remove deletes the upload folder that holds path. Paths outside the store are left alone.
func (s *Store) Remove(path string) error {
	top, ok := topFolder(s.dir, path)
	if !ok {
		return nil
	}
	return os.RemoveAll(filepath.Join(s.dir, top))
}

// topFolder returns the name of the folder directly under dir that holds path.
func topFolder(dir, path string) (string, bool) {
	rel, err := filepath.Rel(dir, path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	top, _, _ := strings.Cut(rel, string(filepath.Separator))
	return top, true
}
