package plugins

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// writeDemoPackage writes an unsigned demo package at version whose plugin.json asks for the given network hosts.
func writeDemoPackage(t *testing.T, path, version string, hosts ...string) {
	t.Helper()
	files := testPackageFiles("demo", version)
	if len(hosts) > 0 {
		quoted := make([]string, len(hosts))
		for i, host := range hosts {
			quoted[i] = `"` + host + `"`
		}
		manifest := strings.Replace(testManifest("demo", version), `"network": ["127.0.0.1"]`,
			`"network": [`+strings.Join(quoted, ", ")+`]`, 1)
		files["plugin.json"] = []byte(manifest)
	}
	writePackageFile(t, path, withChecksums(files))
}

func TestInstallNeedsAcceptedPermissions(t *testing.T) {
	in, _ := newTestInstaller(t)
	source := filepath.Join(t.TempDir(), "demo.llplugin")
	writeDemoPackage(t, source, "1.0.0")
	ctx := context.Background()

	if _, err := in.Install(ctx, source, true, nil); !errors.Is(err, ErrConsent) {
		t.Fatalf("missing consent: err = %v, want ErrConsent", err)
	}
	if _, err := in.Install(ctx, source, true, []string{"network:other.example"}); !errors.Is(err, ErrConsent) {
		t.Fatalf("different consent: err = %v, want ErrConsent", err)
	}
	if _, err := in.Install(ctx, source, true, testConsent); err != nil {
		t.Fatalf("matching consent: %v", err)
	}
	if got := in.state.get("demo").AcceptedPermissions; !slices.Equal(got, testConsent) {
		t.Errorf("recorded consent = %v, want %v", got, testConsent)
	}
}

func TestUpdateNeedsConsentForNewPermissions(t *testing.T) {
	in, _ := newTestInstaller(t)
	source := filepath.Join(t.TempDir(), "demo.llplugin")
	writeDemoPackage(t, source, "1.0.0")
	ctx := context.Background()
	if _, err := in.Install(ctx, source, true, testConsent); err != nil {
		t.Fatal(err)
	}

	writeDemoPackage(t, source, "1.1.0", "127.0.0.1", "example.com")
	err := in.Update(ctx, "demo", true, testConsent)
	if !errors.Is(err, ErrNewPermissions) || !strings.Contains(err.Error(), "network:example.com") {
		t.Fatalf("update without the new permission: err = %v", err)
	}

	accepted := []string{"network:127.0.0.1", "network:example.com"}
	if err := in.Update(ctx, "demo", true, accepted); err != nil {
		t.Fatalf("update with the new permission accepted: %v", err)
	}
	if got := in.state.get("demo").AcceptedPermissions; !slices.Equal(got, accepted) {
		t.Errorf("recorded consent = %v, want %v", got, accepted)
	}
}

func TestUpdateWithoutNewPermissionsNeedsNoConsent(t *testing.T) {
	in, _ := newTestInstaller(t)
	source := filepath.Join(t.TempDir(), "demo.llplugin")
	writeDemoPackage(t, source, "1.0.0")
	ctx := context.Background()
	if _, err := in.Install(ctx, source, true, testConsent); err != nil {
		t.Fatal(err)
	}

	writeDemoPackage(t, source, "1.0.1")
	if err := in.Update(ctx, "demo", true, nil); err != nil {
		t.Fatalf("update with the same permissions: %v", err)
	}
}

func TestUpdateUsesInstalledManifestWhenConsentWasNotRecorded(t *testing.T) {
	in, _ := newTestInstaller(t)
	source := filepath.Join(t.TempDir(), "demo.llplugin")
	writeDemoPackage(t, source, "1.0.0")
	ctx := context.Background()
	if _, err := in.Install(ctx, source, true, testConsent); err != nil {
		t.Fatal(err)
	}
	if err := in.state.update("demo", func(state *PluginState) { state.AcceptedPermissions = nil }); err != nil {
		t.Fatal(err)
	}

	writeDemoPackage(t, source, "1.0.1")
	if err := in.Update(ctx, "demo", true, nil); err != nil {
		t.Fatalf("an installed plugin without a record is compared with its manifest: %v", err)
	}
}

func TestPermissionSummaryIsSortedAndComplete(t *testing.T) {
	summary := PermissionSummary(Permissions{
		Network: []string{"example.com"}, FS: []string{"$GAMES"}, Secrets: true, Browser: true, Exec: true,
	})
	want := []string{"browser", "exec", "fs:$GAMES", "network:example.com", "secrets"}
	if !slices.Equal(summary, want) {
		t.Errorf("summary = %v, want %v", summary, want)
	}
	if empty := PermissionSummary(Permissions{}); empty == nil || len(empty) != 0 {
		t.Errorf("no permissions = %#v, want an empty list", empty)
	}
}
