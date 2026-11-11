package lsfg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// globalConfigPath returns the settings file written by the lsfg-vk tool.
func globalConfigPath() (string, error) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDirectory, ".config", "lsfg-vk", "conf.toml"), nil
}

func findProfileForGameAtPath(gamePath, path string) (*ConfigProfile, int, error) {
	data, err := os.ReadFile(path)
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

// FindProfileForGame returns the LSFG profile that applies to the game, or an error when there is none.
// It reads the tool's global conf.toml; runtimesvc writes per-game profiles, so the readers stay separate.
func FindProfileForGame(gamePath string) (*ConfigProfile, int, error) {
	path, err := globalConfigPath()
	if err != nil {
		return nil, -1, err
	}
	return findProfileForGameAtPath(gamePath, path)
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
