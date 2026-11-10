package appsvc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// loadJSON reads a JSON file. An empty file is an error.
func loadJSON[T any](path string) (T, error) {
	var value T
	data, err := os.ReadFile(path)
	if err != nil {
		return value, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return value, fmt.Errorf("File %s is empty", path)
	}
	err = json.Unmarshal(data, &value)
	return value, err
}

// saveJSON writes the value through a temp file and renames it, so readers never see a partial file.
func saveJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := marshalJSON(value)
	if err != nil {
		return err
	}
	tempPath := fmt.Sprintf("%s.tmp.%d.%s", path, time.Now().UnixMilli(), randomHex(3))
	if err := os.WriteFile(tempPath, data, 0o644); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	return nil
}

// marshalJSON uses two-space indent and keeps <, > and & as they are, like JSON.stringify.
func marshalJSON(value any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buffer.Bytes(), []byte("\n")), nil
}
