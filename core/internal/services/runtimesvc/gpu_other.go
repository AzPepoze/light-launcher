//go:build !linux

package runtimesvc

// listGpuNames reports no GPUs outside Linux. The LSFG and runner code that uses it only runs on Linux.
func listGpuNames() []string {
	return []string{}
}
