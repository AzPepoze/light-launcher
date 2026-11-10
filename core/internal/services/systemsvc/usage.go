package systemsvc

import (
	"context"
	"math"
	"sync"

	"light-launcher/core/internal/types"
)

const kbPerGb = 1024 * 1024

// memInfo holds RAM figures in kilobytes, the unit /proc/meminfo uses.
type memInfo struct {
	totalKB     int64
	availableKB int64
}

// cpuSampler turns cumulative CPU times into a busy percentage between two calls.
type cpuSampler struct {
	mu        sync.Mutex
	lastTotal float64
	lastIdle  float64
}

// busyPercent returns the busy share since the previous call, or 0 on the first call.
func (c *cpuSampler) busyPercent(total, idle float64) float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	busy := 0.0
	if c.lastTotal > 0 {
		busy = busyShare(total-c.lastTotal, idle-c.lastIdle)
	}
	c.lastTotal = total
	c.lastIdle = idle
	return busy
}

func busyShare(totalDelta, idleDelta float64) float64 {
	if totalDelta <= 0 {
		return 0
	}
	return clamp(100*(totalDelta-idleDelta)/totalDelta, 0, 100)
}

func clamp(value, low, high float64) float64 {
	return min(high, max(low, value))
}

// applyMemory fills the RAM fields: used and total in GB, and the used percentage.
func applyMemory(usage *types.SystemUsage, mem memInfo) {
	if mem.totalKB <= 0 {
		return
	}
	usedKB := mem.totalKB - mem.availableKB
	usage.RamUsedGb = math.Round(float64(usedKB)/kbPerGb*10) / 10
	usage.RamTotalGb = math.Round(float64(mem.totalKB) / kbPerGb)
	usage.RAM = clamp(math.Round(float64(usedKB)/float64(mem.totalKB)*100), 0, 100)
}

func (s *Service) systemUsage(ctx context.Context) types.SystemUsage {
	usage := types.SystemUsage{GPUs: []float64{}}
	if total, idle, ok := s.readCPUTimes(); ok {
		usage.CPU = s.cpu.busyPercent(total, idle)
	}
	applyMemory(&usage, s.readMemory())
	usage.GPUs = s.gpuUsages(ctx, s.gpuDevices(ctx))
	return usage
}
