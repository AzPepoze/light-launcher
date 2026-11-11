package runtimesvc

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"light-launcher/core/internal/config"
)

// SaveLsfgProfileInput is the payload of SaveLsfgProfile.
type SaveLsfgProfileInput struct {
	ProfileName     string `json:"profileName"`
	GamePath        string `json:"gamePath"`
	Multiplier      int    `json:"multiplier"`
	PerformanceMode bool   `json:"performanceMode"`
	DllPath         string `json:"dllPath"`
	GPU             string `json:"gpu"`
	FlowScale       string `json:"flowScale"`
	Pacing          string `json:"pacing"`
	AllowFp16       bool   `json:"allowFp16"`
}

// saveLsfgProfile writes the game's profile and the global lsfg-vk file (DLL path and FP16 flag).
// An empty GPU uses the first detected GPU.
func (s *service) saveLsfgProfile(input SaveLsfgProfileInput) error {
	profilePath := s.lsfgConfigPath(input.ProfileName, s.lsfgIdentity(input.GamePath))

	gpu := input.GPU
	if gpu == "" {
		if gpus := listGpuNames(); len(gpus) > 0 {
			gpu = gpus[0]
		}
	}

	if err := os.MkdirAll(filepath.Dir(profilePath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(profilePath, []byte(profileToml(input, gpu)), 0644); err != nil {
		return err
	}
	return writeGlobalLsfgSettings(input.DllPath, input.AllowFp16)
}

func profileToml(input SaveLsfgProfileInput, gpu string) string {
	return fmt.Sprintf(`[profile]
name = "%s"
multiplier = %d
performance_mode = %s
gpu = "%s"
flow_scale = %s
pacing = "%s"
allow_fp16 = %s
`,
		input.ProfileName,
		intOrDefault(input.Multiplier, 2),
		boolToml(input.PerformanceMode),
		gpu,
		strconv.FormatFloat(flowScaleValue(input.FlowScale), 'f', -1, 64),
		stringOrDefault(input.Pacing, "smooth"),
		boolToml(input.AllowFp16),
	)
}

func writeGlobalLsfgSettings(dllPath string, allowFp16 bool) error {
	globalPath := lsfgGlobalConfigPath()
	if err := os.MkdirAll(filepath.Dir(globalPath), 0755); err != nil {
		return err
	}
	content := fmt.Sprintf("[global]\ndll = \"%s\"\nallow_fp16 = %s\n", dllPath, boolToml(allowFp16))
	return os.WriteFile(globalPath, []byte(content), 0644)
}

// disableLsfgProfile adds "disable = true" to the game's profile file, once.
func (s *service) disableLsfgProfile(profileName string, gamePath string) error {
	profilePath := s.lsfgConfigPath(profileName, s.lsfgIdentity(gamePath))
	if !config.PathExists(profilePath) {
		return nil
	}
	content, err := os.ReadFile(profilePath)
	if err != nil {
		return err
	}
	if strings.Contains(string(content), "disable = true") {
		return nil
	}
	updated := string(content) + "\ndisable = true\n"
	return os.WriteFile(profilePath, []byte(updated), 0644)
}

// removeProfile deletes the game's profile file. It does nothing when the game has no config.
func (s *service) removeProfile(mainExecutablePath string) error {
	cfg := s.games.Find(mainExecutablePath)
	if cfg == nil {
		return nil
	}
	profilePath := s.lsfgConfigPath(cfg.Name, cfg.ID)
	if !config.PathExists(profilePath) {
		return nil
	}
	return os.Remove(profilePath)
}

// lsfgConfigPath is the per-game toc file inside the game's config folder.
func (s *service) lsfgConfigPath(name, id string) string {
	return filepath.Join(s.games.Path(name, id), "lsfg_vk.toml")
}

// lsfgIdentity returns the config ID used to find a game's LSFG profile folder.
func (s *service) lsfgIdentity(gamePath string) string {
	cfg := s.games.Find(gamePath)
	if cfg == nil {
		return ""
	}
	return cfg.ID
}

func boolToml(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func stringOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
