package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanPathNormalizesAndTrims(t *testing.T) {
	got := CleanPath(filepath.FromSlash("   /home/user/games/../games/test.exe   "))
	want := filepath.FromSlash("/home/user/games/test.exe")
	if got != want {
		t.Fatalf("CleanPath = %q, want %q", got, want)
	}
	if CleanPath("") != "" {
		t.Fatal("CleanPath of empty input must stay empty")
	}
}

func TestNormalizePathKeepsTrailingSeparator(t *testing.T) {
	if got := NormalizePath(""); got != "." {
		t.Fatalf("NormalizePath(\"\") = %q, want \".\"", got)
	}
	trailing := filepath.FromSlash("/a/b/")
	if got := NormalizePath(trailing); got != trailing {
		t.Fatalf("NormalizePath(%q) = %q, want %q", trailing, got, trailing)
	}
}

func TestNameWithoutExtMatchesNodeParse(t *testing.T) {
	cases := map[string]string{
		"game.exe":         "game",
		"Game.Part.exe":    "Game.Part",
		".bashrc":          ".bashrc",
		"noext":            "noext",
		"dir/Launcher.exe": "Launcher",
	}
	for input, want := range cases {
		if got := NameWithoutExt(filepath.FromSlash(input)); got != want {
			t.Errorf("NameWithoutExt(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestBaseNameReturnsEmptyForRoot(t *testing.T) {
	if BaseName("") != "" || BaseName("/") != "" {
		t.Fatal("BaseName must return an empty string for empty and root paths")
	}
}

func TestPathKeyNormalizesWindowsStylePaths(t *testing.T) {
	want := strings.ToLower(filepath.Clean(filepath.FromSlash("c:/games/game.exe")))
	cases := []string{
		`C:\Games\Game.exe`,
		`c:/games/game.exe`,
		`C:\Games\..\Games\Game.exe`,
		`c:\games\.\game.exe`,
	}
	for _, input := range cases {
		if got := PathKey(input); got != want {
			t.Errorf("PathKey(%q) = %q, want %q", input, got, want)
		}
	}
	if PathKey("") != "" {
		t.Error("PathKey of empty input must stay empty")
	}
}

func TestIsSubPathDetectsChildren(t *testing.T) {
	parent := filepath.FromSlash("/home/user/games")
	if !IsSubPath(parent, filepath.FromSlash("/home/user/games/game1/game.exe")) {
		t.Error("expected a child path to be inside the parent")
	}
	if IsSubPath(parent, filepath.FromSlash("/home/user/other/game.exe")) {
		t.Error("expected a sibling path to be outside the parent")
	}
	if IsSubPath(parent, parent) {
		t.Error("a folder is not a sub path of itself")
	}
}

func TestExpandPathHandlesTildeAndUnicode(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory")
	}
	if got := ExpandPath("~"); got != home {
		t.Errorf("ExpandPath(~) = %q, want %q", got, home)
	}
	if got := ExpandPath("~/games/steam"); got != filepath.Join(home, "games", "steam") {
		t.Errorf("ExpandPath(~/games/steam) = %q", got)
	}
	unicodePath := filepath.Join(string(filepath.Separator), "mnt", "Drive1", "Game♥", "OK", "Lustful Spirit Hunt.exe")
	if got := ExpandPath(unicodePath); got != unicodePath {
		t.Errorf("ExpandPath kept unicode path = %q, want %q", got, unicodePath)
	}
	if ExpandPath("") != "" {
		t.Error("ExpandPath of empty input must stay empty")
	}
}

func TestPathExistsNeverFails(t *testing.T) {
	if PathExists("") {
		t.Error("empty path must not exist")
	}
	if PathExists(filepath.Join(t.TempDir(), "missing", "file.exe")) {
		t.Error("missing path must not exist")
	}
	home, err := os.UserHomeDir()
	if err == nil && !PathExists(home) {
		t.Error("home directory must exist")
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := FirstNonEmpty("", "", "value"); got != "value" {
		t.Errorf("FirstNonEmpty = %q", got)
	}
	if got := FirstNonEmpty("", ""); got != "" {
		t.Errorf("FirstNonEmpty of blanks = %q", got)
	}
}
