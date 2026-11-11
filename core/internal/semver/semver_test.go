package semver

import "testing"

func TestParseAcceptsPlainAndPreReleaseVersions(t *testing.T) {
	for _, raw := range []string{"1.2.3", "v1.2.3", "0.0.0", "2.0.0-beta.1", "1.0.0-rc.1+build.5"} {
		if _, err := Parse(raw); err != nil {
			t.Errorf("Parse(%q): %v", raw, err)
		}
	}
}

func TestParseRejectsBadVersions(t *testing.T) {
	for _, raw := range []string{"", "1.2", "1.2.3.4", "x.y.z", "1.02.3", "1.0.0-", "1.0.0-beta..1", "1.0.0-beta_1"} {
		if _, err := Parse(raw); err == nil {
			t.Errorf("Parse(%q) succeeded, want an error", raw)
		}
	}
}

func TestCompareOrdersPreReleases(t *testing.T) {
	cases := []struct {
		left, right string
		want        int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.1.0", "1.0.9", 1},
		{"1.0.0-beta.1", "1.0.0", -1},
		{"1.0.0", "1.0.0-rc.1", 1},
		{"1.0.0-alpha", "1.0.0-beta", -1},
		{"1.0.0-beta.2", "1.0.0-beta.10", -1},
		{"1.0.0-beta", "1.0.0-beta.1", -1},
		{"1.0.0-beta.1", "1.0.0-beta.1+build", 0},
		{"1.0.0-1", "1.0.0-alpha", -1},
	}
	for _, tc := range cases {
		got, err := CompareStrings(tc.left, tc.right)
		if err != nil {
			t.Fatalf("CompareStrings(%q, %q): %v", tc.left, tc.right, err)
		}
		if got != tc.want {
			t.Errorf("CompareStrings(%q, %q) = %d, want %d", tc.left, tc.right, got, tc.want)
		}
	}
}
