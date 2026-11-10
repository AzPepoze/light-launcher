package plugins

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const trustedKeysFile = "trusted_keys"

// loadTrustedKeys reads ed25519 public keys, one base64 key per line. "#" starts a comment.
// A missing file means no keys are trusted.
func loadTrustedKeys(path string) ([]ed25519.PublicKey, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
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
