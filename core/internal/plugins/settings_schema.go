package plugins

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
)

// loadSettingsSchema reads the JSON Schema named in contributes.settings. It returns nil when the plugin has none.
func loadSettingsSchema(inst installedPlugin) (map[string]any, error) {
	ref := inst.manifest.Contributes.Settings
	if ref == nil {
		return nil, nil
	}
	data, err := os.ReadFile(filepath.Join(inst.dir, filepath.FromSlash(ref.Ref)))
	if err != nil {
		return nil, fmt.Errorf("settings schema: %w", err)
	}
	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("settings schema is not valid JSON: %w", err)
	}
	return schema, nil
}

// validateSettings checks values against the parts of JSON Schema the launcher supports:
// type, enum, minimum, maximum, required, properties and additionalProperties.
func validateSettings(schema map[string]any, values map[string]any) error {
	return validateObject(schema, values, "settings")
}

func validateObject(schema map[string]any, value map[string]any, path string) error {
	for _, name := range requiredNames(schema) {
		if _, ok := value[name]; !ok {
			return fmt.Errorf("%s.%s is required", path, name)
		}
	}
	properties, _ := schema["properties"].(map[string]any)
	for key, item := range value {
		propertySchema, known := properties[key].(map[string]any)
		if !known {
			if schema["additionalProperties"] == false {
				return fmt.Errorf("%s.%s is not a known setting", path, key)
			}
			continue
		}
		if err := validateValue(propertySchema, item, path+"."+key); err != nil {
			return err
		}
	}
	return nil
}

func validateValue(schema map[string]any, value any, path string) error {
	if typeName, ok := schema["type"].(string); ok && !matchesType(typeName, value) {
		return fmt.Errorf("%s must be of type %s", path, typeName)
	}
	if enum, ok := schema["enum"].([]any); ok && !containsValue(enum, value) {
		return fmt.Errorf("%s has a value that is not allowed", path)
	}
	if number, ok := value.(float64); ok {
		if err := checkRange(schema, number, path); err != nil {
			return err
		}
	}
	if object, ok := value.(map[string]any); ok {
		return validateObject(schema, object, path)
	}
	return nil
}

func checkRange(schema map[string]any, number float64, path string) error {
	if minimum, ok := schema["minimum"].(float64); ok && number < minimum {
		return fmt.Errorf("%s must be at least %v", path, minimum)
	}
	if maximum, ok := schema["maximum"].(float64); ok && number > maximum {
		return fmt.Errorf("%s must be at most %v", path, maximum)
	}
	return nil
}

func requiredNames(schema map[string]any) []string {
	raw, _ := schema["required"].([]any)
	var names []string
	for _, item := range raw {
		if name, ok := item.(string); ok {
			names = append(names, name)
		}
	}
	return names
}

// matchesType reports whether a decoded JSON value has the schema type. Unknown types match anything.
func matchesType(typeName string, value any) bool {
	switch typeName {
	case "string":
		_, ok := value.(string)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "number":
		_, ok := value.(float64)
		return ok
	case "integer":
		number, ok := value.(float64)
		return ok && number == math.Trunc(number)
	case "array":
		_, ok := value.([]any)
		return ok
	case "object":
		_, ok := value.(map[string]any)
		return ok
	case "null":
		return value == nil
	default:
		return true
	}
}

func containsValue(options []any, value any) bool {
	for _, option := range options {
		if reflect.DeepEqual(option, value) {
			return true
		}
	}
	return false
}
