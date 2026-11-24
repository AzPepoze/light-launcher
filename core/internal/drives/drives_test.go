package drives

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleMountInfo = `22 1 259:2 / / rw - ext4 /dev/root rw
36 22 259:3 / /mnt/6AB2DBF3B2DBC1AD rw - ntfs3 /dev/nvme0n1p3 rw
40 22 259:4 / /mnt/My\040Games rw - ext4 /dev/sdb1 rw
`

func TestMountPointForPicksLongestMatch(t *testing.T) {
	cases := map[string]string{
		"/mnt/6AB2DBF3B2DBC1AD/Games/OpenIV/OpenIV.exe": "/mnt/6AB2DBF3B2DBC1AD",
		"/home/user/game.exe":                           "/",
		"/mnt/My Games/game.exe":                        "/mnt/My Games",
	}
	for path, want := range cases {
		if got := mountPointFor(sampleMountInfo, path); got != want {
			t.Errorf("mountPointFor(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestMountPointForDoesNotMatchSiblingPrefix(t *testing.T) {
	if got := mountPointFor(sampleMountInfo, "/mnt/6AB2DBF3B2DBC1ADX/game.exe"); got != "/" {
		t.Errorf("sibling folder matched mount %q, want \"/\"", got)
	}
}

func TestFreeLetterSkipsUsedLetters(t *testing.T) {
	letter, ok := freeLetter(map[byte]bool{'d': true, 'e': true})
	if !ok || letter != 'f' {
		t.Errorf("freeLetter = %q, %v; want 'f', true", letter, ok)
	}
}

func TestLinkMountCreatesThenReusesLink(t *testing.T) {
	dosdevices := filepath.Join(t.TempDir(), "dosdevices")
	mount := t.TempDir()

	first, err := linkMount(dosdevices, mount)
	if err != nil || first != "d:" {
		t.Fatalf("first linkMount = %q, %v; want \"d:\", nil", first, err)
	}
	target, err := os.Readlink(filepath.Join(dosdevices, "d:"))
	if err != nil || target != mount {
		t.Fatalf("link target = %q, %v; want %q", target, err, mount)
	}

	again, err := linkMount(dosdevices, mount)
	if err != nil || again != "d:" {
		t.Fatalf("second linkMount = %q, %v; want the same \"d:\"", again, err)
	}
}

func TestLinkMountKeepsManualMapping(t *testing.T) {
	dosdevices := filepath.Join(t.TempDir(), "dosdevices")
	if err := os.MkdirAll(dosdevices, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(dosdevices, "d:")); err != nil {
		t.Fatal(err)
	}

	name, err := linkMount(dosdevices, t.TempDir())
	if err != nil || name != "e:" {
		t.Fatalf("linkMount = %q, %v; want \"e:\" because d: is taken", name, err)
	}
}
