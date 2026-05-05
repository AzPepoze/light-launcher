package config

import (
	"os"
	"path/filepath"
)

func ListPrefixes() ([]string, error) {
	baseDirectory := GetPrefixBaseDirectory()
	if err := os.MkdirAll(baseDirectory, 0755); err != nil {
		return []string{"Default"}, nil
	}
	entries, err := os.ReadDir(baseDirectory)
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
		_ = os.MkdirAll(filepath.Join(baseDirectory, "Default"), 0755)
		prefixes = append(prefixes, "Default")
	}
	return prefixes, nil
}

func CreatePrefix(name string) error {
	path := filepath.Join(GetPrefixBaseDirectory(), name)
	return os.MkdirAll(path, 0755)
}

func RemovePrefix(name string) error {
	if name == "" || name == "Default" {
		return nil
	}
	path := filepath.Join(GetPrefixBaseDirectory(), name)
	return os.RemoveAll(path)
}
