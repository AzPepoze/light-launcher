// Package store reads and writes versioned JSON files atomically.
package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

// ErrVersion reports a file written with a version this build does not understand.
var ErrVersion = errors.New("unsupported file version")

// File is one JSON document at a fixed path. It adds the version field on save and checks it on load.
type File struct {
	mu      sync.Mutex
	path    string
	version int
}

// NewFile opens one versioned JSON document at path.
func NewFile(path string, version int) *File {
	return &File{path: path, version: version}
}

// Load decodes the file into target. It reports whether the file exists.
// An empty, corrupt or wrong-version file is an error, so the caller can fall back to defaults.
func (f *File) Load(target any) (bool, error) {
	data, err := os.ReadFile(f.path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read %s: %w", f.path, err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return false, fmt.Errorf("read %s: file is empty", f.path)
	}

	var header struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return false, fmt.Errorf("decode %s: %w", f.path, err)
	}
	if header.Version != f.version && !(isLegacyVersion(header.Version) && f.version == 1) {
		return false, fmt.Errorf("%w: %s has version %d, want %d", ErrVersion, f.path, header.Version, f.version)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return false, fmt.Errorf("decode %s: %w", f.path, err)
	}
	return true, nil
}

// Save writes value as indented JSON with the version field. The write goes to a unique
// temp file in the same directory and is renamed into place, so readers never see a partial file.
func (f *File) Save(value any) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := encode(f.version, value)
	if err != nil {
		return err
	}
	if err := writeFile(f.path, data); err != nil {
		return fmt.Errorf("write %s: %w", f.path, err)
	}
	return nil
}

// Marshal matches JSON.stringify(value, null, 2): two-space indent, no HTML escaping and no trailing newline.
func Marshal(value any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buffer.Bytes(), []byte("\n")), nil
}

func encode(version int, value any) ([]byte, error) {
	body, err := Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode %T: %w", value, err)
	}
	body = bytes.TrimSpace(body)
	if len(body) < 2 || body[0] != '{' || body[len(body)-1] != '}' {
		return nil, fmt.Errorf("encode %T: value is not a JSON object", value)
	}

	fields := bytes.TrimSpace(body[1 : len(body)-1])
	var document []byte
	if len(fields) == 0 {
		document = []byte(`{"version": ` + strconv.Itoa(version) + `}`)
	} else {
		document = make([]byte, 0, len(body)+16)
		document = append(document, `{"version": `...)
		document = append(document, strconv.Itoa(version)...)
		document = append(document, ',')
		document = append(document, fields...)
		document = append(document, '}')
	}

	var indented bytes.Buffer
	if err := json.Indent(&indented, document, "", "  "); err != nil {
		return nil, fmt.Errorf("indent document: %w", err)
	}
	return indented.Bytes(), nil
}

// isLegacyVersion reports whether a file has no version field, so it uses the version 1 layout.
func isLegacyVersion(version int) bool {
	return version == 0
}

func writeFile(path string, data []byte) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}

	temp, err := os.CreateTemp(directory, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	renamed := false
	defer func() {
		if !renamed {
			_ = os.Remove(tempPath)
		}
	}()

	if err := temp.Chmod(0o644); err != nil {
		_ = temp.Close()
		return err
	}
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		return err
	}
	renamed = true

	// Best effort: make the rename durable where the platform allows syncing a directory.
	if dir, err := os.Open(directory); err == nil {
		_ = dir.Sync()
		_ = dir.Close()
	}
	return nil
}
