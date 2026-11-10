package systemsvc

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const drmDir = "/sys/class/drm"

var pciAddressPattern = regexp.MustCompile(`^([0-9a-fA-F:.]+)\s+`)

// detectGpus lists GPUs from lspci, or from DRM cards when lspci finds none.
func (s *Service) detectGpus(ctx context.Context) []gpuDevice {
	if gpus := s.lspciGpus(ctx); len(gpus) > 0 {
		return gpus
	}
	return s.drmGpus()
}

func (s *Service) lspciGpus(ctx context.Context) []gpuDevice {
	out, err := s.runner.Output(ctx, "lspci")
	if err != nil {
		return nil
	}
	return parseLspci(string(out))
}

// parseLspci keeps display controller lines and sorts them by rank, keeping the lspci order within a rank.
func parseLspci(output string) []gpuDevice {
	var gpus []gpuDevice
	for _, line := range strings.Split(output, "\n") {
		if !isDisplayLine(line) {
			continue
		}
		if gpu := gpuFromLspciLine(line); gpu.Name != "" {
			gpus = append(gpus, gpu)
		}
	}
	sort.SliceStable(gpus, func(i, j int) bool { return gpus[i].Rank < gpus[j].Rank })
	return gpus
}

func isDisplayLine(line string) bool {
	lower := strings.ToLower(line)
	return strings.Contains(lower, "vga") || strings.Contains(lower, "3d") || strings.Contains(lower, "display")
}

// gpuFromLspciLine reads a line such as "01:00.0 VGA compatible controller: NVIDIA ...".
func gpuFromLspciLine(line string) gpuDevice {
	gpu := gpuDevice{Name: line, Rank: displayRank(line)}
	if match := pciAddressPattern.FindStringSubmatch(line); match != nil {
		gpu.PCI = match[1]
	}
	if index := strings.Index(line, ": "); index != -1 {
		gpu.Name = strings.TrimSpace(line[index+2:])
	}
	return gpu
}

func displayRank(line string) int {
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "vga"):
		return 0
	case strings.Contains(lower, "3d"):
		return 1
	default:
		return 2
	}
}

// drmGpus is the fallback: one device per DRM card, named by its product file when present.
func (s *Service) drmGpus() []gpuDevice {
	var gpus []gpuDevice
	for _, card := range s.drmCards() {
		devicePath := filepath.Join(card, "device")
		if !pathExists(devicePath) {
			continue
		}
		name := readTrimmed(filepath.Join(devicePath, "product_name"))
		if name == "" {
			name = readTrimmed(filepath.Join(devicePath, "product"))
		}
		if name == "" {
			name = filepath.Base(card)
		}
		gpus = append(gpus, gpuDevice{Name: name})
	}
	return gpus
}

// drmCards returns the card directories in sysfs. Connector entries such as card0-HDMI-A-1 are skipped.
func (s *Service) drmCards() []string {
	dir := s.path(drmDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var cards []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "card") && !strings.Contains(name, "-") {
			cards = append(cards, filepath.Join(dir, name))
		}
	}
	return cards
}
