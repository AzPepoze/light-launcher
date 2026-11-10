package runtimesvc

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

// The LSFG settings files are small TOML files. These patterns read single values, as the Electron app did.
var (
	globalDllPattern          = regexp.MustCompile(`(?i)dll\s*=\s*"([^"]+)"`)
	globalFp16Pattern         = regexp.MustCompile(`(?i)allow_fp16\s*=\s*(true|false)`)
	profileMultiplierPattern  = regexp.MustCompile(`(?i)multiplier\s*=\s*(\d+)`)
	profilePerformancePattern = regexp.MustCompile(`(?i)performance_mode\s*=\s*(true|false)`)
	profileGpuPattern         = regexp.MustCompile(`(?i)gpu\s*=\s*"([^"]+)"`)
	profileFlowPattern        = regexp.MustCompile(`(?i)flow_scale\s*=\s*([\d.]+)`)
	profilePacingPattern      = regexp.MustCompile(`(?i)pacing\s*=\s*"([^"]+)"`)
)

var losslessDllCandidates = []string{
	"~/.steam/root/steamapps/common/Lossless Scaling/Lossless.dll",
	"~/.local/share/Steam/steamapps/common/Lossless Scaling/Lossless.dll",
	"~/.var/app/com.valvesoftware.Steam/.local/share/Steam/steamapps/common/Lossless Scaling/Lossless.dll",
}

// lsfgGlobalConfigPath is the lsfg-vk settings file. It lives in ~/.config, not in the XDG config folder.
func lsfgGlobalConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "lsfg-vk", "lsfg-vk.toml")
}

// detectLosslessDll returns the first Lossless.dll found in a Steam library, or "".
func detectLosslessDll() string {
	for _, candidate := range losslessDllCandidates {
		expanded := expandPath(candidate)
		if pathExists(expanded) {
			return expanded
		}
	}
	return ""
}

// getLsfgProfileForGame reads the game's LSFG profile. It returns nil when there is no profile file.
func getLsfgProfileForGame(name string, gamePath string) *types.LsfgProfileData {
	profilePath := config.GetGameLsfgConfigPath(name, lsfgIdentity(gamePath))
	dllPath, allowFp16 := readGlobalLsfgSettings()

	content, err := os.ReadFile(profilePath)
	if err != nil {
		return nil
	}
	profile := parseLsfgProfile(string(content))
	profile.Name = name
	if profile.Name == "" {
		profile.Name = nameWithoutExt(filepath.Base(gamePath))
	}
	profile.DllPath = dllPath
	profile.AllowFp16 = allowFp16
	return &profile
}

// readGlobalLsfgSettings returns the DLL path and FP16 flag from the global lsfg-vk file.
func readGlobalLsfgSettings() (string, bool) {
	content, err := os.ReadFile(lsfgGlobalConfigPath())
	if err != nil {
		return "", false
	}
	dllPath := firstGroup(globalDllPattern, string(content))
	allowFp16 := strings.EqualFold(firstGroup(globalFp16Pattern, string(content)), "true")
	return dllPath, allowFp16
}

// parseLsfgProfile reads the per-game profile values. Missing values get the Electron app defaults.
func parseLsfgProfile(content string) types.LsfgProfileData {
	profile := types.LsfgProfileData{Multiplier: 2, FlowScale: 1.0, Pacing: "smooth"}

	if value, ok := firstMatch(profileMultiplierPattern, content); ok {
		if number, ok := parseIntLike(value); ok {
			profile.Multiplier = number
		}
	}
	profile.PerformanceMode = strings.EqualFold(firstGroup(profilePerformancePattern, content), "true")
	profile.GPU = firstGroup(profileGpuPattern, content)
	if value, ok := firstMatch(profileFlowPattern, content); ok {
		if number, ok := parseFloatLike(value); ok {
			profile.FlowScale = float32(number)
		}
	}
	if value, ok := firstMatch(profilePacingPattern, content); ok {
		profile.Pacing = value
	}
	return profile
}

// firstMatch returns the first capture group of pattern and whether it matched.
func firstMatch(pattern *regexp.Regexp, content string) (string, bool) {
	match := pattern.FindStringSubmatch(content)
	if match == nil {
		return "", false
	}
	return match[1], true
}

func firstGroup(pattern *regexp.Regexp, content string) string {
	value, _ := firstMatch(pattern, content)
	return value
}
