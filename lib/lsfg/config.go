package lsfg

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

func GetConfigPath() (string, error) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDirectory, ".config", "lsfg-vk", "conf.toml"), nil
}

func GetProfilePath(gamePath, baseDirectory string) string {
	hasher := sha1.New()
	hasher.Write([]byte(gamePath))
	hash := hex.EncodeToString(hasher.Sum(nil))[:8]
	executableName := filepath.Base(gamePath)
	baseName := strings.TrimSuffix(executableName, ".exe")
	baseName = strings.TrimSuffix(baseName, ".EXE")

	filename := baseName + "-" + hash + ".toml"
	configDirectory := filepath.Join(baseDirectory, "config", "lsfg")
	return filepath.Join(configDirectory, filename)
}

func FindProfileForGameAtPath(gamePath, configPath string) (*ConfigProfile, int, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, -1, fmt.Errorf("failed to read LSFG config: %w", err)
	}

	var config ConfigFile
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, -1, fmt.Errorf("failed to parse LSFG config: %w", err)
	}

	executableName := strings.ToLower(filepath.Base(gamePath))

	for index, profile := range config.Profiles {
		if matchesProfile(executableName, profile.ActiveIn) {
			return &profile, index, nil
		}
	}

	targetName := strings.TrimSuffix(executableName, filepath.Ext(executableName))
	for index, profile := range config.Profiles {
		if strings.EqualFold(profile.Name, targetName) {
			return &profile, index, nil
		}
	}

	return nil, -1, fmt.Errorf("no matching LSFG profile found for %s", executableName)
}

func FindProfileForGame(gamePath string) (*ConfigProfile, int, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, -1, err
	}
	return FindProfileForGameAtPath(gamePath, configPath)
}

func matchesProfile(executableName string, activeIn interface{}) bool {
	executableName = strings.ToLower(executableName)

	switch value := activeIn.(type) {
	case string:
		return strings.EqualFold(value, executableName)
	case []interface{}:
		for _, item := range value {
			if str, ok := item.(string); ok {
				if strings.EqualFold(str, executableName) {
					return true
				}
			}
		}
	}
	return false
}

func SaveProfileToPath(profileName, gamePath, configPath string, multiplier int, performanceMode bool, dllPath, gpu, flowScale, pacing string, allowFp16 bool) error {
	configDirectory := filepath.Dir(configPath)
	if err := os.MkdirAll(configDirectory, 0755); err != nil {
		return err
	}

	var config ConfigFile
	if data, err := os.ReadFile(configPath); err == nil {
		if err := toml.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse existing LSFG config: %w", err)
		}
	} else {
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

	config.Version = 2
	config.Global.Version = 2
	config.Global.DLL = dllPath
	config.Global.AllowFP16 = allowFp16

	executableName := filepath.Base(gamePath)
	found := false
	for index, profile := range config.Profiles {
		if matchesProfile(strings.ToLower(executableName), profile.ActiveIn) || strings.EqualFold(profile.Name, profileName) {
			config.Profiles[index].Name = profileName
			config.Profiles[index].ActiveIn = executableName
			config.Profiles[index].Multiplier = multiplier
			config.Profiles[index].PerformanceMode = performanceMode
			config.Profiles[index].GPU = gpu
			config.Profiles[index].FlowScale = parseFlowScale(flowScale)
			config.Profiles[index].Pacing = pacing
			found = true
			break
		}
	}

	if !found {
		newProfile := ConfigProfile{
			Name:            profileName,
			ActiveIn:        executableName,
			Multiplier:      multiplier,
			PerformanceMode: performanceMode,
			GPU:             gpu,
			FlowScale:       parseFlowScale(flowScale),
			Pacing:          pacing,
		}
		config.Profiles = append(config.Profiles, newProfile)
	}

	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal LSFG config: %w", err)
	}

	return os.WriteFile(configPath, data, 0644)
}

func SaveProfileToGlobal(profileName, gamePath string, multiplier int, performanceMode bool, dllPath, gpu, flowScale, pacing string, allowFp16 bool) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}
	return SaveProfileToPath(profileName, gamePath, configPath, multiplier, performanceMode, dllPath, gpu, flowScale, pacing, allowFp16)
}

func DisableProfileInConfig(profileName, gamePath string) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read LSFG config: %w", err)
	}

	var config ConfigFile
	if err := toml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse LSFG config: %w", err)
	}

	executableName := filepath.Base(gamePath)
	found := false

	for index, profile := range config.Profiles {
		if matchesProfile(strings.ToLower(executableName), profile.ActiveIn) || strings.EqualFold(profile.Name, profileName) {
			config.Profiles[index].ActiveIn = ""
			found = true
			break
		}
	}

	if !found {
		return nil
	}

	data, err = toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal LSFG config: %w", err)
	}

	return os.WriteFile(configPath, data, 0644)
}

func RemoveProfileFromConfig(gamePath string) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read LSFG config: %w", err)
	}

	var config ConfigFile
	if err := toml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse LSFG config: %w", err)
	}

	executableName := filepath.Base(gamePath)
	found := false
	for index, profile := range config.Profiles {
		if matchesProfile(strings.ToLower(executableName), profile.ActiveIn) {
			config.Profiles = append(config.Profiles[:index], config.Profiles[index+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("no profile found for %s", executableName)
	}

	data, err = toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal LSFG config: %w", err)
	}

	return os.WriteFile(configPath, data, 0644)
}

func EditConfigForGame(gamePath string) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("LSFG config file not found at %s", configPath)
	}

	return openFileWithEditor(configPath)
}

func openFileWithEditor(filePath string) error {
	command := exec.Command("xdg-open", filePath)
	return command.Start()
}

func parseFlowScale(flowScale string) float32 {
	var value float32 = 1.0
	if flowScale != "" {
		var floatVal float64
		if _, err := fmt.Sscanf(flowScale, "%f", &floatVal); err == nil {
			value = float32(floatVal)
		}
	}
	return value
}
