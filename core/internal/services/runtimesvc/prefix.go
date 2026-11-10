package runtimesvc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultPrefixName = "Default"

// prefixBaseFolder returns the folder that holds all prefixes, honoring the custom prefix folder setting.
func prefixBaseFolder() string {
	return prefixBaseDir(loadAppSettings().CustomPrefixDir)
}

// listPrefixes returns the prefix folder names, with "Default" always first.
func listPrefixes() ([]string, error) {
	baseDir := prefixBaseFolder()
	if !pathExists(baseDir) {
		return []string{defaultPrefixName}, nil
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return []string{defaultPrefixName}, nil
	}

	prefixes := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			prefixes = append(prefixes, entry.Name())
		}
	}
	if !contains(prefixes, defaultPrefixName) {
		prefixes = append([]string{defaultPrefixName}, prefixes...)
	}
	return prefixes, nil
}

// createPrefix makes a new prefix folder. The name is trimmed first.
func createPrefix(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("Prefix name cannot be empty")
	}
	prefixPath := filepath.Join(prefixBaseFolder(), strings.TrimSpace(name))
	return os.MkdirAll(prefixPath, 0755)
}

// removePrefix deletes a prefix folder and everything in it. The Default prefix cannot be removed.
func removePrefix(name string) error {
	if name == defaultPrefixName {
		return errors.New("Cannot delete Default prefix")
	}
	if err := checkPrefixName(name); err != nil {
		return err
	}
	prefixPath := filepath.Join(prefixBaseFolder(), name)
	if !pathExists(prefixPath) {
		return nil
	}
	return os.RemoveAll(prefixPath)
}

// checkPrefixName rejects names that would point outside the prefix folder, such as "" or "../x".
// The Electron app did not check this, so deleting with an empty name removed every prefix.
func checkPrefixName(name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("Invalid prefix name: %q", name)
	}
	return nil
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
