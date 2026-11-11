package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"light-launcher/core/internal/api"
)

const testSettingsSchema = `{
  "type": "object",
  "required": ["region"],
  "additionalProperties": false,
  "properties": {
    "region": { "type": "string", "enum": ["eu", "us"] },
    "retries": { "type": "integer", "minimum": 0, "maximum": 5, "default": 2 },
    "verbose": { "type": "boolean", "default": true }
  }
}`

const listSettingsSchema = `{
  "type": "object",
  "properties": {
    "hosts": {
      "type": "array",
      "items": { "type": "string", "enum": ["a", "b", "c"] },
      "uniqueItems": true,
      "minItems": 1,
      "maxItems": 2
    },
    "label": { "type": "string", "minLength": 2, "maxLength": 4 }
  }
}`

func schemaFrom(t *testing.T, text string) map[string]any {
	t.Helper()
	var schema map[string]any
	if err := json.Unmarshal([]byte(text), &schema); err != nil {
		t.Fatal(err)
	}
	return schema
}

// fieldErrorsOf runs validation and returns the field errors as "field: message" lines.
func fieldErrorsOf(t *testing.T, schema map[string]any, values map[string]any) []string {
	t.Helper()
	err := validateSettings(schema, values)
	if err == nil {
		return nil
	}
	var fieldsErr *api.FieldsError
	if !errors.As(err, &fieldsErr) {
		t.Fatalf("error %T is not a FieldsError: %v", err, err)
	}
	lines := make([]string, 0, len(fieldsErr.Fields))
	for _, field := range fieldsErr.Fields {
		lines = append(lines, field.Field+": "+field.Message)
	}
	return lines
}

func TestValidateSettingsAcceptsMatchingValues(t *testing.T) {
	values := map[string]any{"region": "eu", "retries": 3.0, "verbose": true}
	if err := validateSettings(schemaFrom(t, testSettingsSchema), values); err != nil {
		t.Fatalf("validateSettings: %v", err)
	}
}

func TestValidateSettingsReportsEachBadField(t *testing.T) {
	schema := schemaFrom(t, testSettingsSchema)
	cases := map[string]struct {
		values map[string]any
		want   []string
	}{
		"missing required": {
			map[string]any{"retries": 1.0},
			[]string{"region: is required"},
		},
		"wrong type": {
			map[string]any{"region": "eu", "verbose": "yes"},
			[]string{"verbose: must be true or false"},
		},
		"not an integer": {
			map[string]any{"region": "eu", "retries": 1.5},
			[]string{"retries: must be a whole number"},
		},
		"not in enum": {
			map[string]any{"region": "asia"},
			[]string{"region: must be one of: eu, us"},
		},
		"below minimum": {
			map[string]any{"region": "eu", "retries": -1.0},
			[]string{"retries: must be at least 0"},
		},
		"above maximum": {
			map[string]any{"region": "eu", "retries": 9.0},
			[]string{"retries: must be at most 5"},
		},
		"unknown key": {
			map[string]any{"region": "eu", "extra": 1.0},
			[]string{"extra: is not a known setting"},
		},
		"all problems at once, sorted by field": {
			map[string]any{"retries": 9.0, "verbose": "x"},
			[]string{"region: is required", "retries: must be at most 5", "verbose: must be true or false"},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := fieldErrorsOf(t, schema, tc.values)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("errors = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestValidateSettingsChecksListsAndStrings(t *testing.T) {
	schema := schemaFrom(t, listSettingsSchema)
	cases := map[string]struct {
		values map[string]any
		want   []string
	}{
		"valid list": {
			map[string]any{"hosts": []any{"b", "a"}, "label": "abc"},
			nil,
		},
		"bad enum item": {
			map[string]any{"hosts": []any{"a", "z"}},
			[]string{"hosts: item 2 must be one of: a, b, c"},
		},
		"duplicate items": {
			map[string]any{"hosts": []any{"a", "a"}},
			[]string{"hosts: must not contain duplicates"},
		},
		"too many items": {
			map[string]any{"hosts": []any{"a", "b", "c"}},
			[]string{"hosts: must have at most 2 items"},
		},
		"empty list": {
			map[string]any{"hosts": []any{}},
			[]string{"hosts: must have at least 1 items"},
		},
		"string too short": {
			map[string]any{"label": "a"},
			[]string{"label: must be at least 2 characters"},
		},
		"string too long": {
			map[string]any{"label": "abcde"},
			[]string{"label: must be at most 4 characters"},
		},
		"list given as text": {
			map[string]any{"hosts": "a"},
			[]string{"hosts: must be a list"},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := fieldErrorsOf(t, schema, tc.values)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("errors = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestApplySettingsDefaultsFillsMissingKeysOnly(t *testing.T) {
	schema := schemaFrom(t, testSettingsSchema)
	saved := map[string]any{"region": "us", "retries": 4.0}
	got := applySettingsDefaults(schema, saved)
	want := map[string]any{"region": "us", "retries": 4.0, "verbose": true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("defaults = %v, want %v", got, want)
	}
	if _, changed := saved["verbose"]; changed {
		t.Fatal("applySettingsDefaults must not change the saved map")
	}
}

func TestApplySettingsDefaultsWithoutSchemaKeepsValues(t *testing.T) {
	got := applySettingsDefaults(nil, map[string]any{"x": 1.0})
	if !reflect.DeepEqual(got, map[string]any{"x": 1.0}) {
		t.Fatalf("got %v", got)
	}
}

func TestSettingsGetServiceAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	env := &pluginEnv{dataDir: dir, settingsSchema: schemaFrom(t, testSettingsSchema)}
	if err := writeJSONFile(filepath.Join(dir, settingsFileName), map[string]any{"region": "us"}); err != nil {
		t.Fatal(err)
	}
	reply, err := serviceSettingsGet(context.Background(), env, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := reply.(map[string]any)["settings"]
	want := map[string]any{"region": "us", "retries": 2.0, "verbose": true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("settings.get = %v, want %v", got, want)
	}
}

func TestWriteJSONFileReplacesAndLeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, settingsFileName)
	if err := writeJSONFile(path, map[string]any{"a": 1.0}); err != nil {
		t.Fatal(err)
	}
	if err := writeJSONFile(path, map[string]any{"b": 2.0}); err != nil {
		t.Fatal(err)
	}
	got := map[string]any{}
	if err := readJSONFile(path, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, map[string]any{"b": 2.0}) {
		t.Fatalf("file = %v, want the second write", got)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("folder has %d entries, want only the settings file", len(entries))
	}
}

func TestWriteJSONFileConcurrentWritersStayValid(t *testing.T) {
	path := filepath.Join(t.TempDir(), settingsFileName)
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := writeJSONFile(path, map[string]any{"n": float64(i)}); err != nil {
				t.Errorf("writeJSONFile: %v", err)
			}
		}()
	}
	wg.Wait()
	got := map[string]any{}
	if err := readJSONFile(path, &got); err != nil {
		t.Fatalf("file is not valid JSON after concurrent writes: %v", err)
	}
}

func TestSettingsSetNeedsInstalledPlugin(t *testing.T) {
	host := newTestHost(t)
	if err := host.SettingsSet("fixture", map[string]any{"x": 1.0}); err == nil {
		t.Fatalf("settings for a plugin that is not installed must fail")
	}
}
