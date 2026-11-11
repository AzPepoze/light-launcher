package uploads

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenClearsEarlierUploads(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "uploads")
	if err := os.MkdirAll(filepath.Join(dir, "upload-old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(dir); err != nil {
		t.Fatal(err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatalf("Open kept %d old uploads", len(entries))
	}
}

func TestRemoveDeletesOnlyTheUploadFolder(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "uploads"))
	if err != nil {
		t.Fatal(err)
	}
	folder, err := store.NewFolder()
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(folder, "myplugin", "plugin.json")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := store.Remove(filepath.Join(folder, "myplugin")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(folder); !os.IsNotExist(err) {
		t.Fatalf("upload folder still exists: %v", err)
	}
	if _, err := os.Stat(store.dir); err != nil {
		t.Fatalf("Remove deleted the store itself: %v", err)
	}
}

func TestRemoveLeavesPathsOutsideTheStore(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "uploads"))
	if err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "demo.llplugin")
	if err := os.WriteFile(outside, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{outside, store.dir, "https://example.com/demo.llplugin"} {
		if err := store.Remove(path); err != nil {
			t.Fatalf("Remove(%q): %v", path, err)
		}
	}
	if _, err := os.Stat(outside); err != nil {
		t.Fatalf("Remove deleted a file outside the store: %v", err)
	}
	if _, err := os.Stat(store.dir); err != nil {
		t.Fatalf("Remove deleted the store folder: %v", err)
	}
}
