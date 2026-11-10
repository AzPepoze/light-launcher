package systemsvc

import (
	"context"
	"fmt"
	"math"

	"light-launcher/core/internal/types"
)

func (s *Service) systemInfo(ctx context.Context) types.SystemInfo {
	info := types.SystemInfo{
		OS:     s.readOSName(),
		Kernel: s.readKernel(ctx),
		CPU:    s.readCPUName(ctx),
		GPU:    "Unknown",
		RAM:    "Unknown",
		Driver: s.readDriver(ctx),
	}

	gpus := s.listGpus(ctx)
	if len(gpus) > 0 {
		info.GPU = gpus[0]
		info.Gpus = gpus
	}

	if mem := s.readMemory(); mem.totalKB > 0 {
		info.RAM = fmt.Sprintf("%d GB", int64(math.Round(float64(mem.totalKB)/kbPerGb)))
	}
	return info
}
