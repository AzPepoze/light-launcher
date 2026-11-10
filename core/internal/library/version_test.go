package library

import "testing"

func TestCompareVersionsNumericParts(t *testing.T) {
	if compareVersions("1.10", "1.9") <= 0 {
		t.Error("1.10 should be newer than 1.9")
	}
	if compareVersions("0.7", "0.7.0") != 0 {
		t.Error("0.7 and 0.7.0 should compare equal")
	}
}

func TestCompareVersionsChapters(t *testing.T) {
	if compareVersions("Ch.5", "Ch.4") <= 0 {
		t.Error("Ch.5 should be newer than Ch.4")
	}
}

func TestIsNewerHandlesEmptyValues(t *testing.T) {
	cases := []struct {
		candidate, current string
		want               bool
	}{
		{"0.8", "0.7", true},
		{"0.7", "0.8", false},
		{"0.7", "0.7", false},
		{"", "1", false},
		{"1", "", true},
	}
	for _, c := range cases {
		if got := isNewer(c.candidate, c.current); got != c.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", c.candidate, c.current, got, c.want)
		}
	}
}

func TestSameVersionIgnoresPrefixAndTrailingZeros(t *testing.T) {
	if !sameVersion("0.7", "v0.7.0") {
		t.Error("0.7 and v0.7.0 should be the same version")
	}
	if sameVersion("0.7", "0.7a") {
		t.Error("0.7 and 0.7a should differ")
	}
}

func TestVersionFromText(t *testing.T) {
	if got := versionFromText("Game_v1.2.3_win.zip"); got != "1.2.3" {
		t.Errorf("versionFromText = %q, want 1.2.3", got)
	}
	if got := versionFromText("Game 2024"); got != "" {
		t.Errorf("a bare year must not be a version, got %q", got)
	}
}
