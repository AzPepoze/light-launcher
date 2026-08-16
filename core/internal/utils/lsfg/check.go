package lsfg

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	ManifestPath = "/usr/share/vulkan/implicit_layer.d/VkLayer_LSFGVK_frame_generation.json"
)

func IsInstalled() bool {
	_, err := os.Stat(ManifestPath)
	return err == nil
}

func GetVersion() string {
	data, err := os.ReadFile(ManifestPath)
	if err != nil {
		return "unknown"
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return "unknown"
	}

	if meta, ok := manifest["manifest_version"]; ok {
		return fmt.Sprintf("%v", meta)
	}

	return "installed"
}

func GetStatus() Status {
	return Status{
		IsInstalled: IsInstalled(),
		Version:     GetVersion(),
	}
}
