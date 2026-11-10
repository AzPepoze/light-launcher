package systemsvc

import (
	"context"
	"sync"
)

// gpuDevice is one GPU found by the platform. Rank orders VGA, then 3D, then other display devices.
type gpuDevice struct {
	Name string
	PCI  string // empty when the bus address is unknown
	Rank int
}

// gpuCache keeps the detected devices for the life of the service.
type gpuCache struct {
	mu      sync.Mutex
	devices []gpuDevice
	filled  bool
}

func (c *gpuCache) get(detect func() []gpuDevice) []gpuDevice {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.filled {
		c.devices = detect()
		c.filled = true
	}
	return c.devices
}

func (s *Service) gpuDevices(ctx context.Context) []gpuDevice {
	return s.gpus.get(func() []gpuDevice { return s.detectGpus(ctx) })
}

func (s *Service) listGpus(ctx context.Context) []string {
	devices := s.gpuDevices(ctx)
	names := make([]string, 0, len(devices))
	for _, device := range devices {
		names = append(names, device.Name)
	}
	return names
}
