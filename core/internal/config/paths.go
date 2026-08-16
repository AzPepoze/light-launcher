package config

import (
	"os"
	"path/filepath"
)

func GetBaseDirectory() string {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return AppName
	}
	return filepath.Join(homeDirectory, AppName)
}

func GetConfigDirectory() string {
	return filepath.Join(GetBaseDirectory(), ConfigDirName)
}

func GetPrefixBaseDirectory() string {
	return filepath.Join(GetBaseDirectory(), PrefixesDirName)
}

func GetExecutableConfigPath(name string, id string) string {
	if id == "" {
		return filepath.Join(GetConfigDirectory(), name)
	}

	return filepath.Join(GetConfigDirectory(), id)
}

func GetGameConfigFilePath(name string, id string) string {
	return filepath.Join(GetExecutableConfigPath(name, id), "config.json")
}

func GetGameLsfgConfigPath(name string, id string) string {
	return filepath.Join(GetExecutableConfigPath(name, id), "lsfg_vk.toml")
}

func GetPrefixConfigPath(prefixName string) string {
	return filepath.Join(GetPrefixBaseDirectory(), prefixName, "light-launcher.json")
}

func GetLogsDirectory() string {
	return filepath.Join(GetBaseDirectory(), LogsDirName)
}
