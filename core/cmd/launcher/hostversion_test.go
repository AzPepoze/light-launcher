package main

import "testing"

func TestHostVersion(t *testing.T) {
	cases := map[string]string{
		"dev":               "2.0.0",
		"":                  "2.0.0",
		"32b4339":           "2.0.0",
		"2.1.0":             "2.1.0",
		"v2.1.0":            "2.1.0",
		"v2.1.0-3-gabc1234": "2.1.0",
	}
	for input, want := range cases {
		if got := hostVersion(input); got != want {
			t.Errorf("hostVersion(%q) = %q, want %q", input, got, want)
		}
	}
}
