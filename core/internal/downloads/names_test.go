package downloads

import (
	"strings"
	"testing"
)

func TestSanitizeFilenameKeepsOnePathComponent(t *testing.T) {
	cases := map[string]string{
		"../../etc/passwd": "passwd",
		`..\evil.exe`:      "evil.exe",
		"a:b*c?.zip":       "a_b_c_.zip",
		"   ":              "download.bin",
		"name. ":           "name",
	}
	for in, want := range cases {
		if got := sanitizeFilename(in); got != want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSanitizeFilenameTruncatesKeepingExtension(t *testing.T) {
	long := strings.Repeat("a", 300) + ".zip"
	got := sanitizeFilename(long)
	if len([]rune(got)) != maxFilenameRunes || !strings.HasSuffix(got, ".zip") {
		t.Fatalf("got %d runes ending %q", len([]rune(got)), got[len(got)-4:])
	}
}

func TestFilenameFromDisposition(t *testing.T) {
	cases := map[string]string{
		`attachment; filename="Game Setup.zip"`:     "Game Setup.zip",
		`attachment; filename=plain.rar`:            "plain.rar",
		`attachment; filename*=UTF-8''My%20Game.7z`: "My Game.7z",
		"":       "",
		"inline": "",
	}
	for in, want := range cases {
		if got := filenameFromDisposition(in); got != want {
			t.Errorf("filenameFromDisposition(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsGenericFilename(t *testing.T) {
	generic := []string{"", "download.bin", "file.zip", "pixeldrain.zip", "setup.exe", "mega.nz"}
	for _, name := range generic {
		if !isGenericFilename(name) {
			t.Errorf("%q should be generic", name)
		}
	}
	if isGenericFilename("Dungeon Crawl v2.zip") {
		t.Error("a real game name must not be generic")
	}
}

func TestEnsureFilenameExtension(t *testing.T) {
	cases := []struct {
		name, contentType, want string
	}{
		{"game.zip", "application/octet-stream", "game.zip"},
		{"game", "application/zip", "game.zip"},
		{"game", "application/x-7z-compressed; charset=binary", "game.7z"},
		{"game", "text/plain", "game.bin"},
	}
	for _, c := range cases {
		if got := ensureFilenameExtension(c.name, c.contentType); got != c.want {
			t.Errorf("ensureFilenameExtension(%q, %q) = %q, want %q", c.name, c.contentType, got, c.want)
		}
	}
}

func TestSameSiteHost(t *testing.T) {
	if !sameSiteHost("cdn.example.com", "example.com") {
		t.Error("subdomain should match its parent")
	}
	if sameSiteHost("evilexample.com", "example.com") {
		t.Error("a suffix match without a dot must not match")
	}
}

func TestDefaultRefererUsesParentDomain(t *testing.T) {
	if got := defaultReferer("cdn.pixeldrain.com"); got != "https://pixeldrain.com/" {
		t.Fatalf("referer = %q", got)
	}
	if got := defaultReferer("example.org"); got != "" {
		t.Fatalf("unexpected referer %q", got)
	}
}
