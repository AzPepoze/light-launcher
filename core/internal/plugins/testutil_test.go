package plugins

import (
	"archive/zip"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type zipEntry struct {
	name string
	data []byte
	mode fs.FileMode
}

// writeZipEntries writes a zip file with exactly these entries, in this order.
func writeZipEntries(t *testing.T, path string, entries []zipEntry) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	writer := zip.NewWriter(file)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Deflate}
		if entry.mode != 0 {
			header.SetMode(entry.mode)
		}
		out, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := out.Write(entry.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
}

// writePackageFile writes a package zip from a map of files, sorted by name.
func writePackageFile(t *testing.T, path string, files map[string][]byte) {
	t.Helper()
	var entries []zipEntry
	for _, name := range sortedNames(files) {
		entries = append(entries, zipEntry{name: name, data: files[name]})
	}
	writeZipEntries(t, path, entries)
}

func sortedNames(files map[string][]byte) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// testManifest returns a valid plugin.json for the running platform.
func testManifest(id, version string) string {
	key := PlatformKey()
	return fmt.Sprintf(`{
  "id": %q,
  "name": "Test %s",
  "version": %q,
  "apiVersion": "1",
  "minHost": "2.0.0",
  "entry": { %q: "bin/%s/%s" },
  "contributes": { "sources": [{ "id": %q, "label": "Test", "capabilities": ["browse"] }] },
  "permissions": { "network": ["127.0.0.1"], "fs": [], "secrets": false, "browser": false, "exec": false }
}`, id, id, version, key, key, id, id)
}

// testPackageFiles returns the files of a package without SHA256SUMS or SIGNATURE.
func testPackageFiles(id, version string) map[string][]byte {
	key := PlatformKey()
	return map[string][]byte{
		"plugin.json":           []byte(testManifest(id, version)),
		"bin/" + key + "/" + id: []byte("#!/bin/sh\necho test\n"),
	}
}

// withChecksums adds a SHA256SUMS file listing every file in files.
func withChecksums(files map[string][]byte) map[string][]byte {
	var lines strings.Builder
	for _, name := range sortedNames(files) {
		sum := sha256.Sum256(files[name])
		fmt.Fprintf(&lines, "%s  %s\n", hex.EncodeToString(sum[:]), name)
	}
	out := map[string][]byte{}
	for name, data := range files {
		out[name] = data
	}
	out[checksumsFile] = []byte(lines.String())
	return out
}

// withSignature signs the SHA256SUMS file with key.
func withSignature(files map[string][]byte, key ed25519.PrivateKey) map[string][]byte {
	signature := ed25519.Sign(key, files[checksumsFile])
	files[signatureFile] = []byte(base64.StdEncoding.EncodeToString(signature) + "\n")
	return files
}

// newTestInstaller returns an installer whose layout lives in a temp folder.
func newTestInstaller(t *testing.T) (*installer, Layout) {
	t.Helper()
	base := t.TempDir()
	layout := Layout{Root: filepath.Join(base, "plugins"), Data: filepath.Join(base, "plugin-data")}
	in, err := newInstaller(layout, "2.0.0")
	if err != nil {
		t.Fatal(err)
	}
	return in, layout
}

// writeTrustedKey writes one public key into the layout's trusted_keys file.
func writeTrustedKey(t *testing.T, layout Layout, key ed25519.PublicKey) {
	t.Helper()
	if err := os.MkdirAll(layout.Root, 0o755); err != nil {
		t.Fatal(err)
	}
	line := base64.StdEncoding.EncodeToString(key) + "\n"
	if err := os.WriteFile(layout.TrustedKeysPath(), []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
}
