package systemsvc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"light-launcher/core/internal/logger"
)

// cleanupProcessNames are game helpers left behind after a launch.
var cleanupProcessNames = []string{
	"umu-run",
	"pressure-vessel",
	"gamescopereaper",
	"steam-runtime-launcher-service",
	"srt-bwrap",
	"reaper",
}

// swapOnScript turns every swap device off, then back on. The devices are passed as arguments, not spliced into the script.
const swapOnScript = `swapoff -a ; for dev in "$@"; do swapon "$dev"; done`

func (s *Service) dropCaches(ctx context.Context) error {
	if _, err := s.runner.Output(ctx, "sync"); err != nil {
		return fmt.Errorf("sync: %w", err)
	}
	if _, err := s.runner.Output(ctx, "pkexec", "sysctl", "-w", "vm.drop_caches=3"); err != nil {
		return fmt.Errorf("drop caches: %w", err)
	}
	return nil
}

// clearSwap logs failures instead of returning them, so the UI does not show an error.
func (s *Service) clearSwap(ctx context.Context) error {
	if err := s.restartSwap(ctx); err != nil {
		logger.Warn("System", "Error clearing swap: %v", err)
	}
	return nil
}

func (s *Service) restartSwap(ctx context.Context) error {
	content, err := os.ReadFile(s.path("/proc/swaps"))
	if err != nil {
		return err
	}
	devices := parseSwapDevices(string(content))
	if len(devices) == 0 {
		return nil
	}
	args := append([]string{"sh", "-c", swapOnScript, "sh"}, devices...)
	_, err = s.runner.Output(ctx, "pkexec", args...)
	return err
}

// parseSwapDevices returns the first column of each /proc/swaps row, skipping the header.
func parseSwapDevices(content string) []string {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	var devices []string
	for _, line := range lines[1:] {
		if fields := strings.Fields(line); len(fields) > 0 {
			devices = append(devices, fields[0])
		}
	}
	return devices
}

// cleanupProcesses runs pkill for each helper. pkill exits non-zero when nothing matches, which is normal.
func (s *Service) cleanupProcesses(ctx context.Context) error {
	for _, name := range cleanupProcessNames {
		_, _ = s.runner.Output(ctx, "pkill", "-f", name)
	}
	return nil
}

func (s *Service) shaderCachePaths() []string {
	return []string{
		filepath.Join(s.home, ".cache/mesa_shader_cache"),
		filepath.Join(s.home, ".cache/nvidia/GLCache"),
		filepath.Join(s.home, ".nv/GLCache"),
		filepath.Join(s.home, ".cache/AMD/VkCache"),
		filepath.Join(s.home, ".cache/radv_builtin_shaders"),
	}
}

func (s *Service) shaderCacheBytes(ctx context.Context) int64 {
	var total int64
	for _, dir := range s.shaderCachePaths() {
		if pathExists(dir) {
			total += s.diskUsageBytes(ctx, dir)
		}
	}
	return total
}

// diskUsageBytes runs `du -sb` on dir. A failing du counts as zero.
func (s *Service) diskUsageBytes(ctx context.Context, dir string) int64 {
	out, err := s.runner.Output(ctx, "du", "-sb", dir)
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		return 0
	}
	size, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return 0
	}
	return size
}

// clearShaderCache removes each cache folder. A folder that cannot be removed is skipped.
func (s *Service) clearShaderCache(_ context.Context) error {
	for _, dir := range s.shaderCachePaths() {
		if pathExists(dir) {
			_ = os.RemoveAll(dir)
		}
	}
	return nil
}
