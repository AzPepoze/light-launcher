package plugins

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"

	"light-launcher/core/internal/api"
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

// applySettingsDefaults returns a copy of values with the schema default of every missing top-level property.
func applySettingsDefaults(schema map[string]any, values map[string]any) map[string]any {
	result := make(map[string]any, len(values))
	for key, value := range values {
		result[key] = value
	}
	properties, _ := schema["properties"].(map[string]any)
	for key, raw := range properties {
		property, _ := raw.(map[string]any)
		defaultValue, hasDefault := property["default"]
		if _, saved := result[key]; hasDefault && !saved {
			result[key] = defaultValue
		}
	}
	return result
}

// validateSettings checks values against the supported JSON Schema keywords. It returns an
// *api.FieldsError with one entry per bad field, or nil when every value is valid.
func validateSettings(schema map[string]any, values map[string]any) error {
	problems := checkObject(schema, values, "", nil)
	if len(problems) == 0 {
		return nil
	}
	return &api.FieldsError{Fields: problems}
}

func checkObject(schema map[string]any, value map[string]any, prefix string, problems []api.FieldError) []api.FieldError {
	for _, name := range requiredNames(schema) {
		if _, ok := value[name]; !ok {
			problems = append(problems, api.FieldError{Field: prefix + name, Message: "is required"})
		}
	}
	properties, _ := schema["properties"].(map[string]any)
	for _, key := range sortedKeys(value) {
		field := prefix + key
		propertySchema, known := properties[key].(map[string]any)
		if !known {
			if schema["additionalProperties"] == false {
				problems = append(problems, api.FieldError{Field: field, Message: "is not a known setting"})
			}
			continue
		}
		problems = checkValue(propertySchema, value[key], field, problems)
	}
	return problems
}

// checkValue checks one value. A type mismatch stops the other checks, so one mistake gives one message.
func checkValue(schema map[string]any, value any, field string, problems []api.FieldError) []api.FieldError {
	if typeName, ok := schema["type"].(string); ok && !matchesType(typeName, value) {
		return append(problems, api.FieldError{Field: field, Message: "must be " + describeType(typeName)})
	}
	if enum, ok := schema["enum"].([]any); ok && !containsValue(enum, value) {
		return append(problems, api.FieldError{Field: field, Message: "must be one of: " + joinOptions(enum)})
	}
	switch typed := value.(type) {
	case float64:
		return checkNumber(schema, typed, field, problems)
	case string:
		return checkString(schema, typed, field, problems)
	case []any:
		return checkArray(schema, typed, field, problems)
	case map[string]any:
		return checkObject(schema, typed, field+".", problems)
	}
	return problems
}

func checkNumber(schema map[string]any, number float64, field string, problems []api.FieldError) []api.FieldError {
	if minimum, ok := schema["minimum"].(float64); ok && number < minimum {
		problems = append(problems, api.FieldError{Field: field, Message: fmt.Sprintf("must be at least %v", minimum)})
	}
	if maximum, ok := schema["maximum"].(float64); ok && number > maximum {
		problems = append(problems, api.FieldError{Field: field, Message: fmt.Sprintf("must be at most %v", maximum)})
	}
	return problems
}

func checkString(schema map[string]any, text string, field string, problems []api.FieldError) []api.FieldError {
	length := utf8.RuneCountInString(text)
	if minLength, ok := schema["minLength"].(float64); ok && float64(length) < minLength {
		problems = append(problems, api.FieldError{Field: field, Message: fmt.Sprintf("must be at least %v characters", minLength)})
	}
	if maxLength, ok := schema["maxLength"].(float64); ok && float64(length) > maxLength {
		problems = append(problems, api.FieldError{Field: field, Message: fmt.Sprintf("must be at most %v characters", maxLength)})
	}
	return problems
}

// checkArray checks the list size, duplicates and each item. Item messages name the item number.
func checkArray(schema map[string]any, items []any, field string, problems []api.FieldError) []api.FieldError {
	if minItems, ok := schema["minItems"].(float64); ok && float64(len(items)) < minItems {
		problems = append(problems, api.FieldError{Field: field, Message: fmt.Sprintf("must have at least %v items", minItems)})
	}
	if maxItems, ok := schema["maxItems"].(float64); ok && float64(len(items)) > maxItems {
		problems = append(problems, api.FieldError{Field: field, Message: fmt.Sprintf("must have at most %v items", maxItems)})
	}
	if schema["uniqueItems"] == true && hasDuplicates(items) {
		problems = append(problems, api.FieldError{Field: field, Message: "must not contain duplicates"})
	}
	itemSchema, ok := schema["items"].(map[string]any)
	if !ok {
		return problems
	}
	for index, item := range items {
		var itemProblems []api.FieldError
		itemProblems = checkValue(itemSchema, item, field, itemProblems)
		for _, problem := range itemProblems {
			problems = append(problems, api.FieldError{Field: field, Message: fmt.Sprintf("item %d %s", index+1, problem.Message)})
		}
	}
	return problems
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

// describeType is the wording used after "must be" in an error message.
func describeType(typeName string) string {
	switch typeName {
	case "string":
		return "text"
	case "boolean":
		return "true or false"
	case "number":
		return "a number"
	case "integer":
		return "a whole number"
	case "array":
		return "a list"
	case "object":
		return "an object"
	default:
		return typeName
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

func hasDuplicates(items []any) bool {
	for i := range items {
		for j := i + 1; j < len(items); j++ {
			if reflect.DeepEqual(items[i], items[j]) {
				return true
			}
		}
	}
	return false
}

func joinOptions(options []any) string {
	texts := make([]string, 0, len(options))
	for _, option := range options {
		texts = append(texts, fmt.Sprint(option))
	}
	return strings.Join(texts, ", ")
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
