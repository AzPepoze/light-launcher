package systemsvc

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// The fake runner stands in for pkexec, sync and pkill, so none of these tests touch the real system.

func TestShaderCacheSizeSumsExistingFolders(t *testing.T) {
	s := newTestService(t, newFakeRunner())
	paths := s.shaderCachePaths()
	for _, dir := range paths[:2] {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	runner := newFakeRunner().
		on("du -sb "+paths[0], "1048576\t"+paths[0]+"\n").
		on("du -sb "+paths[1], "524288\t"+paths[1]+"\n")
	s.runner = runner

	if got := s.shaderCacheSize(context.Background()); got != "1.5 MB" {
		t.Fatalf("shaderCacheSize = %q, want 1.5 MB", got)
	}
}

func TestShaderCacheSizeSkipsMissingFolders(t *testing.T) {
	runner := newFakeRunner()
	s := newTestService(t, runner)

	if got := s.shaderCacheSize(context.Background()); got != "0 MB" {
		t.Fatalf("shaderCacheSize = %q, want 0 MB", got)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("du ran for missing folders: %v", runner.callLines())
	}
}

func TestClearShaderCacheRemovesOnlyCacheFolders(t *testing.T) {
	s := newTestService(t, newFakeRunner())
	cacheDir := s.shaderCachePaths()[0]
	unrelated := filepath.Join(s.home, ".cache", "other-app")
	for _, dir := range []string{cacheDir, unrelated} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	if err := s.clearShaderCache(context.Background()); err != nil {
		t.Fatalf("clearShaderCache: %v", err)
	}
	if pathExists(cacheDir) {
		t.Fatal("shader cache folder was not removed")
	}
	if !pathExists(unrelated) {
		t.Fatal("unrelated folder was removed")
	}
}

func TestDropCachesRunsSyncThenSysctl(t *testing.T) {
	runner := newFakeRunner().on("sync", "").on("pkexec sysctl -w vm.drop_caches=3", "")
	s := newTestService(t, runner)

	if err := s.dropCaches(context.Background()); err != nil {
		t.Fatalf("dropCaches: %v", err)
	}
	want := [][]string{{"sync"}, {"pkexec", "sysctl", "-w", "vm.drop_caches=3"}}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("calls = %v, want %v", runner.calls, want)
	}
}

func TestDropCachesStopsWhenSyncFails(t *testing.T) {
	runner := newFakeRunner().fail("sync")
	s := newTestService(t, runner)

	if err := s.dropCaches(context.Background()); err == nil {
		t.Fatal("dropCaches returned nil after sync failed")
	}
	if len(runner.calls) != 1 {
		t.Fatalf("calls = %v, want only sync", runner.callLines())
	}
}

func TestClearSwapRestartsEverySwapDevice(t *testing.T) {
	const pkexecCall = "pkexec sh -c " + swapOnScript + " sh /dev/zram0 /swapfile"
	runner := newFakeRunner().on(pkexecCall, "")
	s := newTestService(t, runner)
	writeHostFile(t, s, "/proc/swaps", "Filename\tType\tSize\n/dev/zram0 partition 8388604\n/swapfile file 2097148\n")

	if err := s.clearSwap(context.Background()); err != nil {
		t.Fatalf("clearSwap: %v", err)
	}
	want := [][]string{{"pkexec", "sh", "-c", swapOnScript, "sh", "/dev/zram0", "/swapfile"}}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("calls = %v, want %v", runner.calls, want)
	}
}

func TestClearSwapWithoutSwapDoesNothing(t *testing.T) {
	runner := newFakeRunner()
	s := newTestService(t, runner)
	writeHostFile(t, s, "/proc/swaps", "Filename\tType\tSize\n")

	if err := s.clearSwap(context.Background()); err != nil {
		t.Fatalf("clearSwap: %v", err)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("calls = %v, want none", runner.callLines())
	}
}

func TestClearSwapLogsFailureWithoutReturningIt(t *testing.T) {
	runner := newFakeRunner().fail("pkexec sh -c " + swapOnScript + " sh /dev/zram0")
	s := newTestService(t, runner)
	writeHostFile(t, s, "/proc/swaps", "Filename\tType\tSize\n/dev/zram0 partition 8388604\n")

	if err := s.clearSwap(context.Background()); err != nil {
		t.Fatalf("clearSwap returned %v, want nil like the TypeScript service", err)
	}
}

func TestCleanupProcessesRunsPkillForEachHelper(t *testing.T) {
	runner := newFakeRunner()
	s := newTestService(t, runner)

	if err := s.cleanupProcesses(context.Background()); err != nil {
		t.Fatalf("cleanupProcesses: %v", err)
	}
	want := []string{
		"pkill -f umu-run",
		"pkill -f pressure-vessel",
		"pkill -f gamescopereaper",
		"pkill -f steam-runtime-launcher-service",
		"pkill -f srt-bwrap",
		"pkill -f reaper",
	}
	if got := runner.callLines(); !reflect.DeepEqual(got, want) {
		t.Fatalf("calls = %v, want %v", got, want)
	}
}

func TestLsfgInstalledChecksManifestOrLibrary(t *testing.T) {
	s := newTestService(t, newFakeRunner())
	if s.lsfgInstalled() {
		t.Fatal("lsfgInstalled = true with no files")
	}

	writeHostFile(t, s, lsfgLibraryPath, "")
	if !s.lsfgInstalled() {
		t.Fatal("lsfgInstalled = false with the library present")
	}
}
