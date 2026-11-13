package plugins

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// testTrustedKey returns a deterministic public key for trusted key tests.
func testTrustedKey(t *testing.T, seed byte) ed25519.PublicKey {
	t.Helper()
	key := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{seed}, ed25519.SeedSize))
	return key.Public().(ed25519.PublicKey)
}

func trustedKeyText(key ed25519.PublicKey) string {
	return base64.StdEncoding.EncodeToString(key)
}

func wantFingerprint(key ed25519.PublicKey) string {
	sum := sha256.Sum256(key)
	return hex.EncodeToString(sum[:8])
}

// seedTrustedKeys writes trusted_keys with this content and file mode.
func seedTrustedKeys(t *testing.T, host *Host, content string, mode os.FileMode) {
	t.Helper()
	path := host.layout.TrustedKeysPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
}

func readTrustedKeysFile(t *testing.T, host *Host) string {
	t.Helper()
	data, err := os.ReadFile(host.layout.TrustedKeysPath())
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestTrustedKeysListsBase64AndShortFingerprints(t *testing.T) {
	host := newTestHost(t)
	first := testTrustedKey(t, 1)
	second := testTrustedKey(t, 2)
	seedTrustedKeys(t, host, "# console keys\n\n"+trustedKeyText(first)+"  # ada\n"+trustedKeyText(second)+"\n", 0o644)

	keys, err := host.TrustedKeys()
	if err != nil {
		t.Fatalf("TrustedKeys: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("keys = %+v, want two", keys)
	}
	if keys[0].Base64 != trustedKeyText(first) || keys[0].Fingerprint != wantFingerprint(first) {
		t.Fatalf("first key = %+v", keys[0])
	}
	if keys[1].Base64 != trustedKeyText(second) || keys[1].Fingerprint != wantFingerprint(second) {
		t.Fatalf("second key = %+v", keys[1])
	}
	if len(keys[0].Fingerprint) != 16 {
		t.Fatalf("fingerprint %q is not the short 16-character form", keys[0].Fingerprint)
	}
}

func TestTrustedKeysMissingFileIsEmptyList(t *testing.T) {
	host := newTestHost(t)

	keys, err := host.TrustedKeys()
	if err != nil {
		t.Fatalf("TrustedKeys: %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("keys = %+v, want none", keys)
	}
	encoded, err := json.Marshal(keys)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != "[]" {
		t.Fatalf("JSON = %s, want []", encoded)
	}
}

func TestAddTrustedKeyCreatesFile(t *testing.T) {
	host := newTestHost(t)
	key := testTrustedKey(t, 3)

	keys, err := host.AddTrustedKey("\n  " + trustedKeyText(key) + " \n")
	if err != nil {
		t.Fatalf("AddTrustedKey: %v", err)
	}
	if len(keys) != 1 || keys[0].Base64 != trustedKeyText(key) || keys[0].Fingerprint != wantFingerprint(key) {
		t.Fatalf("keys = %+v", keys)
	}
	if content := readTrustedKeysFile(t, host); content != trustedKeyText(key)+"\n" {
		t.Fatalf("trusted_keys = %q", content)
	}
	info, err := os.Stat(host.layout.TrustedKeysPath())
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o644 {
		t.Fatalf("mode = %o, want 644", perm)
	}
}

func TestAddTrustedKeyPreservesCommentsAndOtherKeys(t *testing.T) {
	host := newTestHost(t)
	first := testTrustedKey(t, 4)
	second := testTrustedKey(t, 5)
	seeded := "# console keys\n" + trustedKeyText(first) + "  # ada\n"
	seedTrustedKeys(t, host, seeded, 0o644)

	keys, err := host.AddTrustedKey(trustedKeyText(second))
	if err != nil {
		t.Fatalf("AddTrustedKey: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("keys = %+v, want two", keys)
	}
	if content := readTrustedKeysFile(t, host); content != seeded+trustedKeyText(second)+"\n" {
		t.Fatalf("trusted_keys = %q, want the original lines plus the new key", content)
	}
}

func TestAddTrustedKeyStartsANewLineAfterAFileWithoutFinalNewline(t *testing.T) {
	host := newTestHost(t)
	key := testTrustedKey(t, 6)
	seedTrustedKeys(t, host, "# no final newline", 0o644)

	if _, err := host.AddTrustedKey(trustedKeyText(key)); err != nil {
		t.Fatalf("AddTrustedKey: %v", err)
	}
	want := "# no final newline\n" + trustedKeyText(key) + "\n"
	if content := readTrustedKeysFile(t, host); content != want {
		t.Fatalf("trusted_keys = %q, want %q", content, want)
	}
}

func TestAddTrustedKeyRejectsDuplicates(t *testing.T) {
	host := newTestHost(t)
	key := testTrustedKey(t, 7)
	seeded := "# console keys\n" + trustedKeyText(key) + "\n"
	seedTrustedKeys(t, host, seeded, 0o644)

	_, err := host.AddTrustedKey(" " + trustedKeyText(key) + " ")
	if !errors.Is(err, ErrTrustedKeyDuplicate) {
		t.Fatalf("err = %v, want ErrTrustedKeyDuplicate", err)
	}
	if !strings.Contains(err.Error(), "already trusted") || !strings.Contains(err.Error(), wantFingerprint(key)) {
		t.Fatalf("err = %q, want the duplicate reason and the fingerprint", err)
	}
	if content := readTrustedKeysFile(t, host); content != seeded {
		t.Fatalf("a duplicate key changed trusted_keys: %q", content)
	}
}

func TestAddTrustedKeyRejectsInvalidKeys(t *testing.T) {
	cases := map[string]string{
		"empty":      "",
		"whitespace": "   ",
		"not base64": "not a key!",
		"31 bytes":   base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, ed25519.PublicKeySize-1)),
		"33 bytes":   base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, ed25519.PublicKeySize+1)),
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			host := newTestHost(t)

			_, err := host.AddTrustedKey(input)
			if !errors.Is(err, ErrTrustedKeyInvalid) {
				t.Fatalf("err = %v, want ErrTrustedKeyInvalid", err)
			}
			if !strings.Contains(err.Error(), "32-byte ed25519 public key") {
				t.Fatalf("err = %q, want the key size in the message", err)
			}
			if _, statErr := os.Stat(host.layout.TrustedKeysPath()); !os.IsNotExist(statErr) {
				t.Fatalf("a refused key left trusted_keys behind: %v", statErr)
			}
		})
	}
}

