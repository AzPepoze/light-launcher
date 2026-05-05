package system

import (
	"bufio"
	"fmt"
	"light-launcher/internal/types"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
	lastTotal int64
	lastIdle  int64
)

func GetSystemUsage() types.SystemUsage {
	usage := types.SystemUsage{CPU: "0%", RAM: "0%", GPU: "0%"}

	if file, err := os.Open("/proc/stat"); err == nil {
		scanner := bufio.NewScanner(file)
		if scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				var total int64
				for i := 1; i < len(fields); i++ {
					val, _ := strconv.ParseInt(fields[i], 10, 64)
					total += val
				}
				idle, _ := strconv.ParseInt(fields[4], 10, 64)

				if lastTotal > 0 {
					totalDelta := total - lastTotal
					idleDelta := idle - lastIdle
					if totalDelta > 0 {
						cpuPercent := 100 * (float64(totalDelta) - float64(idleDelta)) / float64(totalDelta)
						usage.CPU = fmt.Sprintf("%.1f%%", cpuPercent)
					}
				}
				lastTotal = total
				lastIdle = idle
			}
		}
		file.Close()
	}

	if file, err := os.Open("/proc/meminfo"); err == nil {
		scanner := bufio.NewScanner(file)
		var total, available int64
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "MemTotal:") {
				fmt.Sscanf(line, "MemTotal: %d kB", &total)
			} else if strings.HasPrefix(line, "MemAvailable:") {
				fmt.Sscanf(line, "MemAvailable: %d kB", &available)
			}
			if total > 0 && available > 0 {
				break
			}
		}
		file.Close()

		if total > 0 {
			used := total - available
			usage.RAM = fmt.Sprintf("%.1f GB / %d GB (%.0f%%)", float64(used)/1024/1024, total/1024/1024, 100*float64(used)/float64(total))
		}
	}

	usage.GPU = GetGpuUsage()

	return usage
}

func GetSystemToolsStatus() types.SystemToolsStatus {
	return types.SystemToolsStatus{
		HasGamescope:  IsCommandAvailable("gamescope"),
		HasMangoHud:   IsCommandAvailable("mangohud"),
		HasGameMode:    IsCommandAvailable("gamemoderun"),
		HasVulkanInfo: IsCommandAvailable("vulkaninfo"),
	}
}

func GetSystemInfo() types.SystemInfo {
	info := types.SystemInfo{
		OS:     "Unknown",
		Kernel: "Unknown",
		CPU:    "Unknown",
		GPU:    "Unknown",
		RAM:    "Unknown",
		Driver: "Unknown",
	}

	if file, err := os.Open("/etc/os-release"); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				info.OS = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
				break
			}
		}
		file.Close()
	}

	if output, err := exec.Command("uname", "-r").Output(); err == nil {
		info.Kernel = strings.TrimSpace(string(output))
	}

	if output, err := exec.Command("sh", "-c", "lscpu | grep 'Model name' | cut -d':' -f2 | xargs").Output(); err == nil {
		info.CPU = strings.TrimSpace(string(output))
	}

	gpus := GetListGpus()
	if len(gpus) > 0 {
		info.GPU = gpus[0]
	}

	if file, err := os.Open("/proc/meminfo"); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var memTotalKb int64
			line := scanner.Text()
			if strings.HasPrefix(line, "MemTotal:") {
				fmt.Sscanf(line, "MemTotal: %d kB", &memTotalKb)
				info.RAM = fmt.Sprintf("%d GB", memTotalKb/1024/1024)
				break
			}
		}
		file.Close()
	}

	if output, err := exec.Command("sh", "-c", "vulkaninfo --summary | grep -m 1 'driverVersion' | awk '{print $3}'").Output(); err == nil {
		info.Driver = strings.TrimSpace(string(output))
	}

	return info
}

func DropCaches() error {
	_ = exec.Command("sync").Run()
	command := exec.Command("pkexec", "sysctl", "-w", "vm.drop_caches=3")
	return command.Run()
}

func ClearSwap() error {
	command := exec.Command("pkexec", "sh", "-c", "swapoff -a && swapon -a")
	return command.Run()
}

func IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
