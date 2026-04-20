package core

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"path/filepath"
)

func GetBaseDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "LightLauncher"
	}
	return filepath.Join(home, "LightLauncher")
}

func GetConfigDir() string {
	return filepath.Join(GetBaseDir(), "config", "executables")
}

func GetPrefixBaseDir() string {
	return filepath.Join(GetBaseDir(), "prefixes")
}

func GetConfigPath(exePath string) string {
	h := sha1.New()
	h.Write([]byte(exePath))
	hash := hex.EncodeToString(h.Sum(nil))[:8]
	// Remove .exe extension
	baseName := filepath.Base(exePath)
	baseName = filepath.Base(baseName[:len(baseName)-len(filepath.Ext(baseName))])

	folderName := baseName + "-" + hash
	return filepath.Join(GetConfigDir(), folderName)
}

// GetGameConfigFile returns the path to the game's config.json file
func GetGameConfigFile(exePath string) string {
	return filepath.Join(GetConfigPath(exePath), "config.json")
}

// GetGameLsfgConfigPath returns the path to the game's LSFG lsfg_vk.toml file
func GetGameLsfgConfigPath(exePath string) string {
	return filepath.Join(GetConfigPath(exePath), "lsfg_vk.toml")
}

func GetPrefixConfigPath(prefixName string) string {
	return filepath.Join(GetPrefixBaseDir(), prefixName, "light-launcher.json")
}

func ListPrefixes() ([]string, error) {
	base := GetPrefixBaseDir()
	if err := os.MkdirAll(base, 0755); err != nil {
		return []string{"Default"}, nil
	}
	entries, err := os.ReadDir(base)
	if err != nil {
		return []string{"Default"}, nil
	}

	var prefixes []string
	for _, entry := range entries {
		if entry.IsDir() {
			prefixes = append(prefixes, entry.Name())
		}
	}

	if len(prefixes) == 0 {
		_ = os.MkdirAll(filepath.Join(base, "Default"), 0755)
		prefixes = append(prefixes, "Default")
	}
	return prefixes, nil
}

func CreatePrefix(name string) error {
	path := filepath.Join(GetPrefixBaseDir(), name)
	return os.MkdirAll(path, 0755)
}