func TestAddTrustedKeyRefusesToRewriteABrokenFile(t *testing.T) {
	host := newTestHost(t)
	seedTrustedKeys(t, host, "this is not a key\n", 0o644)

	_, err := host.AddTrustedKey(trustedKeyText(testTrustedKey(t, 8)))
	if err == nil || !strings.Contains(err.Error(), "line 1 is not an ed25519 public key") {
		t.Fatalf("err = %v, want the broken line in the message", err)
	}
	if content := readTrustedKeysFile(t, host); content != "this is not a key\n" {
		t.Fatalf("a broken file was rewritten: %q", content)
	}
}

func TestRemoveTrustedKeyKeepsCommentsAndOtherKeys(t *testing.T) {
	host := newTestHost(t)
	first := testTrustedKey(t, 9)
	second := testTrustedKey(t, 10)
	seedTrustedKeys(t, host, "# console keys\n"+trustedKeyText(first)+"\n# keep this\n"+trustedKeyText(second)+"\n", 0o644)

	keys, err := host.RemoveTrustedKey(trustedKeyText(first))
	if err != nil {
		t.Fatalf("RemoveTrustedKey: %v", err)
	}
	if len(keys) != 1 || keys[0].Base64 != trustedKeyText(second) || keys[0].Fingerprint != wantFingerprint(second) {
		t.Fatalf("keys = %+v, want only the second key", keys)
	}
	want := "# console keys\n# keep this\n" + trustedKeyText(second) + "\n"
	if content := readTrustedKeysFile(t, host); content != want {
		t.Fatalf("trusted_keys = %q, want %q", content, want)
	}
	info, err := os.Stat(host.layout.TrustedKeysPath())
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o644 {
		t.Fatalf("mode = %o, want 644", perm)
	}
}

