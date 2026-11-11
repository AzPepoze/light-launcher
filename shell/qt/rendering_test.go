package main

import "testing"

const (
	gpuOn      = "--ignore-gpu-blocklist --enable-gpu-rasterization"
	autoscroll = "--enable-blink-features=MiddleClickAutoscroll"
)

func TestChromiumFlags(t *testing.T) {
	tests := []struct {
		name       string
		existing   string
		disableGPU bool
		want       string
	}{
		{"adds the gpu and autoscroll flags to nothing", "", false, gpuOn + " " + autoscroll},
		{"keeps the user's flags first", "--foo=1", false, "--foo=1 " + gpuOn + " " + autoscroll},
		{"does not repeat a flag", "--ignore-gpu-blocklist", false, "--ignore-gpu-blocklist --enable-gpu-rasterization " + autoscroll},
		{"adds nothing when all are set", "--enable-gpu-rasterization --ignore-gpu-blocklist " + autoscroll, false, "--enable-gpu-rasterization --ignore-gpu-blocklist " + autoscroll},
		{"the kill switch turns the gpu off", "", true, "--disable-gpu " + autoscroll},
		{"the kill switch keeps the user's flags", "--foo=1", true, "--foo=1 --disable-gpu " + autoscroll},
		{"extra spaces are tidied", "  --foo=1   --bar ", false, "--foo=1 --bar " + gpuOn + " " + autoscroll},
		{"joins the user's blink features", "--enable-blink-features=Foo,Bar", false, "--enable-blink-features=Foo,Bar,MiddleClickAutoscroll " + gpuOn},
		{"does not repeat the autoscroll feature", "--enable-blink-features=MiddleClickAutoscroll,Foo", false, "--enable-blink-features=MiddleClickAutoscroll,Foo " + gpuOn},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := chromiumFlags(test.existing, test.disableGPU); got != test.want {
				t.Fatalf("chromiumFlags(%q, %v) = %q, want %q", test.existing, test.disableGPU, got, test.want)
			}
		})
	}
}
