package plugins

import (
	"path/filepath"
)

// SettingsGet returns the settings the user saved for a plugin. Empty when none were saved.
func (h *Host) SettingsGet(id string) (map[string]any, error) {
	if err := requireID(id); err != nil {
		return nil, err
	}
	settings := map[string]any{}
	if err := readJSONFile(h.settingsPath(id), &settings); err != nil {
		return nil, err
	}
	return settings, nil
}

// SettingsSet saves the settings after they pass the plugin's settings schema.
func (h *Host) SettingsSet(id string, values map[string]any) error {
	inst, err := h.loadInstalled(id)
	if err != nil {
		return err
	}
	schema, err := loadSettingsSchema(inst)
	if err != nil {
		return err
	}
	if schema != nil {
		if err := validateSettings(schema, values); err != nil {
			return err
		}
	}
	return writeJSONFile(h.settingsPath(id), values)
}

func (h *Host) settingsPath(id string) string {
	return filepath.Join(h.layout.DataDir(id), settingsFileName)
}
