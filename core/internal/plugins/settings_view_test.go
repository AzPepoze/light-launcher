package plugins

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSettingsGetReturnsSchemaAndValues(t *testing.T) {
	host := newTestHost(t)
	ctx := context.Background()
	manifest := fixtureManifest(Permissions{})
	manifest.Contributes.Settings = &SettingsRef{Ref: "settings.schema.json"}
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.ReadFile(fixtureBinary)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{
		"plugin.json":                 manifestJSON,
		manifest.Entry[PlatformKey()]: binary,
		"settings.schema.json":        []byte(testSettingsSchema),
	}
	source := filepath.Join(t.TempDir(), "settings.llplugin")
	writePackageFile(t, source, withChecksums(files))

	info, err := host.Install(ctx, InstallRequest{Source: source, AllowUnsigned: true})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if !info.HasSettings {
		t.Error("a plugin with a settings schema must report hasSettings")
	}
	if err := host.SettingsSet("fixture", map[string]any{"region": "eu"}); err != nil {
		t.Fatal(err)
	}
	view, err := host.SettingsGet("fixture")
	if err != nil {
		t.Fatal(err)
	}
	if view.Schema["type"] != "object" {
		t.Errorf("schema = %v, want the object schema", view.Schema)
	}
	if view.Values["region"] != "eu" {
		t.Errorf("values = %v, want region eu", view.Values)
	}
}

func TestSettingsGetWithoutSchemaReturnsNullSchema(t *testing.T) {
	host := newTestHost(t)
	source := filepath.Join(t.TempDir(), "plain.llplugin")
	writeFixturePackage(t, source, "0.1.0")
	if _, err := host.Install(context.Background(), InstallRequest{Source: source, AllowUnsigned: true}); err != nil {
		t.Fatalf("Install: %v", err)
	}

	view, err := host.SettingsGet("fixture")
	if err != nil {
		t.Fatal(err)
	}
	if view.Schema != nil || view.Values == nil {
		t.Errorf("view = %+v, want a nil schema and empty values", view)
	}
}
