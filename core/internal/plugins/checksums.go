package plugins

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	checksumsFile = "SHA256SUMS"
	signatureFile = "SIGNATURE"
)

// verifyChecksums checks every file in the extracted package against SHA256SUMS, and every
// SHA256SUMS line against a file. It returns the raw SHA256SUMS bytes for signature checks.
func verifyChecksums(root string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(root, checksumsFile))
	if err != nil {
		return nil, fmt.Errorf("package has no %s file", checksumsFile)
	}
	expected, err := parseChecksums(data)
	if err != nil {
		return nil, err
	}
	files, err := packageFiles(root)
	if err != nil {
		return nil, err
	}
	for _, name := range files {
		want, listed := expected[name]
		if !listed {
			return nil, fmt.Errorf("%s is not listed in %s", name, checksumsFile)
		}
		got, err := fileSHA256(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return nil, err
		}
		if got != want {
			return nil, fmt.Errorf("checksum mismatch for %s", name)
		}
		delete(expected, name)
	}
	for name := range expected {
		return nil, fmt.Errorf("%s is listed in %s but missing from the package", name, checksumsFile)
	}
	return data, nil
}

// parseChecksums reads sha256sum-style lines: "<64 hex digits>  <relative path>".
func parseChecksums(data []byte) (map[string]string, error) {
	sums := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		digest, name, ok := strings.Cut(line, "  ")
		if !ok || len(digest) != sha256.Size*2 {
			return nil, fmt.Errorf("bad line in %s: %q", checksumsFile, line)
		}
		if _, err := hex.DecodeString(digest); err != nil {
			return nil, fmt.Errorf("bad digest in %s for %q", checksumsFile, name)
		}
		if _, err := safeEntryName(name); err != nil {
			return nil, err
		}
		if _, dup := sums[name]; dup {
			return nil, fmt.Errorf("%s lists %q twice", checksumsFile, name)
		}
		sums[name] = digest
	}
	return sums, nil
}

// packageFiles lists the slash paths of every file except SHA256SUMS and SIGNATURE.
func packageFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(full string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, full)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		if name == checksumsFile || name == signatureFile {
			return nil
		}
		files = append(files, name)
		return nil
	})
	return files, err
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// checkPackageChecksum compares a package file to the digest a release published. An empty digest skips the check.
func checkPackageChecksum(path, expected string) error {
	if expected == "" {
		return nil
	}
	expected = strings.ToLower(strings.TrimSpace(expected))
	if len(expected) != sha256.Size*2 {
		return fmt.Errorf("release checksum %q is not a SHA256 digest", expected)
	}
	if _, err := hex.DecodeString(expected); err != nil {
		return fmt.Errorf("release checksum %q is not a SHA256 digest", expected)
	}
	got, err := fileSHA256(path)
	if err != nil {
		return fmt.Errorf("checksum package: %w", err)
	}
	if got != expected {
		return errors.New("package does not match the release SHA256SUMS")
	}
	return nil
}
