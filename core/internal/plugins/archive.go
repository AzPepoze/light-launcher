package plugins

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	maxEntries        = 5000
	maxFileBytes      = 256 << 20
	maxExtractedBytes = 512 << 20
)

var drivePattern = regexp.MustCompile(`^[A-Za-z]:`)

// extractZip unpacks archivePath into dest. It refuses paths that leave dest, links,
// duplicate names and content above the size limits.
func extractZip(archivePath, dest string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open package: %w", err)
	}
	defer reader.Close()

	if len(reader.File) > maxEntries {
		return fmt.Errorf("package has %d entries, the limit is %d", len(reader.File), maxEntries)
	}
	seen := map[string]bool{}
	var total int64
	for _, file := range reader.File {
		name, err := safeEntryName(file.Name)
		if err != nil {
			return err
		}
		if seen[name] {
			return fmt.Errorf("package lists %q twice", name)
		}
		seen[name] = true
		if err := extractEntry(file, name, dest, &total); err != nil {
			return err
		}
	}
	return nil
}

// safeEntryName returns the cleaned slash path of a zip entry, or an error if it escapes the root.
func safeEntryName(name string) (string, error) {
	if name == "" || strings.ContainsAny(name, "\\") || strings.HasPrefix(name, "/") || drivePattern.MatchString(name) {
		return "", fmt.Errorf("unsafe package path %q", name)
	}
	clean := path.Clean(strings.TrimSuffix(name, "/"))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("unsafe package path %q", name)
	}
	return clean, nil
}

func extractEntry(file *zip.File, name, dest string, total *int64) error {
	target := filepath.Join(dest, filepath.FromSlash(name))
	if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
		return fmt.Errorf("unsafe package path %q", name)
	}
	mode := file.Mode()
	if file.FileInfo().IsDir() {
		return os.MkdirAll(target, 0o755)
	}
	if !mode.IsRegular() || mode&fs.ModeSymlink != 0 {
		return fmt.Errorf("package entry %q is not a regular file", name)
	}
	if file.UncompressedSize64 > maxFileBytes {
		return fmt.Errorf("package file %q is larger than %d bytes", name, maxFileBytes)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return writeZipFile(file, target, total)
}

func writeZipFile(file *zip.File, target string, total *int64) error {
	source, err := file.Open()
	if err != nil {
		return fmt.Errorf("read package file %q: %w", file.Name, err)
	}
	defer source.Close()

	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()

	written, err := io.Copy(out, io.LimitReader(source, maxFileBytes+1))
	if err != nil {
		return fmt.Errorf("extract %q: %w", file.Name, err)
	}
	if written > maxFileBytes {
		return fmt.Errorf("package file %q is larger than %d bytes", file.Name, maxFileBytes)
	}
	*total += written
	if *total > maxExtractedBytes {
		return fmt.Errorf("package expands past %d bytes", maxExtractedBytes)
	}
	return nil
}
