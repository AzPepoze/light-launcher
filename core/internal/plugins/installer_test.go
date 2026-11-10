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

func TestInstallRequiresAllowUnsignedForUnsignedPackage(t *testing.T) {
	in, layout := newTestInstaller(t)
	source := filepath.Join(t.TempDir(), "demo.llplugin")
	writePackageFile(t, source, withChecksums(testPackageFiles("demo", "1.0.0")))

	if _, err := in.Install(context.Background(), source, false); err != ErrUnsigned {
		t.Fatalf("Install without allowUnsigned: err = %v, want ErrUnsigned", err)
	}
	if exists(layout.PluginDir("demo")) {
		t.Fatalf("a refused install left files behind")
	}

	id, err := in.Install(context.Background(), source, true)
	if err != nil {
		t.Fatalf("Install with allowUnsigned: %v", err)
	}
	if id != "demo" {
		t.Fatalf("id = %q", id)
	}
	if current, _ := readPointer(layout.PluginDir("demo"), currentPointer); current != "1.0.0" {
		t.Fatalf("current = %q", current)
	}
	if !in.state.get("demo").Unsigned || in.state.get("demo").Source != source {
		t.Fatalf("state = %+v", in.state.get("demo"))
	}
}

func TestInstallRejectsTamperedPackage(t *testing.T) {
	in, _ := newTestInstaller(t)
	files := withChecksums(testPackageFiles("demo", "1.0.0"))
	files["bin/"+PlatformKey()+"/demo"] = []byte("changed after the checksums were made")
	source := filepath.Join(t.TempDir(), "demo.llplugin")
	writePackageFile(t, source, files)

	_, err := in.Install(context.Background(), source, true)
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("err = %v, want a checksum mismatch", err)
	}
}

func TestInstallRejectsPackageWithoutChecksums(t *testing.T) {
	in, _ := newTestInstaller(t)
	source := filepath.Join(t.TempDir(), "demo.llplugin")
	writePackageFile(t, source, testPackageFiles("demo", "1.0.0"))

	if _, err := in.Install(context.Background(), source, true); err == nil || !strings.Contains(err.Error(), "SHA256SUMS") {
		t.Fatalf("err = %v, want a missing SHA256SUMS error", err)
	}
}

func TestInstallRejectsFileMissingFromChecksums(t *testing.T) {
	in, _ := newTestInstaller(t)
	files := withChecksums(testPackageFiles("demo", "1.0.0"))
	files["extra.txt"] = []byte("not listed")
	source := filepath.Join(t.TempDir(), "demo.llplugin")
	writePackageFile(t, source, files)

	if _, err := in.Install(context.Background(), source, true); err == nil || !strings.Contains(err.Error(), "not listed") {
		t.Fatalf("err = %v, want a not-listed error", err)
	}
}

func TestInstallAcceptsPackageSignedByTrustedKey(t *testing.T) {
	in, layout := newTestInstaller(t)
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	writeTrustedKey(t, layout, public)
	source := filepath.Join(t.TempDir(), "demo.llplugin")
	writePackageFile(t, source, withSignature(withChecksums(testPackageFiles("demo", "1.0.0")), private))

	if _, err := in.Install(context.Background(), source, false); err != nil {
		t.Fatalf("signed install: %v", err)
	}
	if in.state.get("demo").Unsigned {
		t.Fatalf("a trusted signature must not mark the plugin unsigned")
	}
}

func TestInstallRejectsSignatureFromUnknownKey(t *testing.T) {
	in, _ := newTestInstaller(t)
	_, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "demo.llplugin")
	writePackageFile(t, source, withSignature(withChecksums(testPackageFiles("demo", "1.0.0")), private))

	if _, err := in.Install(context.Background(), source, true); err == nil || !strings.Contains(err.Error(), "trusted key") {
		t.Fatalf("err = %v, want a trusted key error", err)
	}
}

