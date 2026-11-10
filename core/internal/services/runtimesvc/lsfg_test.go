package runtimesvc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

// isolateHome points HOME (for ~/.config/lsfg-vk) and the config folder at temp dirs.
func isolateHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("APPDATA", home)
	return home
}

func TestParseLsfgProfileReadsValues(t *testing.T) {
	content := `[profile]
name = "Game"
multiplier = 3
performance_mode = true
gpu = "NVIDIA RTX"
flow_scale = 1.5
pacing = "hard"
`
	profile := parseLsfgProfile(content)
	if profile.Multiplier != 3 || !profile.PerformanceMode || profile.GPU != "NVIDIA RTX" {
		t.Errorf("profile = %+v, want multiplier 3, performance on, GPU NVIDIA RTX", profile)
	}
	if profile.FlowScale != 1.5 || profile.Pacing != "hard" {
		t.Errorf("flow/pacing = %v/%q, want 1.5/hard", profile.FlowScale, profile.Pacing)
	}
}

func TestParseLsfgProfileUsesDefaults(t *testing.T) {
	profile := parseLsfgProfile("[profile]\n")
	if profile.Multiplier != 2 || profile.FlowScale != 1.0 || profile.Pacing != "smooth" || profile.PerformanceMode {
		t.Errorf("defaults = %+v, want multiplier 2, flow 1.0, pacing smooth", profile)
	}
}

func TestFlowScaleValueFallsBackToOne(t *testing.T) {
	tests := map[string]float64{"": 1.0, "abc": 1.0, "0": 1.0, "2.5x": 2.5, " 0.75": 0.75}
	for input, want := range tests {
		if got := flowScaleValue(input); got != want {
			t.Errorf("flowScaleValue(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestParseIntLikeReadsLeadingDigits(t *testing.T) {
	if got, ok := parseIntLike("3x"); !ok || got != 3 {
		t.Errorf("parseIntLike(3x) = %d, %v, want 3, true", got, ok)
	}
	if _, ok := parseIntLike("x"); ok {
		t.Error("parseIntLike(x) reported a number")
	}
}

func TestFindLsfgAssetPicksFirstCompatiblePackage(t *testing.T) {
	releases := []lsfgRelease{
		{Assets: []lsfgAsset{{Name: "source.tar.gz", BrowserDownloadURL: "https://x/source"}}},
		{Assets: []lsfgAsset{
			{Name: "lsfg-vk-2.0-x86_64.TAR.ZST", BrowserDownloadURL: "https://x/zst"},
			{Name: "lsfg-vk-linux.tar.xz", BrowserDownloadURL: "https://x/xz"},
		}},
	}

	url, name := findLsfgAsset(releases)
	if url != "https://x/zst" || name != "lsfg-vk-2.0-x86_64.TAR.ZST" {
		t.Errorf("asset = %q %q, want the x86_64 zst package", url, name)
	}

	if url, _ := findLsfgAsset([]lsfgRelease{{Assets: []lsfgAsset{{Name: "notes.txt"}}}}); url != "" {
		t.Errorf("no compatible asset gave %q, want empty", url)
	}
}

func TestSaveLsfgProfileRoundTrip(t *testing.T) {
	home := isolateHome(t)
	gamePath := filepath.Join(home, "games", "Game.exe")

	input := SaveLsfgProfileInput{
		ProfileName:     "Game",
		GamePath:        gamePath,
		Multiplier:      3,
		PerformanceMode: true,
		DllPath:         "/libs/Lossless.dll",
		GPU:             "GPU0",
		FlowScale:       "1.5",
		Pacing:          "hard",
		AllowFp16:       true,
	}
	if err := saveLsfgProfile(input); err != nil {
		t.Fatalf("saveLsfgProfile: %v", err)
	}

	profile := getLsfgProfileForGame("Game", gamePath)
	if profile == nil {
		t.Fatal("getLsfgProfileForGame returned nil after save")
	}
	if profile.Multiplier != 3 || profile.GPU != "GPU0" || profile.DllPath != "/libs/Lossless.dll" || !profile.AllowFp16 {
		t.Errorf("profile = %+v, want saved values", profile)
	}
	if profile.FlowScale != 1.5 || profile.Pacing != "hard" || !profile.PerformanceMode {
		t.Errorf("profile flow/pacing/perf = %v %q %v, want saved values", profile.FlowScale, profile.Pacing, profile.PerformanceMode)
	}
}

func TestDisableLsfgProfileAppendsOnce(t *testing.T) {
	isolateHome(t)
	input := SaveLsfgProfileInput{ProfileName: "Game", GamePath: "/games/Game.exe", GPU: "GPU0", FlowScale: "1"}
	if err := saveLsfgProfile(input); err != nil {
		t.Fatalf("saveLsfgProfile: %v", err)
	}

	for i := 0; i < 2; i++ {
		if err := disableLsfgProfile("Game", "/games/Game.exe"); err != nil {
			t.Fatalf("disableLsfgProfile: %v", err)
		}
	}

	content, err := os.ReadFile(config.GetGameLsfgConfigPath("Game", ""))
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}
	if count := strings.Count(string(content), "disable = true"); count != 1 {
		t.Errorf("disable line count = %d, want 1", count)
	}
}

func TestRemoveProfileDeletesGameProfileFile(t *testing.T) {
	isolateHome(t)
	cfg := types.LaunchOptions{Name: "Game", GamePath: "/games/Game.exe"}
	if err := saveGameConfig(&cfg); err != nil {
		t.Fatalf("saveGameConfig: %v", err)
	}
	profilePath := config.GetGameLsfgConfigPath("Game", cfg.ID)
	if err := os.MkdirAll(filepath.Dir(profilePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(profilePath, []byte("[profile]\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := removeProfile("/games/Game.exe"); err != nil {
		t.Fatalf("removeProfile: %v", err)
	}
	if pathExists(profilePath) {
		t.Error("profile file still exists after removeProfile")
	}
}

func TestDetectLosslessDllReturnsEmptyWhenMissing(t *testing.T) {
	isolateHome(t)
	if got := detectLosslessDll(); got != "" {
		t.Errorf("detectLosslessDll = %q, want empty", got)
	}
}
