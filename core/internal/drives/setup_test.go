package drives

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func newPrefix(t *testing.T) string {
	t.Helper()
	prefix := t.TempDir()
	if err := os.MkdirAll(filepath.Join(prefix, "dosdevices"), 0755); err != nil {
		t.Fatal(err)
	}
	return prefix
}

func TestSaveAndLoadSetupRoundTrip(t *testing.T) {
	file := filepath.Join(t.TempDir(), "drive-setup.json")
	folder := t.TempDir()

	if err := SaveSetup(file, []SetupEntry{{Letter: "D", Path: folder}}); err != nil {
		t.Fatalf("SaveSetup = %v", err)
	}
	entries, err := LoadSetup(file)
	if err != nil || len(entries) != 1 || entries[0].Letter != "d:" {
		t.Fatalf("LoadSetup = %+v, %v; want one d: entry", entries, err)
	}
}

func TestLoadSetupMissingFileIsEmpty(t *testing.T) {
	entries, err := LoadSetup(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil || entries == nil || len(entries) != 0 {
		t.Errorf("LoadSetup = %#v, %v; want an empty slice and no error", entries, err)
	}
}

func TestSaveSetupRejectsBadEntries(t *testing.T) {
	file := filepath.Join(t.TempDir(), "drive-setup.json")
	folder := t.TempDir()

	if err := SaveSetup(file, []SetupEntry{{Letter: "z:", Path: folder}}); !errors.Is(err, ErrLockedLetter) {
		t.Errorf("z: err = %v, want ErrLockedLetter", err)
	}
	if err := SaveSetup(file, []SetupEntry{{Letter: "a:", Path: folder}}); !errors.Is(err, ErrInvalidLetter) {
		t.Errorf("a: err = %v, want ErrInvalidLetter", err)
	}
	if err := SaveSetup(file, []SetupEntry{{Letter: "d:", Path: folder}, {Letter: "D", Path: folder}}); !errors.Is(err, ErrLetterTaken) {
		t.Errorf("duplicate err = %v, want ErrLetterTaken", err)
	}
	if err := SaveSetup(file, []SetupEntry{{Letter: "d:", Path: filepath.Join(folder, "missing")}}); !errors.Is(err, ErrNotFolder) {
		t.Errorf("missing folder err = %v, want ErrNotFolder", err)
	}
}

func TestApplySetupLinksEveryPrefixAndKeepsFolder(t *testing.T) {
	prefix := newPrefix(t)
	folder := t.TempDir()

	if err := ApplySetup(prefix, []SetupEntry{{Letter: "d:", Path: folder}}); err != nil {
		t.Fatalf("ApplySetup = %v", err)
	}
	list, err := ListBindings(prefix)
	if err != nil || len(list) != 1 || list[0].Letter != "d:" || list[0].Kind != "folder" || !list[0].Exists {
		t.Fatalf("ListBindings = %+v, %v; want one d: folder", list, err)
	}
	if _, err := os.Stat(folder); err != nil {
		t.Errorf("the folder itself was touched: %v", err)
	}
}

func TestApplySetupRemovesDroppedLetters(t *testing.T) {
	prefix := newPrefix(t)
	folder := t.TempDir()
	if err := ApplySetup(prefix, []SetupEntry{{Letter: "e:", Path: folder}}); err != nil {
		t.Fatal(err)
	}

	if err := ApplySetup(prefix, nil); err != nil {
		t.Fatalf("second ApplySetup = %v", err)
	}
	list, _ := ListBindings(prefix)
	if len(list) != 0 {
		t.Errorf("drives after removing from setup = %+v, want none", list)
	}
}

func TestApplySetupReplacesCAndRestoresWineDefault(t *testing.T) {
	prefix := newPrefix(t)
	if err := os.MkdirAll(filepath.Join(prefix, "drive_c"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(wineSystemTarget, filepath.Join(prefix, "dosdevices", "c:")); err != nil {
		t.Fatal(err)
	}
	disk := t.TempDir()

	if err := ApplySetup(prefix, []SetupEntry{{Letter: "c:", Path: disk}}); err != nil {
		t.Fatalf("ApplySetup with C: = %v", err)
	}
	list, _ := ListBindings(prefix)
	if len(list) != 1 || list[0].Letter != "c:" || list[0].Target != disk {
		t.Fatalf("C: after setup = %+v, want C: to the chosen disk", list)
	}

	if err := ApplySetup(prefix, nil); err != nil {
		t.Fatalf("ApplySetup without C: = %v", err)
	}
	list, _ = ListBindings(prefix)
	if len(list) != 1 || list[0].Kind != "system" || list[0].Target != wineSystemTarget {
		t.Fatalf("C: after removing it from setup = %+v, want Wine's default system drive", list)
	}
}

func TestApplySetupLeavesRealCFolderAlone(t *testing.T) {
	prefix := newPrefix(t)
	realC := filepath.Join(prefix, "dosdevices", "c:")
	if err := os.MkdirAll(realC, 0755); err != nil {
		t.Fatal(err)
	}

	if err := ApplySetup(prefix, []SetupEntry{{Letter: "c:", Path: t.TempDir()}}); err == nil {
		t.Fatal("ApplySetup replaced a real folder named c:")
	}
	if info, err := os.Stat(realC); err != nil || !info.IsDir() {
		t.Errorf("the real c: folder was changed: %v", err)
	}
}
