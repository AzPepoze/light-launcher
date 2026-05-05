package config

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"path/filepath"
)

func GetBaseDirectory() string {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return "LightLauncher"
	}
	return filepath.Join(homeDirectory, "LightLauncher")
}

func GetConfigDirectory() string {
	return filepath.Join(GetBaseDirectory(), "config", "executables")
}

func GetPrefixBaseDirectory() string {
	return filepath.Join(GetBaseDirectory(), "prefixes")
}

func GetExecutableConfigPath(executablePath string) string {
	hasher := sha1.New()
	hasher.Write([]byte(executablePath))
	hash := hex.EncodeToString(hasher.Sum(nil))[:8]

	baseName := filepath.Base(executablePath)
	extension := filepath.Ext(baseName)
	baseName = baseName[:len(baseName)-len(extension)]

	folderName := baseName + "-" + hash
	return filepath.Join(GetConfigDirectory(), folderName)
}

func GetGameConfigFilePath(executablePath string) string {
	return filepath.Join(GetExecutableConfigPath(executablePath), "config.json")
}

func GetGameLsfgConfigPath(executablePath string) string {
	return filepath.Join(GetExecutableConfigPath(executablePath), "lsfg_vk.toml")
}

func GetPrefixConfigPath(prefixName string) string {
	return filepath.Join(GetPrefixBaseDirectory(), prefixName, "light-launcher.json")
}
