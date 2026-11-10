package plugins

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFileSecretsRoundTrip(t *testing.T) {
	ctx := context.Background()
	store := newFileSecrets(t.TempDir())

	if _, found, err := store.Get(ctx, "f95", "token"); err != nil || found {
		t.Fatalf("empty store: found=%v err=%v", found, err)
	}
	if err := store.Set(ctx, "f95", "token", "abc123"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	value, found, err := store.Get(ctx, "f95", "token")
	if err != nil || !found || value != "abc123" {
		t.Fatalf("Get = %q, %v, %v", value, found, err)
	}

	if err := store.Delete(ctx, "f95", "token"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, found, _ := store.Get(ctx, "f95", "token"); found {
		t.Fatalf("secret still present after Delete")
	}
}

func TestFileSecretsAreNamespacedByPlugin(t *testing.T) {
	ctx := context.Background()
	store := newFileSecrets(t.TempDir())
	if err := store.Set(ctx, "itch", "token", "itch-value"); err != nil {
		t.Fatal(err)
	}
	if _, found, _ := store.Get(ctx, "f95", "token"); found {
		t.Fatalf("plugin f95 read a secret that belongs to itch")
	}
}

func TestFileSecretsAreEncryptedOnDisk(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	store := newFileSecrets(dir)
	if err := store.Set(ctx, "demo", "token", "plain-text-marker"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, secretsFileName))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte("plain-text-marker")) {
		t.Fatalf("secret is stored in clear text")
	}
}

func TestFileSecretsWrongKeyFails(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	if err := newFileSecrets(dir).Set(ctx, "demo", "token", "value"); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, secretsKeyName)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := newFileSecrets(dir).Get(ctx, "demo", "token"); err == nil {
		t.Fatalf("a new key must not open the old file")
	}
}

func TestFileSecretsKeyFileIsPrivate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file modes are not enforced on Windows")
	}
	dir := t.TempDir()
	if err := newFileSecrets(dir).Set(context.Background(), "demo", "token", "value"); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(dir, secretsKeyName))
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Fatalf("key file mode = %o, want 600", mode)
	}
}
