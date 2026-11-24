package runtimesvc

import (
	"fmt"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/drives"
)

// existingPrefixFolder expands a prefix path and checks that the prefix folder exists.
func existingPrefixFolder(prefixPath string) (string, error) {
	expanded := config.ExpandPath(prefixPath)
	if expanded == "" || !config.PathExists(expanded) {
		return "", fmt.Errorf("Prefix folder does not exist yet. Run a Wine tool in it first: %q", prefixPath)
	}
	return expanded, nil
}

func listPrefixDrives(prefixPath string) ([]drives.PrefixDrive, error) {
	folder, err := existingPrefixFolder(prefixPath)
	if err != nil {
		return nil, err
	}
	return drives.ListBindings(folder)
}
