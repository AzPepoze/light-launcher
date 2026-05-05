package system

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func GetListGpus() []string {
	gpus := []string{}
	detected := make(map[string]bool)

	if vulkanGpus := detectVulkanGpus(); len(vulkanGpus) > 0 {
		for _, gpu := range vulkanGpus {
			if !detected[gpu] {
				gpus = append(gpus, gpu)
				detected[gpu] = true
			}
		}
	}

	return gpus
}

func GetGpuUsage() string {
	if output, err := exec.Command("nvidia-smi", "--query-gpu=utilization.gpu", "--format=csv,noheader,nounits").Output(); err == nil {
		return strings.TrimSpace(string(output)) + "%"
	}

	for index := 0; index <= 5; index++ {
		path := fmt.Sprintf("/sys/class/drm/card%d/device/gpu_busy_percent", index)
		if data, err := os.ReadFile(path); err == nil {
			value := strings.TrimSpace(string(data))
			if value != "" && value != "0" {
				return value + "%"
			}
		}
		intelPath := fmt.Sprintf("/sys/class/drm/card%d/device/i915_gpu_busy100", index)
		if data, err := os.ReadFile(intelPath); err == nil {
			value := strings.TrimSpace(string(data))
			if value != "" && value != "0" {
				return value + "%"
			}
		}
	}
	return "0%"
}

func detectVulkanGpus() []string {
	command := exec.Command("vulkaninfo")
	output, err := command.CombinedOutput()
	if err != nil {
		return nil
	}

	var gpus []string
	lines := strings.Split(string(output), "\n")
	gpuPattern := regexp.MustCompile(`GPU\s+id\s*=\s*\d+\s*\((.+)\)`)

	for _, line := range lines {
		matches := gpuPattern.FindStringSubmatch(line)
		if len(matches) >= 2 {
			gpu := strings.TrimSpace(matches[1])
			if len(gpu) > 0 && !contains(gpus, gpu) {
				gpus = append(gpus, gpu)
			}
		}
	}
	return gpus
}

func contains(slice []string, search string) bool {
	for _, item := range slice {
		if item == search {
			return true
		}
	}
	return false
}
