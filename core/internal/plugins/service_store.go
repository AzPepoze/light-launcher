package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const (
	storageFileName  = "storage.json"
	settingsFileName = "settings.json"
)

type storageParams struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

// serviceStorageGet reads one value from the plugin's key/value file.
func serviceStorageGet(_ context.Context, env *pluginEnv, params json.RawMessage) (any, error) {
	req, err := decodeParams[storageParams](params)
	if err != nil {
		return nil, err
	}
	if req.Key == "" {
		return nil, invalidParams("key is required")
	}
	env.storageMu.Lock()
	defer env.storageMu.Unlock()
	book, err := env.readStorage()
	if err != nil {
		return nil, internalError(err.Error())
	}
	raw, found := book[req.Key]
	if !found {
		return valueReply{Value: nil}, nil
	}
	return valueReply{Value: raw}, nil
}

// serviceStorageSet writes one value to the plugin's key/value file.
func serviceStorageSet(_ context.Context, env *pluginEnv, params json.RawMessage) (any, error) {
	req, err := decodeParams[storageParams](params)
	if err != nil {
		return nil, err
	}
	if req.Key == "" {
		return nil, invalidParams("key is required")
	}
	if len(req.Value) == 0 {
		return nil, invalidParams("value is required")
	}
	env.storageMu.Lock()
	defer env.storageMu.Unlock()
	book, err := env.readStorage()
	if err != nil {
		return nil, internalError(err.Error())
	}
	book[req.Key] = req.Value
	if err := writeJSONFile(filepath.Join(env.dataDir, storageFileName), book); err != nil {
		return nil, internalError(err.Error())
	}
	return nil, nil
}

// serviceSettingsGet returns {"settings": saved} for this plugin.
func serviceSettingsGet(_ context.Context, env *pluginEnv, _ json.RawMessage) (any, error) {
	settings := map[string]any{}
	if err := readJSONFile(filepath.Join(env.dataDir, settingsFileName), &settings); err != nil {
		return nil, internalError(err.Error())
	}
	return map[string]any{"settings": settings}, nil
}

func (env *pluginEnv) readStorage() (map[string]json.RawMessage, error) {
	book := map[string]json.RawMessage{}
	err := readJSONFile(filepath.Join(env.dataDir, storageFileName), &book)
	return book, err
}

// readJSONFile decodes a JSON file into target. A missing file leaves target as it is.
func readJSONFile(path string, target any) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// writeJSONFile replaces a JSON file through a temp file, so a crash never leaves half a file.
func writeJSONFile(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
