//go:build linux

package runtimesvc

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"light-launcher/core/internal/config"
)

const drmFolder = "/sys/class/drm"

type gpuDevice struct {
	name string
	rank int
}

var (
	gpuOnce    sync.Once
	gpuDevices []gpuDevice
)

// listGpuNames returns GPU names with VGA devices first. The lookup runs once and is cached.
func listGpuNames() []string {
	gpuOnce.Do(func() {
		gpuDevices = detectGpuDevices()
	})
	names := make([]string, 0, len(gpuDevices))
	for _, device := range gpuDevices {
		names = append(names, device.name)
	}
	return names
}

func detectGpuDevices() []gpuDevice {
	devices := lspciDevices()
	if len(devices) == 0 {
		devices = drmDevices()
	}
	return devices
}

// lspciDevices lists display controllers from lspci. It returns nothing when lspci is missing.
func lspciDevices() []gpuDevice {
	output, err := exec.Command("lspci").Output()
	if err != nil {
		return nil
	}

	var devices []gpuDevice
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		lower := strings.ToLower(line)
		if !strings.Contains(lower, "vga") && !strings.Contains(lower, "3d") && !strings.Contains(lower, "display") {
			continue
		}
		name := lspciName(line)
		if name == "" {
			continue
		}
		devices = append(devices, gpuDevice{name: name, rank: lspciRank(lower)})
	}
	sort.SliceStable(devices, func(i, j int) bool {
		return devices[i].rank < devices[j].rank
	})
	return devices
}

// lspciName returns the text after the PCI address and the first ": ", or the whole line without it.
func lspciName(line string) string {
	if index := strings.Index(line, ": "); index != -1 {
		return strings.TrimSpace(line[index+2:])
	}
	return strings.TrimSpace(line)
}

func lspciRank(lowerLine string) int {
	switch {
	case strings.Contains(lowerLine, "vga"):
		return 0
	case strings.Contains(lowerLine, "3d"):
		return 1
	default:
		return 2
	}
}

// drmDevices reads GPU names from /sys/class/drm when lspci finds nothing.
func drmDevices() []gpuDevice {
	entries, err := os.ReadDir(drmFolder)
	if err != nil {
		return nil
	}

	var devices []gpuDevice
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "card") || strings.Contains(entry.Name(), "-") {
			continue
		}
		devicePath := filepath.Join(drmFolder, entry.Name(), "device")
		if !config.PathExists(devicePath) {
			continue
		}
		name := drmProductName(devicePath)
		if name == "" {
			name = entry.Name()
		}
		devices = append(devices, gpuDevice{name: name, rank: 0})
	}
	return devices
}

func drmProductName(devicePath string) string {
	for _, file := range []string{"product_name", "product"} {
		content, err := os.ReadFile(filepath.Join(devicePath, file))
		if err != nil {
			continue
		}
		if value := strings.TrimSpace(string(content)); value != "" {
			return value
		}
	}
	return ""
}
