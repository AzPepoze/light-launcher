package plugins

import (
	"encoding/json"
	"strings"
	"testing"
)

const testSettingsSchema = `{
  "type": "object",
  "required": ["region"],
  "additionalProperties": false,
  "properties": {
    "region": { "type": "string", "enum": ["eu", "us"] },
    "retries": { "type": "integer", "minimum": 0, "maximum": 5 },
    "verbose": { "type": "boolean" }
  }
}`

func schemaFor(t *testing.T) map[string]any {
	t.Helper()
	var schema map[string]any
	if err := json.Unmarshal([]byte(testSettingsSchema), &schema); err != nil {
		t.Fatal(err)
	}
	return schema
}

func TestValidateSettingsAcceptsMatchingValues(t *testing.T) {
	values := map[string]any{"region": "eu", "retries": 3.0, "verbose": true}
	if err := validateSettings(schemaFor(t), values); err != nil {
		t.Fatalf("validateSettings: %v", err)
	}
}

func TestValidateSettingsRejectsBadValues(t *testing.T) {
	cases := map[string]struct {
		values  map[string]any
		wantErr string
	}{
		"missing required": {map[string]any{"retries": 1.0}, "settings.region is required"},
		"wrong type":       {map[string]any{"region": "eu", "verbose": "yes"}, "settings.verbose must be of type boolean"},
		"not an integer":   {map[string]any{"region": "eu", "retries": 1.5}, "settings.retries must be of type integer"},
		"not in enum":      {map[string]any{"region": "asia"}, "settings.region has a value that is not allowed"},
		"below minimum":    {map[string]any{"region": "eu", "retries": -1.0}, "settings.retries must be at least 0"},
		"above maximum":    {map[string]any{"region": "eu", "retries": 9.0}, "settings.retries must be at most 5"},
		"unknown key":      {map[string]any{"region": "eu", "extra": 1.0}, "settings.extra is not a known setting"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := validateSettings(schemaFor(t), tc.values)
			if err == nil {
				t.Fatalf("expected an error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error %q does not contain %q", err, tc.wantErr)
			}
		})
	}
}

func TestSettingsSetNeedsInstalledPlugin(t *testing.T) {
	host := newTestHost(t)
	if err := host.SettingsSet("fixture", map[string]any{"x": 1.0}); err == nil {
		t.Fatalf("settings for a plugin that is not installed must fail")
	}
}
