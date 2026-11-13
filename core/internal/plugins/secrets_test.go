package plugins

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"light-launcher/core/internal/contract"
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

func TestFileSecretsDeleteAllKeepsOtherPlugins(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	store := newFileSecrets(dir)
	for _, secret := range []struct{ plugin, key, value string }{
		{"demo", "token", "one"},
		{"demo", "apiKey", "two"},
		{"other", "token", "keep"},
	} {
		if err := store.Set(ctx, secret.plugin, secret.key, secret.value); err != nil {
			t.Fatal(err)
		}
	}

	if err := store.DeleteAll(ctx, "demo"); err != nil {
		t.Fatalf("DeleteAll: %v", err)
	}
	for _, key := range []string{"token", "apiKey"} {
		if _, found, err := store.Get(ctx, "demo", key); err != nil || found {
			t.Fatalf("secret %q after DeleteAll: found = %v, err = %v; want it deleted", key, found, err)
		}
	}
	if value, found, err := store.Get(ctx, "other", "token"); err != nil || !found || value != "keep" {
		t.Fatalf("another plugin's secret = %q, %v, %v; want it kept", value, found, err)
	}

	reopened := newFileSecrets(dir)
	if _, found, _ := reopened.Get(ctx, "demo", "token"); found {
		t.Fatalf("deleted secret came back after reopening the file")
	}
	if value, found, err := reopened.Get(ctx, "other", "token"); err != nil || !found || value != "keep" {
		t.Fatalf("reopened another plugin's secret = %q, %v, %v; want it kept", value, found, err)
	}
}

func TestFileSecretsDeleteAllWithoutSecretsIsNoOp(t *testing.T) {
	dir := t.TempDir()
	if err := newFileSecrets(dir).DeleteAll(context.Background(), "demo"); err != nil {
		t.Fatalf("DeleteAll on an empty store: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, secretsKeyName)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("DeleteAll created the key file: %v", err)
	}
}

func TestTrackedSecretsDeleteAllClearsTrackedKeys(t *testing.T) {
	ctx := context.Background()
	state, err := loadState(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	inner := newFakeSecrets(nil)
	tracked := trackedSecrets{inner: inner, state: state}
	for _, key := range []string{"token", "apiKey"} {
		if err := tracked.Set(ctx, "demo", key, "value"); err != nil {
			t.Fatal(err)
		}
	}

	if err := tracked.DeleteAll(ctx, "demo"); err != nil {
		t.Fatalf("DeleteAll: %v", err)
	}
	if keys := state.get("demo").SecretKeys; len(keys) != 0 {
		t.Fatalf("tracked keys after DeleteAll = %v, want none", keys)
	}
	if keys := inner.keys("demo"); len(keys) != 0 {
		t.Fatalf("stored keys after DeleteAll = %v, want none", keys)
	}
}

func TestTrackedSecretsDeleteAllKeepsKeysWhenStoreCannotList(t *testing.T) {
	ctx := context.Background()
	state, err := loadState(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	inner := newFakeSecrets(nil)
	inner.notListable = true
	tracked := trackedSecrets{inner: inner, state: state}
	for _, key := range []string{"token", "apiKey"} {
		if err := tracked.Set(ctx, "demo", key, "value"); err != nil {
			t.Fatal(err)
		}
	}

	if err := tracked.DeleteAll(ctx, "demo"); !errors.Is(err, contract.ErrSecretsNotListable) {
		t.Fatalf("DeleteAll error = %v, want ErrSecretsNotListable", err)
	}
	if keys := state.get("demo").SecretKeys; !slices.Equal(keys, []string{"token", "apiKey"}) {
		t.Fatalf("tracked keys = %v, want [token apiKey] kept for the fallback", keys)
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
