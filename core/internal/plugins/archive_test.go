package plugins

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractZipUnpacksRegularFiles(t *testing.T) {
	base := t.TempDir()
	archive := filepath.Join(base, "ok.zip")
	writeZipEntries(t, archive, []zipEntry{
		{name: "dir/", data: nil, mode: fs.ModeDir | 0o755},
		{name: "dir/file.txt", data: []byte("hello")},
	})
	dest := filepath.Join(base, "out")
	if err := extractZip(archive, dest); err != nil {
		t.Fatalf("extractZip: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dest, "dir", "file.txt"))
	if err != nil || string(data) != "hello" {
		t.Fatalf("file = %q, %v", data, err)
	}
}

func TestExtractZipRejectsUnsafePaths(t *testing.T) {
	cases := map[string]string{
		"parent":    "../evil.txt",
		"nested":    "a/../../evil.txt",
		"absolute":  "/etc/evil.txt",
		"drive":     "C:evil.txt",
		"backslash": "a\\..\\evil.txt",
	}
	for name, entry := range cases {
		t.Run(name, func(t *testing.T) {
			base := t.TempDir()
			archive := filepath.Join(base, "bad.zip")
			writeZipEntries(t, archive, []zipEntry{{name: entry, data: []byte("x")}})
			if err := extractZip(archive, filepath.Join(base, "out")); err == nil {
				t.Fatalf("entry %q was accepted", entry)
			}
			if _, err := os.Stat(filepath.Join(base, "evil.txt")); err == nil {
				t.Fatalf("entry %q escaped the destination", entry)
			}
		})
	}
}

func TestExtractZipRejectsLinksAndDuplicates(t *testing.T) {
	base := t.TempDir()

	link := filepath.Join(base, "link.zip")
	writeZipEntries(t, link, []zipEntry{{name: "run", data: []byte("/etc/passwd"), mode: fs.ModeSymlink | 0o777}})
	if err := extractZip(link, filepath.Join(base, "out-link")); err == nil {
		t.Fatalf("a symlink entry was accepted")
	}

	duplicate := filepath.Join(base, "dup.zip")
	writeZipEntries(t, duplicate, []zipEntry{{name: "a.txt", data: []byte("1")}, {name: "a.txt", data: []byte("2")}})
	if err := extractZip(duplicate, filepath.Join(base, "out-dup")); err == nil {
		t.Fatalf("a duplicate entry was accepted")
	}
}

func TestSafeEntryNameCleansSlashPaths(t *testing.T) {
	name, err := safeEntryName("bin//app/")
	if err != nil || name != "bin/app" {
		t.Fatalf("safeEntryName = %q, %v", name, err)
	}
}
