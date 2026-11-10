package systemsvc

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

const (
	windowsVersionKey = `SOFTWARE\Microsoft\Windows NT\CurrentVersion`
	cpuKey            = `HARDWARE\DESCRIPTION\System\CentralProcessor\0`
	// windows11MinBuild is the first build number of Windows 11, which still reports "Windows 10" in ProductName.
	windows11MinBuild = 22000
)

// readCPUTimes returns all CPU time and idle time, the two numbers the shared CPU sampler needs.
func (s *Service) readCPUTimes() (float64, float64, bool) {
	idle, kernel, user, ok := systemTimes()
	if !ok {
		return 0, 0, false
	}
	return float64(kernel + user), float64(idle), true
}

func (s *Service) readMemory() memInfo {
	total, avail, ok := globalMemory()
	if !ok {
		return memInfo{}
	}
	return memInfo{totalKB: int64(total / 1024), availableKB: int64(avail / 1024)}
}

func (s *Service) readOSName() string {
	name := readRegistryString(windowsVersionKey, "ProductName")
	if name == "" {
		return "Unknown"
	}
	return windowsDisplayName(name, readRegistryString(windowsVersionKey, "CurrentBuildNumber"))
}

// windowsDisplayName fixes the ProductName of Windows 11, which Microsoft never renamed in the registry.
func windowsDisplayName(productName, build string) string {
	buildNumber, err := strconv.Atoi(build)
	if err != nil || buildNumber < windows11MinBuild {
		return productName
	}
	return strings.Replace(productName, "Windows 10", "Windows 11", 1)
}

func (s *Service) readKernel(_ context.Context) string {
	build := readRegistryString(windowsVersionKey, "CurrentBuildNumber")
	if build == "" {
		return "Unknown"
	}
	return "10.0." + build
}

// readCPUName returns the processor name with the logical thread count from the runtime.
func (s *Service) readCPUName(_ context.Context) string {
	name := strings.TrimSpace(readRegistryString(cpuKey, "ProcessorNameString"))
	if name == "" {
		return "Unknown"
	}
	return fmt.Sprintf("%s (%d threads)", name, runtime.NumCPU())
}

// readDriver reports Unknown: no GPU driver lookup exists on Windows yet.
func (s *Service) readDriver(_ context.Context) string {
	return "Unknown"
}

// detectGpus returns nil: GPU enumeration is not implemented on Windows yet.
func (s *Service) detectGpus(_ context.Context) []gpuDevice {
	return nil
}

func (s *Service) gpuUsages(_ context.Context, _ []gpuDevice) []float64 {
	return []float64{}
}
