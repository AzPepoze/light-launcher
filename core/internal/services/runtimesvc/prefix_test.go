package runtimesvc

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"light-launcher/core/internal/activity"
	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

// isolateConfigHome points the config folder at a temp dir and returns the default prefix folder.
func isolateConfigHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("APPDATA", home)
	return config.GetPrefixBaseDirectory()
}

func newTestService(t *testing.T) *service {
	t.Helper()
	settings := config.NewSettingsStore(config.SettingsFilePath())
	return newService(
		settings,
		config.NewGameStore(config.GetConfigDirectory()),
		activity.New(config.ActivityFilePath(), settings),
		config.NewPrefixStore(settings),
	)
}

func mkdirAll(t *testing.T, target string) {
	t.Helper()
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", target, err)
	}
}

func writeFile(t *testing.T, target string, content string) {
	t.Helper()
	mkdirAll(t, filepath.Dir(target))
	if err := os.WriteFile(target, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", target, err)
	}
}

func TestListPrefixesWithoutBaseFolderReturnsDefault(t *testing.T) {
	isolateConfigHome(t)
	svc := newTestService(t)

	prefixes, err := svc.listPrefixes()
	if err != nil {
		t.Fatalf("listPrefixes: %v", err)
	}
	if want := []string{"Default"}; !reflect.DeepEqual(prefixes, want) {
		t.Errorf("prefixes = %v, want %v", prefixes, want)
	}
}

func TestListPrefixesListsFoldersWithDefaultFirst(t *testing.T) {
	base := isolateConfigHome(t)
	svc := newTestService(t)
	mkdirAll(t, filepath.Join(base, "Zeta"))
	mkdirAll(t, filepath.Join(base, "Game A"))
	writeFile(t, filepath.Join(base, "notes.txt"), "not a prefix")

	prefixes, err := svc.listPrefixes()
	if err != nil {
		t.Fatalf("listPrefixes: %v", err)
	}
	if want := []string{"Default", "Game A", "Zeta"}; !reflect.DeepEqual(prefixes, want) {
		t.Errorf("prefixes = %v, want %v", prefixes, want)
	}
}

func TestListPrefixesUsesCustomFolderFromSettings(t *testing.T) {
	isolateConfigHome(t)
	custom := t.TempDir()
	mkdirAll(t, filepath.Join(custom, "Custom One"))
	if err := config.NewSettingsStore(config.SettingsFilePath()).Save(types.AppSettings{
		CustomPrefixDir:   custom,
		ScanFolders:       []string{},
		ScanFolderConfigs: []types.ScanFolderConfig{},
		Blacklist:         []string{},
	}); err != nil {
		t.Fatal(err)
	}
	svc := newTestService(t)

	prefixes, err := svc.listPrefixes()
	if err != nil {
		t.Fatalf("listPrefixes: %v", err)
	}
	if want := []string{"Default", "Custom One"}; !reflect.DeepEqual(prefixes, want) {
		t.Errorf("prefixes = %v, want %v", prefixes, want)
	}
}

func TestCreatePrefixTrimsNameAndRejectsEmpty(t *testing.T) {
	base := isolateConfigHome(t)
	svc := newTestService(t)

	if err := svc.createPrefix("   "); err == nil {
		t.Error("createPrefix with blank name succeeded, want error")
	}
	if err := svc.createPrefix("  My Game  "); err != nil {
		t.Fatalf("createPrefix: %v", err)
	}
	if !config.PathExists(filepath.Join(base, "My Game")) {
		t.Error("trimmed prefix folder was not created")
	}
}

func TestRemovePrefixDeletesFolderAndKeepsDefault(t *testing.T) {
	base := isolateConfigHome(t)
	svc := newTestService(t)
	writeFile(t, filepath.Join(base, "Game", "drive_c", "file.dat"), "data")

	if err := svc.removePrefix("Default"); err == nil {
		t.Error("removing Default succeeded, want error")
	}
	if err := svc.removePrefix("Game"); err != nil {
		t.Fatalf("removePrefix: %v", err)
	}
	if config.PathExists(filepath.Join(base, "Game")) {
		t.Error("prefix folder still exists after removal")
	}
}

func TestRemovePrefixRejectsUnsafeNames(t *testing.T) {
	base := isolateConfigHome(t)
	svc := newTestService(t)
	writeFile(t, filepath.Join(base, "Keep", "file.dat"), "data")

	for _, name := range []string{"", ".", "..", "../Keep", "a/b"} {
		if err := svc.removePrefix(name); err == nil {
			t.Errorf("removePrefix(%q) succeeded, want error", name)
		}
	}
	if !config.PathExists(filepath.Join(base, "Keep", "file.dat")) {
		t.Error("an unsafe name removed existing prefix data")
	}
}

func TestGetPrefixStatsSumsFileSizes(t *testing.T) {
	base := isolateConfigHome(t)
	svc := newTestService(t)
	writeFile(t, filepath.Join(base, "Game", "a.bin"), "12345")
	writeFile(t, filepath.Join(base, "Game", "deep", "b.bin"), "123")

	stats := svc.getPrefixStats("Game")
	if stats.SizeBytes == nil || *stats.SizeBytes != 8 {
		t.Errorf("sizeBytes = %v, want 8", stats.SizeBytes)
	}

	missing := svc.getPrefixStats("Missing")
	if missing.SizeBytes != nil || missing.CreatedAt != nil {
		t.Errorf("missing prefix stats = %+v, want null size and creation time", missing)
	}
}

func TestSavePrefixConfigRoundTrip(t *testing.T) {
	isolateConfigHome(t)
	svc := newTestService(t)
	mkdirAll(t, filepath.Join(config.GetPrefixBaseDirectory(), "Game"))

	cfg := config.DefaultPrefixOptions("Game")
	cfg.ProtonPath = "GE-Proton9-25"
	if err := svc.savePrefixConfig("Game", cfg); err != nil {
		t.Fatalf("savePrefixConfig: %v", err)
	}

	loaded := svc.loadPrefixConfig("Game")
	if loaded.ProtonPath != "GE-Proton9-25" {
		t.Errorf("loaded ProtonPath = %q, want GE-Proton9-25", loaded.ProtonPath)
	}
}

func TestLoadPrefixConfigDefaultsWhenMissing(t *testing.T) {
	isolateConfigHome(t)
	svc := newTestService(t)

	cfg := svc.loadPrefixConfig("Nothing")
	if cfg.Name != "Nothing" || cfg.Extras.Gamescope.Width != config.DefaultWidth {
		t.Errorf("default config = %+v, want name Nothing and default gamescope width", cfg)
	}
}
