package appsvc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompileWildcardPatternsMatchesGlobs(t *testing.T) {
	regexes := compileWildcardPatterns([]string{"*.exe", "unins*"})
	if len(regexes) != 2 {
		t.Fatalf("got %d regexes, want 2", len(regexes))
	}
	if !regexes[0].MatchString("game.EXE") || regexes[0].MatchString("game.txt") {
		t.Error("*.exe must match any case of .exe and nothing else")
	}
	if !regexes[1].MatchString("unins000.exe") {
		t.Error("unins* must match unins000.exe")
	}
}

func TestCompileWildcardPatternsPrefixMatchesExtensionless(t *testing.T) {
	regexes := compileWildcardPatterns([]string{"UnityCrashHandler"})
	if len(regexes) != 1 {
		t.Fatalf("got %d regexes, want 1", len(regexes))
	}
	if !regexes[0].MatchString("UnityCrashHandler64.exe") || !regexes[0].MatchString("UnityCrashHandler.exe") {
		t.Error("extension-less names must match by prefix")
	}
	if regexes[0].MatchString("MyUnityCrashHandler.exe") {
		t.Error("prefix match must start at the beginning of the name")
	}
}

func TestCompileWildcardPatternsKeepsDottedNamesExact(t *testing.T) {
	regexes := compileWildcardPatterns([]string{"Steam.exe"})
	if len(regexes) != 1 {
		t.Fatalf("got %d regexes, want 1", len(regexes))
	}
	if !regexes[0].MatchString("Steam.exe") || regexes[0].MatchString("SteamTmp.exe") {
		t.Error("dotted names must match exactly")
	}
}

func TestScanDirectoryFilesFiltersByDepthExtensionAndExclude(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "top.exe"))
	mustWrite(t, filepath.Join(root, "Steam.exe"))
	mustWrite(t, filepath.Join(root, "readme.txt"))
	mustWrite(t, filepath.Join(root, "game", "play.EXE"))
	mustWrite(t, filepath.Join(root, "game", "deep", "deeper.exe"))

	found := scanDirectoryFiles(root, scanOptions{
		MaxDepth:        1,
		Extensions:      []string{".exe"},
		ExcludePatterns: []string{"Steam.exe"},
	})

	want := map[string]bool{
		filepath.Join(root, "top.exe"):          true,
		filepath.Join(root, "game", "play.EXE"): true,
	}
	if len(found) != len(want) {
		t.Fatalf("found %v, want %d files", found, len(want))
	}
	for _, path := range found {
		if !want[path] {
			t.Errorf("unexpected file %s", path)
		}
	}
}

func TestScanDirectoryFilesMissingFolderIsEmpty(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	if found := scanDirectoryFiles(missing, scanOptions{MaxDepth: unlimitedDepth}); len(found) != 0 {
		t.Fatalf("missing folder returned %v", found)
	}
}

func mustWrite(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}
