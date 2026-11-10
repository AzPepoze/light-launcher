package runtimesvc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/types"
)

// loadJSON reads a JSON file into T. Empty files are an error, as in the Electron app.
func loadJSON[T any](filePath string) (T, error) {
	var value T
	data, err := os.ReadFile(filePath)
	if err != nil {
		return value, err
	}
	if strings.TrimSpace(string(data)) == "" {
		return value, fmt.Errorf("File %s is empty", filePath)
	}
	err = json.Unmarshal(data, &value)
	return value, err
}

// saveJSON writes value as indented JSON through a temp file, so readers never see a partial file.
func saveJSON(filePath string, value any) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}
	data, err := marshalIndentJS(value)
	if err != nil {
		return err
	}

	temp, err := os.CreateTemp(filepath.Dir(filePath), filepath.Base(filePath)+".tmp.*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		os.Remove(tempPath)
		return err
	}
	if err := temp.Close(); err != nil {
		os.Remove(tempPath)
		return err
	}
	if err := os.Rename(tempPath, filePath); err != nil {
		os.Remove(tempPath)
		return err
	}
	return nil
}

// marshalIndentJS matches JSON.stringify(value, null, 2): no HTML escaping and no trailing newline.
func marshalIndentJS(value any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buffer.Bytes(), []byte("\n")), nil
}

// loadAppSettings reads settings.json. Missing or broken files give empty settings.
// Unlike the Electron app, it does not rewrite the file to backfill defaults.
func loadAppSettings() types.AppSettings {
	var settings types.AppSettings
	settingsPath := filepath.Join(config.GetBaseDirectory(), "settings.json")
	if !pathExists(settingsPath) {
		return settings
	}
	loaded, err := loadJSON[types.AppSettings](settingsPath)
	if err != nil {
		logger.Error("Config", "Error loading app settings: %v", err)
		return settings
	}
	return loaded
}
