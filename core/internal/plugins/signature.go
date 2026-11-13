package plugins

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const trustedKeysFile = "trusted_keys"

// Errors returned when the trusted key list cannot be changed. Use errors.Is on them; the
// messages add the bad input or the key fingerprint.
var (
	ErrTrustedKeyInvalid   = errors.New("trusted key is not a 32-byte ed25519 public key")
	ErrTrustedKeyDuplicate = errors.New("trusted key is already trusted")
	ErrTrustedKeyNotFound  = errors.New("trusted key is not trusted")
)

// trustedKeysMu serializes read-modify-write cycles on trusted_keys. The file itself is replaced
// by rename, so readers always see a complete list even without the lock.
var trustedKeysMu sync.Mutex

// TrustedKey is one ed25519 public key that plugin signatures are verified against.
// Base64 is the standard base64 key, as written in trusted_keys. Fingerprint is the first
// 8 bytes of the key's SHA256, hex encoded, so keys can be compared without showing them.
type TrustedKey struct {
	Base64      string `json:"base64"`
	Fingerprint string `json:"fingerprint"`
}

// loadTrustedKeys reads ed25519 public keys, one base64 key per line. "#" starts a comment.
// A missing file means no keys are trusted.
func loadTrustedKeys(path string) ([]ed25519.PublicKey, error) {
	data, err := readTrustedKeysData(path)
	if err != nil {
		return nil, err
	}
	return parseTrustedKeysData(data)
}

// parseTrustedKeysData parses the trusted_keys contents. Comments and blank lines are skipped.
func parseTrustedKeysData(data []byte) ([]ed25519.PublicKey, error) {
	var keys []ed25519.PublicKey
	for i, line := range strings.Split(string(data), "\n") {
		line, _, _ = strings.Cut(line, "#")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, err := base64.StdEncoding.DecodeString(line)
		if err != nil || len(key) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("%s line %d is not an ed25519 public key", trustedKeysFile, i+1)
		}
		keys = append(keys, ed25519.PublicKey(key))
	}
	return keys, nil
}

// TrustedKeys lists the keys that signatures are verified against, in file order.
func (h *Host) TrustedKeys() ([]TrustedKey, error) {
	trustedKeysMu.Lock()
	defer trustedKeysMu.Unlock()
	keys, err := loadTrustedKeys(h.layout.TrustedKeysPath())
	if err != nil {
		return nil, err
	}
	return trustedKeyViews(keys), nil
}

// AddTrustedKey appends a base64 ed25519 public key to trusted_keys. Comments and every other
// line are kept as they are. The file is created when it is missing. The refreshed list is returned.
func (h *Host) AddTrustedKey(text string) ([]TrustedKey, error) {
	key, err := parseTrustedKeyText(text)
	if err != nil {
		return nil, err
	}
	trustedKeysMu.Lock()
	defer trustedKeysMu.Unlock()
	path := h.layout.TrustedKeysPath()
	data, err := readTrustedKeysData(path)
	if err != nil {
		return nil, err
	}
	keys, err := parseTrustedKeysData(data)
	if err != nil {
		return nil, err
	}
	for _, existing := range keys {
		if bytes.Equal(existing, key) {
			return nil, fmt.Errorf("%w: %s", ErrTrustedKeyDuplicate, trustedKeyFingerprint(key))
		}
	}
	if err := writeTrustedKeysData(path, appendTrustedKeyLine(data, key)); err != nil {
		return nil, err
	}
	return trustedKeyViews(append(keys, key)), nil
}

