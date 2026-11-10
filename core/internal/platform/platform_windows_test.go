package platform

import "testing"

func TestCurrentFeaturesOnWindows(t *testing.T) {
	tests := []struct {
		name    string
		enabled func(Features) bool
		want    bool
	}{
		{"proton", func(f Features) bool { return f.Proton }, false},
		{"prefix", func(f Features) bool { return f.Prefix }, false},
		{"lsfg", func(f Features) bool { return f.Lsfg }, false},
		{"gamescope", func(f Features) bool { return f.Gamescope }, false},
		{"mangohud", func(f Features) bool { return f.MangoHud }, false},
		{"gamemode", func(f Features) bool { return f.GameMode }, false},
		{"live terminal logs", func(f Features) bool { return f.LiveTerminalLogs }, false},
	}

	features := Current()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.enabled(features); got != test.want {
				t.Errorf("enabled = %v, want %v", got, test.want)
			}
		})
	}
}

func TestOSIsWindows(t *testing.T) {
	if got := OS(); got != "windows" {
		t.Errorf("OS() = %q, want %q", got, "windows")
	}
}
