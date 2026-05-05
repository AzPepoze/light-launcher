package system

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetShaderCacheSize() string {
	homeDirectory, _ := os.UserHomeDir()
	paths := []string{
		filepath.Join(homeDirectory, ".steam/root/steamapps/shadercache"),
		filepath.Join(homeDirectory, ".local/share/Steam/steamapps/shadercache"),
	}

	var totalSize int64
	for _, path := range paths {
		filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				totalSize += info.Size()
			}
			return nil
		})
	}

	if totalSize == 0 {
		return "0 MB"
	}

	return fmt.Sprintf("%.1f GB", float64(totalSize)/(1024*1024*1024))
}

func ClearShaderCache() error {
	homeDirectory, _ := os.UserHomeDir()
	paths := []string{
		filepath.Join(homeDirectory, ".steam/root/steamapps/shadercache"),
		filepath.Join(homeDirectory, ".local/share/Steam/steamapps/shadercache"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			entries, _ := os.ReadDir(path)
			for _, entry := range entries {
				os.RemoveAll(filepath.Join(path, entry.Name()))
			}
		}
	}
	return nil
}
