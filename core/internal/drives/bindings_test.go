package drives

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeLetter(t *testing.T) {
	cases := map[string]string{"D": "d:", " e: ": "e:", "c": "c:", "Y:": "y:"}
	for input, want := range cases {
		got, err := normalizeLetter(input)
		if err != nil || got != want {
			t.Errorf("normalizeLetter(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
	if _, err := normalizeLetter("z"); !errors.Is(err, ErrLockedLetter) {
		t.Errorf("z err = %v, want ErrLockedLetter", err)
	}
	if _, err := normalizeLetter("1:"); !errors.Is(err, ErrInvalidLetter) {
		t.Errorf("1: err = %v, want ErrInvalidLetter", err)
	}
}

func TestListBindingsShowsWineSystemDrive(t *testing.T) {
	prefix := newPrefix(t)
	if err := os.MkdirAll(filepath.Join(prefix, "drive_c"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(wineSystemTarget, filepath.Join(prefix, "dosdevices", "c:")); err != nil {
		t.Fatal(err)
	}

	list, err := ListBindings(prefix)
	if err != nil || len(list) != 1 {
		t.Fatalf("ListBindings = %+v, %v", list, err)
	}
	if system := list[0]; system.Letter != "c:" || system.Kind != "system" || !system.Exists {
		t.Errorf("system drive = %+v", system)
	}
}

func TestListBindingsWithoutPrefixIsEmpty(t *testing.T) {
	list, err := ListBindings(filepath.Join(t.TempDir(), "missing"))
	if err != nil || list == nil || len(list) != 0 {
		t.Errorf("ListBindings = %#v, %v; want an empty slice and no error", list, err)
	}
}
