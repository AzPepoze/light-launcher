package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"light-launcher/internal/types"
	"light-launcher/lib/lsfg"

	"github.com/pelletier/go-toml/v2"
)

func GenerateID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

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
	if options.ID == "" {
		options.ID = GenerateID()
	}

	path := GetGameConfigFilePath(options.Name, options.ID)
	
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	return SaveConfig(path, options)
}

func LoadGameConfigByID(name string, id string) (*types.LaunchOptions, error) {
	path := GetGameConfigFilePath(name, id)
	var options types.LaunchOptions
	if err := LoadConfig(path, &options); err != nil {
		return nil, err
	}
	return &options, nil
}

func LoadGameConfig(executablePath string) (*types.LaunchOptions, error) {
	configs, err := ListGameConfigs()
	if err != nil {
		return nil, err
	}

	cleanPath := filepath.Clean(executablePath)
	if absPath, err := filepath.Abs(cleanPath); err == nil {
		cleanPath = absPath
	}

	for _, cfg := range configs {
		cfgGamePath := filepath.Clean(cfg.GamePath)
		if abs, err := filepath.Abs(cfgGamePath); err == nil {
			cfgGamePath = abs
		}

		cfgRunnerPath := filepath.Clean(cfg.RunnerPath)
		if abs, err := filepath.Abs(cfgRunnerPath); err == nil {
			cfgRunnerPath = abs
		}

		if cfgGamePath == cleanPath || cfgRunnerPath == cleanPath {
			return &cfg, nil
		}
	}

	return nil, fmt.Errorf("config not found for path: %s", executablePath)
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

func SaveLsfgProfile(name string, id string, profile lsfg.InternalProfile) error {
	configPath := GetExecutableConfigPath(name, id)

	if err := os.MkdirAll(configPath, 0755); err != nil {
		return err
	}

	profilePath := filepath.Join(configPath, "lsfg_vk.toml")
	return SaveConfig(profilePath, profile)
}

func LoadLsfgProfile(name string, id string) (*lsfg.InternalProfile, error) {
	configPath := GetExecutableConfigPath(name, id)

	profilePath := filepath.Join(configPath, "lsfg_vk.toml")
	var profile lsfg.InternalProfile
	if err := LoadConfig(profilePath, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func GetAppSettingsPath() string {
	return filepath.Join(GetBaseDirectory(), "settings.json")
}

func LoadAppSettings() *types.AppSettings {
	path := GetAppSettingsPath()
	var settings types.AppSettings
	
	if err := LoadConfig(path, &settings); err != nil {
		return &types.AppSettings{
			TransparentMode: true,
		}
	}
	return &settings
}

func SaveAppSettings(settings types.AppSettings) error {
	path := GetAppSettingsPath()
	return SaveConfig(path, settings)
}
