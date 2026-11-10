package downloads

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// writeTree creates empty files at the given relative paths under a new temp folder.
func writeTree(t *testing.T, files map[string]int) string {
	t.Helper()
	root := t.TempDir()
	for rel, size := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, make([]byte, size), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// launcherExt is the native launcher extension for the test platform.
func launcherExt(t *testing.T) string {
	t.Helper()
	switch runtime.GOOS {
	case "linux":
		return ".x86_64"
	case "windows":
		return ".exe"
	}
	t.Skip("launcher names are tested on linux and windows")
	return ""
}

func TestDetectExecutablePrefersShallowAndSkipsUninstallers(t *testing.T) {
	ext := launcherExt(t)
	root := writeTree(t, map[string]int{
		"game" + ext:                100,
		"tools/big-tool" + ext:      5000,
		"unins000" + ext:            9000,
		"CrashReporter/crash" + ext: 9000,
	})
	got, err := detectExecutable(root)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(root, "game"+ext); got != want {
		t.Fatalf("launcher = %s, want %s", got, want)
	}
}

func TestDetectExecutableBreaksTiesBySize(t *testing.T) {
	ext := launcherExt(t)
	root := writeTree(t, map[string]int{
		"bin/small" + ext: 10,
		"bin/large" + ext: 100,
	})
	got, _ := detectExecutable(root)
	if want := filepath.Join(root, "bin", "large"+ext); got != want {
		t.Fatalf("launcher = %s, want %s", got, want)
	}
}

func TestDetectExecutableIsEmptyWhenNothingFits(t *testing.T) {
	root := writeTree(t, map[string]int{"readme.txt": 1, "unins000.exe": 1})
	got, err := detectExecutable(root)
	if err != nil || got != "" {
		t.Fatalf("got %q, %v; want no launcher", got, err)
	}
}

func TestNativeLauncherBeatsWindowsBinaryOnLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux only")
	}
	root := writeTree(t, map[string]int{"Game.exe": 10, "Game.sh": 10})
	got, _ := detectExecutable(root)
	if want := filepath.Join(root, "Game.sh"); got != want {
		t.Fatalf("launcher = %s, want %s", got, want)
	}
}

func TestTreeSizeAndCleanIgnoreJunk(t *testing.T) {
	root := writeTree(t, map[string]int{
		"__MACOSX/x": 5,
		".DS_Store":  5,
		"Thumbs.db":  5,
		"real.bin":   42,
	})
	files, err := listFiles(root)
	if err != nil || len(files) != 1 || files[0] != "real.bin" {
		t.Fatalf("listFiles = %v, %v; want only real.bin", files, err)
	}
	if got := treeSize(root); got != 42 {
		t.Fatalf("treeSize = %d, want 42", got)
	}
	if err := cleanTree(root); err != nil {
		t.Fatal(err)
	}
	for _, junk := range []string{"__MACOSX", ".DS_Store", "Thumbs.db"} {
		if _, err := os.Stat(filepath.Join(root, junk)); !os.IsNotExist(err) {
			t.Fatalf("%s should be removed", junk)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "real.bin")); err != nil {
		t.Fatal("real file must stay")
	}
}

func TestEntryPathRefusesEscapesAndSkipsJunk(t *testing.T) {
	cases := map[string]struct {
		rel  string
		skip bool
		fail bool
	}{
		"game/a.txt":      {rel: "game/a.txt"},
		`game\b.txt`:      {rel: "game/b.txt"},
		"__MACOSX/._x":    {skip: true},
		"game/.DS_Store":  {skip: true},
		"../evil":         {fail: true},
		"game/../../evil": {fail: true},
		"C:/windows/evil": {fail: true},
	}
	for name, want := range cases {
		rel, skip, err := entryPath(name)
		if (err != nil) != want.fail {
			t.Errorf("%q: err = %v, want failure %v", name, err, want.fail)
			continue
		}
		if skip != want.skip || rel != want.rel {
			t.Errorf("%q: rel=%q skip=%v, want rel=%q skip=%v", name, rel, skip, want.rel, want.skip)
		}
	}
}
