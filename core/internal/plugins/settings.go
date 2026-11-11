package plugins

import (
	"path/filepath"
)

// SettingsView is what the settings form needs: the schema from plugin.json, null when the plugin has none,
// and the values the user saved, empty when none were saved.
type SettingsView struct {
	Schema map[string]any `json:"schema"`
	Values map[string]any `json:"values"`
}

// SettingsGet returns the settings schema and the saved values of an installed plugin.
func (h *Host) SettingsGet(id string) (SettingsView, error) {
	if err := requireID(id); err != nil {
		return SettingsView{}, err
	}
	inst, err := h.loadInstalled(id)
	if err != nil {
		return SettingsView{}, err
	}
	schema, err := loadSettingsSchema(inst)
	if err != nil {
		return SettingsView{}, err
	}
	saved := map[string]any{}
	if err := readJSONFile(h.settingsPath(id), &saved); err != nil {
		return SettingsView{}, err
	}
	return SettingsView{Schema: schema, Values: applySettingsDefaults(schema, saved)}, nil
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
	if values == nil {
		values = map[string]any{}
	}
	if err := validateSettings(schema, values); err != nil {
		return err
	}
	return writeJSONFile(h.settingsPath(id), values)
}

func (h *Host) settingsPath(id string) string {
	return filepath.Join(h.layout.DataDir(id), settingsFileName)
}
