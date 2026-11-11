package library

import "testing"

func TestPrettyTitle(t *testing.T) {
	cases := map[string]string{
		"Mad_Island_0.34":     "Mad Island 0.34",
		"koh_windows_v02":     "koh v02",
		"Some Game (Patreon)": "Some Game",
		"Cool Game-x64":       "Cool Game",
		"Plain":               "Plain",
	}
	for in, want := range cases {
		if got := prettyTitle(in); got != want {
			t.Errorf("prettyTitle(%q) = %q, want %q", in, got, want)
		}
	}
}