func TestRemoveTrustedKeyRemovesTheLastKey(t *testing.T) {
	host := newTestHost(t)
	key := testTrustedKey(t, 11)
	seedTrustedKeys(t, host, "# console keys\n"+trustedKeyText(key)+"\n", 0o644)

	keys, err := host.RemoveTrustedKey(trustedKeyText(key))
	if err != nil {
		t.Fatalf("RemoveTrustedKey: %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("keys = %+v, want none", keys)
	}
	if content := readTrustedKeysFile(t, host); content != "# console keys\n" {
		t.Fatalf("trusted_keys = %q, want the comment", content)
	}
}

func TestRemoveTrustedKeyRejectsUnknownKey(t *testing.T) {
	host := newTestHost(t)
	key := testTrustedKey(t, 12)
	unknown := testTrustedKey(t, 13)
	seeded := trustedKeyText(key) + "\n"
	seedTrustedKeys(t, host, seeded, 0o644)

	_, err := host.RemoveTrustedKey(trustedKeyText(unknown))
	if !errors.Is(err, ErrTrustedKeyNotFound) {
		t.Fatalf("err = %v, want ErrTrustedKeyNotFound", err)
	}
	if !strings.Contains(err.Error(), "not trusted") || !strings.Contains(err.Error(), wantFingerprint(unknown)) {
		t.Fatalf("err = %q, want the reason and the fingerprint", err)
	}
	if content := readTrustedKeysFile(t, host); content != seeded {
		t.Fatalf("removing an unknown key changed trusted_keys: %q", content)
	}

	emptyHost := newTestHost(t)
	if _, err := emptyHost.RemoveTrustedKey(trustedKeyText(unknown)); !errors.Is(err, ErrTrustedKeyNotFound) {
		t.Fatalf("missing file: err = %v, want ErrTrustedKeyNotFound", err)
	}
}

func TestRemoveTrustedKeyRejectsInvalidKey(t *testing.T) {
	host := newTestHost(t)
	seedTrustedKeys(t, host, trustedKeyText(testTrustedKey(t, 14))+"\n", 0o644)

	if _, err := host.RemoveTrustedKey("not base64"); !errors.Is(err, ErrTrustedKeyInvalid) {
		t.Fatalf("err = %v, want ErrTrustedKeyInvalid", err)
	}
}

func TestTrustedKeyRewriteUses0644AndLeavesNoTempFile(t *testing.T) {
	host := newTestHost(t)
	first := testTrustedKey(t, 15)
	second := testTrustedKey(t, 16)
	seedTrustedKeys(t, host, trustedKeyText(first)+"\n", 0o600)

	if _, err := host.AddTrustedKey(trustedKeyText(second)); err != nil {
		t.Fatalf("AddTrustedKey: %v", err)
	}
	if _, err := host.RemoveTrustedKey(trustedKeyText(first)); err != nil {
		t.Fatalf("RemoveTrustedKey: %v", err)
	}
	info, err := os.Stat(host.layout.TrustedKeysPath())
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o644 {
		t.Fatalf("mode = %o, want 644", perm)
	}
	entries, err := os.ReadDir(filepath.Dir(host.layout.TrustedKeysPath()))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), trustedKeysFile+".tmp-") {
			t.Fatalf("left a temp file behind: %s", entry.Name())
		}
	}
}

func TestTrustedKeyWritesDoNotShareAListBetweenCalls(t *testing.T) {
	host := newTestHost(t)
	first := testTrustedKey(t, 17)
	second := testTrustedKey(t, 18)

	added, err := host.AddTrustedKey(trustedKeyText(first))
	if err != nil {
		t.Fatal(err)
	}
	added[0].Base64 = "changed by the caller"
	if content := readTrustedKeysFile(t, host); content != trustedKeyText(first)+"\n" {
		t.Fatalf("caller mutation changed trusted_keys: %q", content)
	}

	if _, err := host.AddTrustedKey(trustedKeyText(second)); err != nil {
		t.Fatal(err)
	}
	listed, err := host.TrustedKeys()
	if err != nil {
		t.Fatal(err)
	}
	listed[0].Fingerprint = "changed by the caller"
	fresh, err := host.TrustedKeys()
	if err != nil {
		t.Fatal(err)
	}
	if fresh[0].Fingerprint != wantFingerprint(first) {
		t.Fatalf("caller mutation leaked into a later list: %+v", fresh[0])
	}
}

func TestConcurrentAddTrustedKeysKeepsEveryKey(t *testing.T) {
	host := newTestHost(t)
	const count = 8
	keys := make([]ed25519.PublicKey, count)
	for i := range keys {
		keys[i] = testTrustedKey(t, byte(20+i))
	}

	var wg sync.WaitGroup
	errs := make(chan error, count)
	for _, key := range keys {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := host.AddTrustedKey(trustedKeyText(key)); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("AddTrustedKey: %v", err)
	}

	listed, err := host.TrustedKeys()
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != count {
		t.Fatalf("keys = %d, want %d: one add overwrote another", len(listed), count)
	}
}
