package plugins

import (
	"encoding/json"
	"strings"
	"testing"
)

// manifestWith returns validManifest with the given top-level fields replaced.
func manifestWith(t *testing.T, fields map[string]any) []byte {
	t.Helper()
	var manifest map[string]any
	if err := json.Unmarshal([]byte(validManifest), &manifest); err != nil {
		t.Fatal(err)
	}
	for key, value := range fields {
		manifest[key] = value
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestParseManifestReadsOptionalInfo(t *testing.T) {
	data := manifestWith(t, map[string]any{
		"author":      "Ada",
		"description": "Finds demo games.",
		"homepage":    "https://example.com/demo",
	})
	manifest, err := ParseManifest(data)
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	if manifest.Author != "Ada" || manifest.Description != "Finds demo games." || manifest.Homepage != "https://example.com/demo" {
		t.Errorf("info = %q, %q, %q", manifest.Author, manifest.Description, manifest.Homepage)
	}
}

func TestParseManifestRejectsBadInfo(t *testing.T) {
	cases := map[string]map[string]any{
		"homepage without http": {"homepage": "ftp://example.com/demo"},
		"homepage not a url":    {"homepage": "not a url"},
		"author too long":       {"author": strings.Repeat("a", 81)},
		"description too long":  {"description": strings.Repeat("a", 1001)},
	}
	for name, fields := range cases {
		if _, err := ParseManifest(manifestWith(t, fields)); err == nil {
			t.Errorf("%s: must be rejected", name)
		}
	}
}
