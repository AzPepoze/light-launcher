package main

import "testing"

func TestChromiumFlags(t *testing.T) {
	tests := []struct {
		name       string
		existing   string
		disableGPU bool
		want       string
	}{
		{"adds the gpu flags to nothing", "", false, "--ignore-gpu-blocklist --enable-gpu-rasterization"},
		{"keeps the user's flags first", "--foo=1", false, "--foo=1 --ignore-gpu-blocklist --enable-gpu-rasterization"},
		{"does not repeat a flag", "--ignore-gpu-blocklist", false, "--ignore-gpu-blocklist --enable-gpu-rasterization"},
		{"adds nothing when all are set", "--enable-gpu-rasterization --ignore-gpu-blocklist", false, "--enable-gpu-rasterization --ignore-gpu-blocklist"},
		{"the kill switch turns the gpu off", "", true, "--disable-gpu"},
		{"the kill switch keeps the user's flags", "--foo=1", true, "--foo=1 --disable-gpu"},
		{"extra spaces are tidied", "  --foo=1   --bar ", false, "--foo=1 --bar --ignore-gpu-blocklist --enable-gpu-rasterization"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := chromiumFlags(test.existing, test.disableGPU); got != test.want {
				t.Fatalf("chromiumFlags(%q, %v) = %q, want %q", test.existing, test.disableGPU, got, test.want)
			}
		})
	}
}
