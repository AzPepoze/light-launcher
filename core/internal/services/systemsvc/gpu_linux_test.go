package systemsvc

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const nvidiaQuery = "nvidia-smi --query-gpu=pci.bus_id,utilization.gpu --format=csv,noheader,nounits"

func TestParseLspciKeepsDisplayLinesSortedByRank(t *testing.T) {
	output := "00:02.0 VGA compatible controller: Intel Corporation UHD\n" +
		"00:14.0 USB controller: Intel Corporation Filtered\n" +
		"02:00.0 Display controller: Other Display\n" +
		"01:00.0 3D controller: NVIDIA Corporation GA107M\n"

	want := []gpuDevice{
		{Name: "Intel Corporation UHD", PCI: "00:02.0", Rank: 0},
		{Name: "NVIDIA Corporation GA107M", PCI: "01:00.0", Rank: 1},
		{Name: "Other Display", PCI: "02:00.0", Rank: 2},
	}
	if got := parseLspci(output); !reflect.DeepEqual(got, want) {
		t.Fatalf("parseLspci = %+v, want %+v", got, want)
	}
}

func TestDetectGpusFallsBackToDrmCards(t *testing.T) {
	runner := newFakeRunner().fail("lspci")
	s := newTestService(t, runner)
	writeHostFile(t, s, "/sys/class/drm/card0/device/product_name", "Radeon Demo\n")
	writeHostFile(t, s, "/sys/class/drm/card0-HDMI-A-1/status", "connected\n")
	if err := os.MkdirAll(s.path("/sys/class/drm/card1"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeHostFile(t, s, "/sys/class/drm/card2/device/placeholder", "")

	names := make([]string, 0)
	for _, gpu := range s.detectGpus(context.Background()) {
		names = append(names, gpu.Name)
	}
	want := []string{"Radeon Demo", "card2"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("detectGpus names = %v, want %v", names, want)
	}
}

func TestNvidiaUsageMatchesBusId(t *testing.T) {
	runner := newFakeRunner().on(nvidiaQuery, "00000000:01:00.0, 37\n")
	s := newTestService(t, runner)
	devices := []gpuDevice{{Name: "NVIDIA", PCI: "01:00.0"}}

	usages := s.gpuUsages(context.Background(), devices)
	if !reflect.DeepEqual(usages, []float64{37}) {
		t.Fatalf("gpuUsages = %v, want [37]", usages)
	}
}

func TestParseNvidiaUtilizationSkipsNotAvailableRows(t *testing.T) {
	rows := parseNvidiaUtilization("00000000:01:00.0, [N/A]\n00000000:02:00.0, 12\n")

	want := []busUtilization{{bus: "00000000:02:00.0", utilization: 12}}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("parseNvidiaUtilization = %+v, want %+v", rows, want)
	}
}

func TestGpuUsagesReadsDrmBusyPercentThroughSymlink(t *testing.T) {
	s := newTestService(t, newFakeRunner())
	// card0/device is a symlink to a PCI directory whose name contains the bus address.
	writeHostFile(t, s, "/sys/devices/pci/0000:00:02.0/gpu_busy_percent", "42\n")
	if err := os.MkdirAll(s.path("/sys/class/drm/card0"), 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join("..", "..", "..", "devices", "pci", "0000:00:02.0")
	if err := os.Symlink(target, s.path("/sys/class/drm/card0/device")); err != nil {
		t.Fatal(err)
	}

	usages := s.gpuUsages(context.Background(), []gpuDevice{{Name: "Intel", PCI: "00:02.0"}})
	if !reflect.DeepEqual(usages, []float64{42}) {
		t.Fatalf("gpuUsages = %v, want [42]", usages)
	}
}

func TestFrequencyPercentFromIntelCounters(t *testing.T) {
	s := newTestService(t, newFakeRunner())
	card := s.path("/sys/class/drm/card0")
	writeHostFile(t, s, "/sys/class/drm/card0/gt_act_freq_mhz", "500\n")
	writeHostFile(t, s, "/sys/class/drm/card0/gt_max_freq_mhz", "1000\n")
	writeHostFile(t, s, "/sys/class/drm/card0/gt_min_freq_mhz", "100\n")

	got, ok := frequencyPercent(card)
	// (500 - 100) / (1000 - 100) = 44.4, rounded.
	if !ok || got != 44 {
		t.Fatalf("frequencyPercent = (%v, %v), want (44, true)", got, ok)
	}
}

func TestGpuUsagesWithoutDevicesIsEmptyArray(t *testing.T) {
	s := newTestService(t, newFakeRunner())
	usages := s.gpuUsages(context.Background(), nil)
	if usages == nil || len(usages) != 0 {
		t.Fatalf("gpuUsages(nil) = %#v, want empty non-nil slice", usages)
	}
}