func TestInstallRejectsUnsupportedAPIVersionAndNewerMinHost(t *testing.T) {
	cases := map[string]string{
		"api":     strings.Replace(testManifest("demo", "1.0.0"), `"apiVersion": "1"`, `"apiVersion": "2"`, 1),
		"minHost": strings.Replace(testManifest("demo", "1.0.0"), `"minHost": "2.0.0"`, `"minHost": "9.0.0"`, 1),
	}
	for name, manifest := range cases {
		t.Run(name, func(t *testing.T) {
			in, _ := newTestInstaller(t)
			files := testPackageFiles("demo", "1.0.0")
			files["plugin.json"] = []byte(manifest)
			source := filepath.Join(t.TempDir(), "demo.llplugin")
			writePackageFile(t, source, withChecksums(files))
			if _, err := in.Install(context.Background(), source, true); err == nil {
				t.Fatalf("expected an error")
			}
		})
	}
}

func TestPreviewReturnsManifestWithoutInstalling(t *testing.T) {
	in, layout := newTestInstaller(t)
	source := filepath.Join(t.TempDir(), "demo.llplugin")
	writePackageFile(t, source, withChecksums(testPackageFiles("demo", "1.0.0")))

	preview, err := in.Preview(context.Background(), source)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Manifest.ID != "demo" || !preview.Unsigned {
		t.Fatalf("preview = %+v", preview)
	}
	if exists(layout.PluginDir("demo")) {
		t.Fatalf("preview installed the plugin")
	}
}

func TestUpdateAndRollback(t *testing.T) {
	in, layout := newTestInstaller(t)
	ctx := context.Background()
	source := filepath.Join(t.TempDir(), "demo.llplugin")
	writePackageFile(t, source, withChecksums(testPackageFiles("demo", "1.0.0")))
	if _, err := in.Install(ctx, source, true); err != nil {
		t.Fatal(err)
	}

	if err := in.Update(ctx, "demo", true); err != ErrNoUpdate {
		t.Fatalf("update to the same version: err = %v, want ErrNoUpdate", err)
	}

	writePackageFile(t, source, withChecksums(testPackageFiles("demo", "1.1.0")))
	if err := in.Update(ctx, "demo", true); err != nil {
		t.Fatalf("Update: %v", err)
	}
	assertPointers(t, layout, "1.1.0", "1.0.0")

	if err := in.Rollback("demo"); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	assertPointers(t, layout, "1.0.0", "1.1.0")
}

func TestRollbackWithoutPreviousFails(t *testing.T) {
	in, _ := newTestInstaller(t)
	source := filepath.Join(t.TempDir(), "demo.llplugin")
	writePackageFile(t, source, withChecksums(testPackageFiles("demo", "1.0.0")))
	if _, err := in.Install(context.Background(), source, true); err != nil {
		t.Fatal(err)
	}
	if err := in.Rollback("demo"); err != ErrNoPrevious {
		t.Fatalf("err = %v, want ErrNoPrevious", err)
	}
}

func TestRemoveDeletesFilesAndOptionalData(t *testing.T) {
	in, layout := newTestInstaller(t)
	source := filepath.Join(t.TempDir(), "demo.llplugin")
	writePackageFile(t, source, withChecksums(testPackageFiles("demo", "1.0.0")))
	if _, err := in.Install(context.Background(), source, true); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(layout.DataDir("demo"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := in.Remove("demo", false); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if exists(layout.PluginDir("demo")) || !exists(layout.DataDir("demo")) {
		t.Fatalf("Remove without deleteData must keep the data folder")
	}
	if in.state.get("demo").Source != "" {
		t.Fatalf("state entry was not removed")
	}
	if err := in.Remove("demo", true); err != ErrNotInstalled {
		t.Fatalf("second Remove: err = %v, want ErrNotInstalled", err)
	}
}

func assertPointers(t *testing.T, layout Layout, wantCurrent, wantPrevious string) {
	t.Helper()
	current, err := readPointer(layout.PluginDir("demo"), currentPointer)
	if err != nil {
		t.Fatal(err)
	}
	previous, err := readPointer(layout.PluginDir("demo"), previousPointer)
	if err != nil {
		t.Fatal(err)
	}
	if current != wantCurrent || previous != wantPrevious {
		t.Fatalf("current = %q previous = %q, want %q and %q", current, previous, wantCurrent, wantPrevious)
	}
}
