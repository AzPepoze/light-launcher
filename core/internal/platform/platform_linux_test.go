package platform

import "testing"

func TestCurrentFeaturesOnLinux(t *testing.T) {
	tests := []struct {
		name    string
		enabled func(Features) bool
		want    bool
	}{
		{"proton", func(f Features) bool { return f.Proton }, true},
		{"prefix", func(f Features) bool { return f.Prefix }, true},
		{"lsfg", func(f Features) bool { return f.Lsfg }, true},
		{"gamescope", func(f Features) bool { return f.Gamescope }, true},
		{"mangohud", func(f Features) bool { return f.MangoHud }, true},
		{"gamemode", func(f Features) bool { return f.GameMode }, true},
		{"live terminal logs", func(f Features) bool { return f.LiveTerminalLogs }, true},
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

func TestOSIsLinux(t *testing.T) {
	if got := OS(); got != "linux" {
		t.Errorf("OS() = %q, want %q", got, "linux")
	}
}
