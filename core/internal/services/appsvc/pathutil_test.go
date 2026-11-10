package appsvc

import (
	"path/filepath"
	"testing"
)

func fromSlash(p string) string {
	return filepath.FromSlash(p)
}

func TestCleanPathNormalizesAndTrims(t *testing.T) {
	got := cleanPath(fromSlash("   /home/user/games/../games/test.exe   "))
	want := fromSlash("/home/user/games/test.exe")
	if got != want {
		t.Fatalf("cleanPath = %q, want %q", got, want)
	}
	if cleanPath("") != "" {
		t.Fatal("cleanPath of empty input must stay empty")
	}
}

func TestIsSubPathDetectsChildren(t *testing.T) {
	parent := fromSlash("/home/user/games")
	if !isSubPath(parent, fromSlash("/home/user/games/game1/game.exe")) {
		t.Error("expected a child path to be inside the parent")
	}
	if isSubPath(parent, fromSlash("/home/user/other/game.exe")) {
		t.Error("expected a sibling path to be outside the parent")
	}
	if isSubPath(parent, parent) {
		t.Error("a folder is not a sub path of itself")
	}
}

func TestNormalizePathKeepsTrailingSeparator(t *testing.T) {
	if got := normalizePath(""); got != "." {
		t.Fatalf("normalizePath(\"\") = %q, want \".\"", got)
	}
	trailing := fromSlash("/a/b/")
	if got := normalizePath(trailing); got != trailing {
		t.Fatalf("normalizePath(%q) = %q, want %q", trailing, got, trailing)
	}
}

func TestFileNameWithoutExtMatchesNodeParse(t *testing.T) {
	cases := map[string]string{
		"game.exe":         "game",
		"Game.Part.exe":    "Game.Part",
		".bashrc":          ".bashrc",
		"noext":            "noext",
		"dir/Launcher.exe": "Launcher",
	}
	for input, want := range cases {
		if got := fileNameWithoutExt(fromSlash(input)); got != want {
			t.Errorf("fileNameWithoutExt(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestBaseNameReturnsEmptyForRoot(t *testing.T) {
	if baseName("") != "" || baseName("/") != "" {
		t.Fatal("baseName must return an empty string for empty and root paths")
	}
}
