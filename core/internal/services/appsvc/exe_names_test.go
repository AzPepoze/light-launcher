package appsvc

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCachedVersionStringsReadsOnlyWhenFileChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Launcher.exe")
	firstTime := time.Unix(1_700_000_000, 0)
	writeTestExe(t, path, "First Name", "")
	setModTime(t, path, firstTime)
	if got := cachedVersionStrings(statScanned(t, path)).productName; got != "First Name" {
		t.Fatalf("first read = %q, want %q", got, "First Name")
	}

	// Same size and modification time: the cached value is kept, even though the bytes changed.
	writeTestExe(t, path, "Other Name", "")
	setModTime(t, path, firstTime)
	if got := cachedVersionStrings(statScanned(t, path)).productName; got != "First Name" {
		t.Fatalf("unchanged file = %q, want the cached %q", got, "First Name")
	}

	// A new modification time means the file is read again.
	setModTime(t, path, firstTime.Add(time.Hour))
	if got := cachedVersionStrings(statScanned(t, path)).productName; got != "Other Name" {
		t.Fatalf("changed file = %q, want %q", got, "Other Name")
	}
}

func setModTime(t *testing.T, path string, when time.Time) {
	t.Helper()
	if err := os.Chtimes(path, when, when); err != nil {
		t.Fatal(err)
	}
}

// statScanned returns the scanned file for path, as a scan would record it.
func statScanned(t *testing.T, path string) scannedFile {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return scannedFile{Path: path, Size: info.Size(), ModTime: info.ModTime()}
}
