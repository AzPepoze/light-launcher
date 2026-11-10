package systemsvc

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// busUtilization is one row of nvidia-smi output. The bus id is lowercase.
type busUtilization struct {
	bus         string
	utilization float64
}

// gpuUsages returns a 0-100 busy percentage for each device, in the same order.
func (s *Service) gpuUsages(ctx context.Context, devices []gpuDevice) []float64 {
	usages := make([]float64, 0, len(devices))
	if len(devices) == 0 {
		return usages
	}
	nvidia := s.nvidiaUtilization(ctx)
	cards := s.drmCards()
	for _, gpu := range devices {
		usages = append(usages, gpuUsage(gpu, nvidia, cards))
	}
	return usages
}

// gpuUsage prefers nvidia-smi, then sysfs. A GPU with neither reports 0.
func gpuUsage(gpu gpuDevice, nvidia []busUtilization, cards []string) float64 {
	if util, ok := nvidiaUsage(gpu.PCI, nvidia); ok {
		return util
	}
	if util, ok := drmUsage(gpu.PCI, cards); ok {
		return util
	}
	return 0
}

// nvidiaUtilization returns nil when nvidia-smi is missing or fails.
func (s *Service) nvidiaUtilization(ctx context.Context) []busUtilization {
	out, err := s.runner.Output(ctx, "nvidia-smi",
		"--query-gpu=pci.bus_id,utilization.gpu", "--format=csv,noheader,nounits")
	if err != nil {
		return nil
	}
	return parseNvidiaUtilization(string(out))
}

func parseNvidiaUtilization(output string) []busUtilization {
	var rows []busUtilization
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			continue
		}
		util, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			continue
		}
		bus := strings.ToLower(strings.TrimSpace(parts[0]))
		rows = append(rows, busUtilization{bus: bus, utilization: util})
	}
	return rows
}

func nvidiaUsage(pci string, rows []busUtilization) (float64, bool) {
	if pci == "" {
		return 0, false
	}
	want := strings.ToLower(pci)
	for _, row := range rows {
		if strings.Contains(row.bus, want) {
			return row.utilization, true
		}
	}
	return 0, false
}

// drmUsage reads the first card that belongs to pci. With an unknown pci any card matches.
func drmUsage(pci string, cards []string) (float64, bool) {
	for _, card := range cards {
		if !cardMatchesPCI(card, pci) {
			continue
		}
		if util, ok := gpuBusyPercent(card); ok {
			return util, true
		}
		if util, ok := frequencyPercent(card); ok {
			return util, true
		}
	}
	return 0, false
}

// cardMatchesPCI checks that the card's device symlink points at the PCI address.
func cardMatchesPCI(card, pci string) bool {
	if pci == "" {
		return true
	}
	devicePath := filepath.Join(card, "device")
	if !pathExists(devicePath) {
		return false
	}
	target, err := os.Readlink(devicePath)
	return err == nil && strings.Contains(target, pci)
}

// gpuBusyPercent reads gpu_busy_percent, exposed by amdgpu and some other drivers.
func gpuBusyPercent(card string) (float64, bool) {
	value, ok := readInt(filepath.Join(card, "device", "gpu_busy_percent"))
	if !ok {
		return 0, false
	}
	return clamp(float64(value), 0, 100), true
}

// frequencyPercent estimates load from the Intel frequency files: (act - min) / (max - min).
func frequencyPercent(card string) (float64, bool) {
	act, actOK := readInt(filepath.Join(card, "gt_act_freq_mhz"))
	maxFreq, maxOK := readInt(filepath.Join(card, "gt_max_freq_mhz"))
	if !actOK || !maxOK {
		return 0, false
	}
	minFreq, ok := minFrequency(card)
	if !ok {
		return 0, false
	}
	if maxFreq > minFreq && act > minFreq {
		return math.Round(float64(act-minFreq) / float64(maxFreq-minFreq) * 100), true
	}
	return 0, true
}

// minFrequency returns 0 when the min file is absent, and false when it exists but cannot be read.
func minFrequency(card string) (int64, bool) {
	minPath := filepath.Join(card, "gt_min_freq_mhz")
	if !pathExists(minPath) {
		return 0, true
	}
	return readInt(minPath)
}