// RemoveTrustedKey deletes a base64 ed25519 public key from trusted_keys. Comments and every
// other line are kept as they are. Removing a key that is not trusted is an error.
// The refreshed list is returned.
func (h *Host) RemoveTrustedKey(text string) ([]TrustedKey, error) {
	key, err := parseTrustedKeyText(text)
	if err != nil {
		return nil, err
	}
	trustedKeysMu.Lock()
	defer trustedKeysMu.Unlock()
	path := h.layout.TrustedKeysPath()
	data, err := readTrustedKeysData(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	kept := make([]string, 0, len(lines))
	var remaining []ed25519.PublicKey
	found := false
	for i, line := range lines {
		body, _, _ := strings.Cut(line, "#")
		body = strings.TrimSpace(body)
		if body == "" {
			kept = append(kept, line)
			continue
		}
		decoded, err := base64.StdEncoding.DecodeString(body)
		if err != nil || len(decoded) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("%s line %d is not an ed25519 public key", trustedKeysFile, i+1)
		}
		if bytes.Equal(decoded, key) {
			found = true
			continue
		}
		kept = append(kept, line)
		remaining = append(remaining, ed25519.PublicKey(decoded))
	}
	if !found {
		return nil, fmt.Errorf("%w: %s", ErrTrustedKeyNotFound, trustedKeyFingerprint(key))
	}
	if err := writeTrustedKeysData(path, withFinalNewline([]byte(strings.Join(kept, "\n")))); err != nil {
		return nil, err
	}
	return trustedKeyViews(remaining), nil
}

// verifySignature reports whether the package is signed by a trusted key. A package without
// a SIGNATURE file is unsigned, not an error. A SIGNATURE that no trusted key accepts is an error.
// The signature is a base64 ed25519 signature over the SHA256SUMS bytes.
func verifySignature(root string, checksums []byte, keys []ed25519.PublicKey) (bool, error) {
	data, err := os.ReadFile(filepath.Join(root, signatureFile))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	signature, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(data)))
	if err != nil || len(signature) != ed25519.SignatureSize {
		return false, fmt.Errorf("%s is not a valid ed25519 signature", signatureFile)
	}
	for _, key := range keys {
		if ed25519.Verify(key, checksums, signature) {
			return true, nil
		}
	}
	return false, fmt.Errorf("%s does not match any trusted key", signatureFile)
}

// parseTrustedKeyText turns one pasted key into an ed25519 public key. Surrounding whitespace
// is ignored so a key copied from a file or terminal still works.
func parseTrustedKeyText(text string) (ed25519.PublicKey, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("%w: the key is empty", ErrTrustedKeyInvalid)
	}
	decoded, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return nil, fmt.Errorf("%w: it is not base64", ErrTrustedKeyInvalid)
	}
	if len(decoded) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("%w: it is %d bytes", ErrTrustedKeyInvalid, len(decoded))
	}
	return ed25519.PublicKey(decoded), nil
}

// trustedKeyView builds the UI view of one key.
func trustedKeyView(key ed25519.PublicKey) TrustedKey {
	return TrustedKey{
		Base64:      base64.StdEncoding.EncodeToString(key),
		Fingerprint: trustedKeyFingerprint(key),
	}
}

// trustedKeyViews builds the UI view of a key list, never nil so JSON sends [].
func trustedKeyViews(keys []ed25519.PublicKey) []TrustedKey {
	out := make([]TrustedKey, 0, len(keys))
	for _, key := range keys {
		out = append(out, trustedKeyView(key))
	}
	return out
}

// trustedKeyFingerprint is the first 8 bytes of the key's SHA256, hex encoded.
func trustedKeyFingerprint(key ed25519.PublicKey) string {
	sum := sha256.Sum256(key)
	return hex.EncodeToString(sum[:8])
}

// readTrustedKeysData reads trusted_keys. A missing file reads as no keys.
func readTrustedKeysData(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return data, err
}

// appendTrustedKeyLine adds one key line after the existing contents. A file that does not end
// with a newline gets one first, so the new key never joins the last comment or key.
func appendTrustedKeyLine(data []byte, key ed25519.PublicKey) []byte {
	line := base64.StdEncoding.EncodeToString(key)
	out := make([]byte, 0, len(data)+len(line)+2)
	out = append(out, data...)
	if len(out) > 0 && out[len(out)-1] != '\n' {
		out = append(out, '\n')
	}
	out = append(out, line...)
	return append(out, '\n')
}

// withFinalNewline keeps a rewritten key file ending in a newline, as a hand-edited one does.
func withFinalNewline(data []byte) []byte {
	if len(data) == 0 || data[len(data)-1] == '\n' {
		return data
	}
	return append(data, '\n')
}

// writeTrustedKeysData replaces trusted_keys atomically: write a temp file next to it, then
// rename over the old file. The mode is 0644 so the CLI and the app can both read and write it.
func writeTrustedKeysData(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, trustedKeysFile+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
