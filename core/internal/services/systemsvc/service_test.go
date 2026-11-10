package systemsvc

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/platform"
	"light-launcher/core/internal/types"
)

func TestRegisterAddsEveryMethod(t *testing.T) {
	router := api.NewRouter()
	newTestService(t, newFakeRunner()).register(router)

	want := []string{
		"CleanupProcesses", "ClearShaderCache", "ClearSwap", "DropCaches",
		"GetListGpus", "GetPlatform", "GetShaderCacheSize", "GetSystemInfo",
		"GetSystemToolsStatus", "GetSystemUsage", "GetUtilsStatus",
	}
	got := router.Methods()
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("registered methods = %v, want %v", got, want)
	}
}

func TestPlatformReportsOSAndFeatures(t *testing.T) {
	router := api.NewRouter()
	newTestService(t, newFakeRunner()).register(router)

	result, err := router.Dispatch(context.Background(), "GetPlatform", nil)
	if err != nil {
		t.Fatalf("GetPlatform: %v", err)
	}
	info, ok := result.(PlatformInfo)
	if !ok {
		t.Fatalf("GetPlatform returned %T", result)
	}
	if info.OS != platform.OS() {
		t.Fatalf("os = %q, want %q", info.OS, platform.OS())
	}
	assertKeys(t, info, []string{"features", "os"})
	assertKeys(t, info.Features, []string{
		"gameMode", "gamescope", "liveTerminalLogs", "lsfg", "mangoHud", "prefix", "proton",
	})
}

// The UI reads these JSON keys, so they must match src/shared/types/system.types.ts.
func TestJSONKeysMatchTypeScript(t *testing.T) {
	assertKeys(t, types.SystemToolsStatus{}, []string{
		"hasGameMode", "hasGamescope", "hasMangoHud", "hasVulkanInfo",
	})
	assertKeys(t, types.SystemInfo{}, []string{"cpu", "driver", "gpu", "kernel", "os", "ram"})
	assertKeys(t, types.SystemInfo{Gpus: []string{"GPU"}}, []string{"cpu", "driver", "gpu", "gpus", "kernel", "os", "ram"})
	assertKeys(t, types.SystemUsage{}, []string{"cpu", "gpus", "ram", "ramTotalGb", "ramUsedGb"})
	assertKeys(t, types.UtilsStatus{}, []string{"isLsfgInstalled", "lsfgVersion"})
}

func TestListGpusIsEmptyArrayNotNull(t *testing.T) {
	s := newTestService(t, newFakeRunner())
	data := mustMarshal(t, s.listGpus(context.Background()))
	if string(data) != "[]" {
		t.Fatalf("GetListGpus JSON = %s, want []", data)
	}
}

func TestSystemUsageGpusIsEmptyArrayNotNull(t *testing.T) {
	s := newTestService(t, newFakeRunner())
	usage := s.systemUsage(context.Background())
	data := mustMarshal(t, usage.GPUs)
	if string(data) != "[]" {
		t.Fatalf("gpus JSON = %s, want []", data)
	}
}

func TestCPUSamplerReportsBusyShareBetweenCalls(t *testing.T) {
	var sampler cpuSampler

	if got := sampler.busyPercent(1000, 800); got != 0 {
		t.Fatalf("first sample = %v, want 0", got)
	}
	// 200 jiffies passed, 100 of them idle: half busy.
	if got := sampler.busyPercent(1200, 900); got != 50 {
		t.Fatalf("second sample = %v, want 50", got)
	}
}

func TestApplyMemoryFillsGbAndPercent(t *testing.T) {
	var usage types.SystemUsage
	applyMemory(&usage, memInfo{totalKB: 16 * kbPerGb, availableKB: 4 * kbPerGb})

	if usage.RamTotalGb != 16 || usage.RamUsedGb != 12 || usage.RAM != 75 {
		t.Fatalf("got total=%v used=%v percent=%v, want 16, 12, 75",
			usage.RamTotalGb, usage.RamUsedGb, usage.RAM)
	}
}

func TestApplyMemoryIgnoresMissingTotal(t *testing.T) {
	var usage types.SystemUsage
	applyMemory(&usage, memInfo{})

	if usage.RamTotalGb != 0 || usage.RamUsedGb != 0 || usage.RAM != 0 {
		t.Fatalf("missing total must leave RAM at zero, got %+v", usage)
	}
}

func TestFormatCacheSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 MB"},
		{1024 * 1024, "1.0 MB"},
		{1572864, "1.5 MB"},
	}
	for _, tt := range tests {
		if got := formatCacheSize(tt.bytes); got != tt.want {
			t.Errorf("formatCacheSize(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

func assertKeys(t *testing.T, value any, want []string) {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(mustMarshal(t, value), &fields); err != nil {
		t.Fatalf("unmarshal %T: %v", value, err)
	}
	got := make([]string, 0, len(fields))
	for key := range fields {
		got = append(got, key)
	}
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%T JSON keys = %v, want %v", value, got, want)
	}
}

func mustMarshal(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal %T: %v", value, err)
	}
	return data
}
