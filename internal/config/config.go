package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"light-launcher/internal/types"
	"light-launcher/lib/lsfg"

	"github.com/pelletier/go-toml/v2"
)

func LoadConfig(path string, value interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if strings.HasSuffix(path, ".toml") {
		return toml.Unmarshal(data, value)
	}
	return json.Unmarshal(data, value)
}

func SaveConfig(path string, value interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	var data []byte
	var err error

	if strings.HasSuffix(path, ".toml") {
		data, err = toml.Marshal(value)
	} else {
		data, err = json.MarshalIndent(value, "", "  ")
	}

	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func SavePrefixConfig(prefixName string, options types.LaunchOptions) error {
	path := GetPrefixConfigPath(prefixName)
	return SaveConfig(path, options)
}

func LoadPrefixConfig(prefixName string) (*types.LaunchOptions, error) {
	path := GetPrefixConfigPath(prefixName)
	var options types.LaunchOptions

	if err := LoadConfig(path, &options); err != nil {
		return &types.LaunchOptions{
			Extras: types.ExtrasConfig{
				Lsfg: types.LsfgConfig{
					Multiplier: "2",
				},
				Memory: types.MemoryConfig{
					Value: "4G",
				},
				Gamescope: types.GamescopeConfig{
					Width:       "1920",
					Height:      "1080",
					RefreshRate: "60",
				},
			},
		}, nil
	}
	return &options, nil
}

func SaveGameConfig(options types.LaunchOptions) error {
	path := GetGameConfigFilePath(options.GamePath)
	return SaveConfig(path, options)
}

func LoadGameConfig(executablePath string) (*types.LaunchOptions, error) {
	path := GetGameConfigFilePath(executablePath)
	var options types.LaunchOptions
	if err := LoadConfig(path, &options); err != nil {
		return nil, err
	}
	return &options, nil
}

func ListGameConfigs() ([]types.LaunchOptions, error) {
	configDirectory := GetConfigDirectory()
	if _, err := os.Stat(configDirectory); os.IsNotExist(err) {
		return make([]types.LaunchOptions, 0), nil
	}

	entries, err := os.ReadDir(configDirectory)
	if err != nil {
		return nil, err
	}

	configs := make([]types.LaunchOptions, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		configPath := filepath.Join(configDirectory, entry.Name(), "config.json")
		var options types.LaunchOptions
		if err := LoadConfig(configPath, &options); err == nil {
			configs = append(configs, options)
		}
	}

	return configs, nil
}

func SaveLsfgProfile(profile lsfg.InternalProfile) error {
	configPath := GetExecutableConfigPath(profile.GamePath)

	if err := os.MkdirAll(configPath, 0755); err != nil {
		return err
	}

	profilePath := filepath.Join(configPath, "lsfg_vk.toml")
	return SaveConfig(profilePath, profile)
}

func LoadLsfgProfile(gamePath string) (*lsfg.InternalProfile, error) {
	configPath := GetExecutableConfigPath(gamePath)

	profilePath := filepath.Join(configPath, "lsfg_vk.toml")
	var profile lsfg.InternalProfile
	if err := LoadConfig(profilePath, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}
