package systemsvc

import (
	"context"
	"reflect"
	"testing"
)

func TestParseCPUTimesSumsAggregateLine(t *testing.T) {
	content := "cpu  100 0 50 800 10 0 0 0 0 0\ncpu0 50 0 25 400 5 0 0 0 0 0\n"

	total, idle, ok := parseCPUTimes(content)
	if !ok || total != 960 || idle != 800 {
		t.Fatalf("parseCPUTimes = (%v, %v, %v), want (960, 800, true)", total, idle, ok)
	}
}

func TestParseCPUTimesRejectsOtherLines(t *testing.T) {
	if _, _, ok := parseCPUTimes("intr 1 2 3\n"); ok {
		t.Fatal("parseCPUTimes accepted a non-cpu first line")
	}
}

func TestParseMemInfoReadsTotalAndAvailable(t *testing.T) {
	content := "MemTotal:       16318464 kB\nMemFree:         1000 kB\nMemAvailable:    8000000 kB\n"

	mem := parseMemInfo(content)
	if mem.totalKB != 16318464 || mem.availableKB != 8000000 {
		t.Fatalf("parseMemInfo = %+v", mem)
	}
}

func TestReadOSNameStripsQuotes(t *testing.T) {
	s := newTestService(t, newFakeRunner())
	writeHostFile(t, s, "/etc/os-release", "NAME=\"Demo\"\nPRETTY_NAME=\"Demo Linux 1.0\"\n")

	if got := s.readOSName(); got != "Demo Linux 1.0" {
		t.Fatalf("readOSName = %q", got)
	}
}

func TestReadOSNameMissingFileIsUnknown(t *testing.T) {
	s := newTestService(t, newFakeRunner())
	if got := s.readOSName(); got != "Unknown" {
		t.Fatalf("readOSName = %q, want Unknown", got)
	}
}

func TestReadKernelUsesUnameOutput(t *testing.T) {
	runner := newFakeRunner().on("uname -r", "6.7.1-demo\n")
	s := newTestService(t, runner)

	if got := s.readKernel(context.Background()); got != "6.7.1-demo" {
		t.Fatalf("readKernel = %q", got)
	}
}

func TestReadKernelFailureIsUnknown(t *testing.T) {
	runner := newFakeRunner().fail("uname -r")
	s := newTestService(t, runner)

	if got := s.readKernel(context.Background()); got != "Unknown" {
		t.Fatalf("readKernel = %q, want Unknown", got)
	}
}

func TestParseCPUModelKeepsSecondColonField(t *testing.T) {
	output := "Architecture:        x86_64\nModel name:          AMD Ryzen 7 5800X\n"

	if got := parseCPUModel(output); got != "AMD Ryzen 7 5800X" {
		t.Fatalf("parseCPUModel = %q", got)
	}
}

func TestParseCPUModelWithoutLscpuIsEmpty(t *testing.T) {
	if got := parseCPUModel(""); got != "" {
		t.Fatalf("parseCPUModel = %q, want empty", got)
	}
}

func TestParseDriverVersionReadsThirdField(t *testing.T) {
	output := "Vulkan Instance Version: 1.3.280\n\tdriverVersion = 550.100\n\tdriverVersion = 1\n"

	if got := parseDriverVersion(output); got != "550.100" {
		t.Fatalf("parseDriverVersion = %q", got)
	}
}

func TestReadDriverWithoutVulkaninfoIsUnknown(t *testing.T) {
	s := newTestService(t, newFakeRunner())
	if got := s.readDriver(context.Background()); got != "Unknown" {
		t.Fatalf("readDriver = %q, want Unknown", got)
	}
}

func TestParseSwapDevicesSkipsHeader(t *testing.T) {
	content := "Filename\t\t\t\tType\t\tSize\t\tUsed\t\tPriority\n" +
		"/dev/zram0                               partition\t8388604\t0\t100\n" +
		"/swapfile                                file\t\t2097148\t0\t-2\n"

	want := []string{"/dev/zram0", "/swapfile"}
	if got := parseSwapDevices(content); !reflect.DeepEqual(got, want) {
		t.Fatalf("parseSwapDevices = %v, want %v", got, want)
	}
}

func TestParseSwapDevicesEmptyWhenOnlyHeader(t *testing.T) {
	if got := parseSwapDevices("Filename\tType\n"); len(got) != 0 {
		t.Fatalf("parseSwapDevices = %v, want none", got)
	}
}

func TestSystemUsageReadsCPUAndRAMFromRoot(t *testing.T) {
	s := newTestService(t, newFakeRunner())
	writeHostFile(t, s, "/proc/stat", "cpu  1000 0 0 1000 0 0 0 0 0 0\n")
	writeHostFile(t, s, "/proc/meminfo", "MemTotal:       2097152 kB\nMemAvailable:    1048576 kB\n")

	usage := s.systemUsage(context.Background())
	if usage.CPU != 0 {
		t.Fatalf("first CPU sample = %v, want 0", usage.CPU)
	}
	if usage.RamTotalGb != 2 || usage.RamUsedGb != 1 || usage.RAM != 50 {
		t.Fatalf("RAM = total %v used %v percent %v", usage.RamTotalGb, usage.RamUsedGb, usage.RAM)
	}
}
