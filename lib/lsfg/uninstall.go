package lsfg

import (
	"fmt"
	"os/exec"
	"strings"
)

var filesToRemove = []string{
	"/usr/share/vulkan/implicit_layer.d/VkLayer_LSFGVK_frame_generation.json",
	"/usr/lib/liblsfg-vk-layer.so",
	"/usr/share/icons/hicolor/256x256/apps/gay.pancake.lsfg-vk-ui.png",
	"/usr/share/applications/gay.pancake.lsfg-vk-ui.desktop",
	"/usr/bin/lsfg-vk-cli",
	"/usr/bin/lsfg-vk-ui",
}

func Uninstall(onLog func(string)) error {
	onLog("Starting LSFG-VK removal...")

	files := strings.Join(filesToRemove, " ")
	command := exec.Command("pkexec", "sh", "-c", fmt.Sprintf("rm -f %s 2>&1", files))
	output, err := command.CombinedOutput()

	if err != nil {
		onLog(fmt.Sprintf("Failed to remove files: %v", err))
		if len(output) > 0 {
			onLog(fmt.Sprintf("Details: %s", string(output)))
		}
		return fmt.Errorf("uninstall failed: %w", err)
	}

	onLog("Removal complete.")
	return nil
}
