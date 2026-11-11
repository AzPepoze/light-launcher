package plugins

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFolder writes files under a new folder and returns the folder path.
func writeFolder(t *testing.T, files map[string][]byte) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "drop")
	for name, data := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestInstallAcceptsZipFile(t *testing.T) {
	in, layout := newTestInstaller(t)
	source := filepath.Join(t.TempDir(), "demo.zip")
	writePackageFile(t, source, withChecksums(testPackageFiles("demo", "1.0.0")))

	if _, err := in.Install(context.Background(), source, true, testConsent); err != nil {
		t.Fatalf("Install zip: %v", err)
	}
	if !exists(layout.PluginDir("demo")) {
		t.Fatalf("zip install left nothing behind")
	}
}

func TestInstallFolderNeedsAllowUnsigned(t *testing.T) {
	in, layout := newTestInstaller(t)
	source := writeFolder(t, testPackageFiles("demo", "1.0.0"))

	if _, err := in.Install(context.Background(), source, false, testConsent); err != ErrUnsigned {
		t.Fatalf("Install folder without allowUnsigned: err = %v, want ErrUnsigned", err)
	}
	if exists(layout.PluginDir("demo")) {
		t.Fatalf("a refused folder install left files behind")
	}

	id, err := in.Install(context.Background(), source, true, testConsent)
	if err != nil {
		t.Fatalf("Install folder with allowUnsigned: %v", err)
	}
	if id != "demo" || !in.state.get("demo").Unsigned || in.state.get("demo").Source != source {
		t.Fatalf("id %q, state %+v", id, in.state.get("demo"))
	}
	if !exists(filepath.Join(layout.VersionDir("demo", "1.0.0"), "plugin.json")) {
		t.Fatalf("folder contents were not installed")
	}
}

func TestInstallFolderIgnoresItsOwnSignature(t *testing.T) {
	in, layout := newTestInstaller(t)
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	writeTrustedKey(t, layout, public)
	files := withSignature(withChecksums(testPackageFiles("demo", "1.0.0")), private)
	source := writeFolder(t, files)

	if _, err := in.Install(context.Background(), source, false, testConsent); err != ErrUnsigned {
		t.Fatalf("folder with a trusted SIGNATURE: err = %v, want ErrUnsigned", err)
	}
}

func TestInstallFolderWithoutPluginJSON(t *testing.T) {
	in, _ := newTestInstaller(t)
	files := testPackageFiles("demo", "1.0.0")
	delete(files, "plugin.json")
	source := writeFolder(t, files)

	if _, err := in.Install(context.Background(), source, true, testConsent); err == nil || !strings.Contains(err.Error(), "plugin.json") {
		t.Fatalf("err = %v, want a missing plugin.json error", err)
	}
}

func TestInstallFolderWithoutEntryBinary(t *testing.T) {
	in, _ := newTestInstaller(t)
	files := testPackageFiles("demo", "1.0.0")
	delete(files, "bin/"+PlatformKey()+"/demo")
	source := writeFolder(t, files)

	if _, err := in.Install(context.Background(), source, true, testConsent); err == nil || !strings.Contains(err.Error(), "not in the package") {
		t.Fatalf("err = %v, want a missing entry binary error", err)
	}
}
