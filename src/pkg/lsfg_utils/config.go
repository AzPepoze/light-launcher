package lsfg_utils

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// GetConfigPath returns the path to the lsfg-vk config file (~/.config/lsfg-vk/conf.toml)
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "lsfg-vk", "conf.toml"), nil
}

// GetProfilePath returns the path to store LSFG profile for a game exe
// Uses format: LightLauncher/config/lsfg/exename-hash.toml
func GetProfilePath(gamePath, baseDir string) string {
	h := sha1.New()
	h.Write([]byte(gamePath))
	hash := hex.EncodeToString(h.Sum(nil))[:8] // First 8 chars of hash
	exeName := filepath.Base(gamePath)
	// Remove .exe extension for readability
	baseName := strings.TrimSuffix(exeName, ".exe")
	baseName = strings.TrimSuffix(baseName, ".EXE")

	filename := baseName + "-" + hash + ".toml"
	configDir := filepath.Join(baseDir, "config", "lsfg")
	return filepath.Join(configDir, filename)
}

// FindProfileForGameAtPath finds the profile for a game at a specific config path
func FindProfileForGameAtPath(gamePath, configPath string) (*ConfigProfile, int, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, -1, fmt.Errorf("failed to read LSFG config: %w", err)
	}

	var config ConfigFile
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, -1, fmt.Errorf("failed to parse LSFG config: %w", err)
	}

	exeName := strings.ToLower(filepath.Base(gamePath))

	// Find the matching profile
	for i, profile := range config.Profiles {
		if matchesProfile(exeName, profile.ActiveIn) {
			return &profile, i, nil
		}
	}

	return nil, -1, fmt.Errorf("no matching LSFG profile found for %s", exeName)
}

// FindProfileForGame finds the profile for a game using default config path
func FindProfileForGame(gamePath string) (*ConfigProfile, int, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, -1, err
	}
	return FindProfileForGameAtPath(gamePath, configPath)
}

// matchesProfile checks if the exe matches the profile's active_in list
func matchesProfile(exeName string, activeIn interface{}) bool {
	exeName = strings.ToLower(exeName)

	switch v := activeIn.(type) {
	case string:
		return strings.EqualFold(v, exeName)
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				if strings.EqualFold(s, exeName) {
					return true
				}
			}
		}
	}
	return false
}

// SaveProfileToPath saves a profile to the LSFG-VK config file at a specific path
func SaveProfileToPath(gamePath, configPath string, multiplier int, perfMode bool, dllPath, gpu, flowScale, pacing string, allowFp16 bool) error {
	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Read existing config or create new one
	var config ConfigFile
	if data, err := os.ReadFile(configPath); err == nil {
		if err := toml.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse existing LSFG config: %w", err)
		}
	} else {
		// Create new config with version 2
		config = ConfigFile{
			Version: 2,
			Global: GlobalConfig{
				Version:   2,
				AllowFP16: allowFp16,
				DLL:       dllPath,
			},
			Profiles: []ConfigProfile{},
		}
	}

	// Update global settings
	config.Version = 2
	config.Global.Version = 2
	config.Global.DLL = dllPath
	config.Global.AllowFP16 = allowFp16

	// Extract just the exe name for active_in
	exeName := filepath.Base(gamePath)

	// Find if profile already exists for this exe
	found := false
	for i, profile := range config.Profiles {
		if matchesProfile(strings.ToLower(exeName), profile.ActiveIn) {
			// Update existing profile
			config.Profiles[i].Multiplier = multiplier
			config.Profiles[i].PerformanceMode = perfMode
			config.Profiles[i].GPU = gpu
			config.Profiles[i].FlowScale = parseFlowScale(flowScale)
			config.Profiles[i].Pacing = pacing
			found = true
			break
		}
	}

	if !found {
		// Create new profile
		newProfile := ConfigProfile{
			Name:            strings.TrimSuffix(exeName, filepath.Ext(exeName)),
			ActiveIn:        exeName,
			Multiplier:      multiplier,
			PerformanceMode: perfMode,
			GPU:             gpu,
			FlowScale:       parseFlowScale(flowScale),
			Pacing:          pacing,
		}
		config.Profiles = append(config.Profiles, newProfile)
	}

	// Write back to file
	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal LSFG config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write LSFG config: %w", err)
	}

	return nil
}

// SaveProfileToGlobal saves a profile to the lsfg-vk config file
func SaveProfileToGlobal(gamePath string, multiplier int, perfMode bool, dllPath, gpu, flowScale, pacing string, allowFp16 bool) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}
	return SaveProfileToPath(gamePath, configPath, multiplier, perfMode, dllPath, gpu, flowScale, pacing, allowFp16)
}

// RemoveProfileFromConfig removes a profile for a game from the config file
func RemoveProfileFromConfig(gamePath string) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read LSFG config: %w", err)
	}

	var config ConfigFile
	if err := toml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse LSFG config: %w", err)
	}

	// Extract just the exe name for matching
	exeName := filepath.Base(gamePath)

	// Find and remove the profile for this exe
	found := false
	for i, profile := range config.Profiles {
		if matchesProfile(strings.ToLower(exeName), profile.ActiveIn) {
			// Remove profile by slicing
			config.Profiles = append(config.Profiles[:i], config.Profiles[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("no profile found for %s", exeName)
	}

	// Write back to file
	data, err = toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal LSFG config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write LSFG config: %w", err)
	}

	return nil
}

// EditConfigForGame opens the LSFG config file with the default editor
func EditConfigForGame(gamePath string) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("LSFG config file not found at %s", configPath)
	}

	// Open with default editor
	return openFileWithEditor(configPath)
}

// openFileWithEditor opens a file with the system default editor
func openFileWithEditor(filePath string) error {
	// Try xdg-open first (Linux)
	if cmd := exec.Command("xdg-open", filePath); cmd != nil {
		return cmd.Start()
	}
	return fmt.Errorf("failed to open editor")
}

// parseFlowScale parses a flowScale string to float32
func parseFlowScale(flowScale string) float32 {
	var val float32 = 1.0
	if flowScale != "" {
		// Try to parse as float
		var f float64
		if _, err := fmt.Sscanf(flowScale, "%f", &f); err == nil {
			val = float32(f)
		}
	}
	return val
}
