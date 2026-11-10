package systemsvc

// The paths match the TypeScript LSFG service, not the Go lsfg package.
const (
	lsfgManifestPath = "/usr/share/vulkan/implicit_layer.d/lsfg-vk.json"
	lsfgLibraryPath  = "/usr/lib/liblsfg-vk-layer.so"
)

func (s *Service) lsfgInstalled() bool {
	return pathExists(s.path(lsfgManifestPath)) || pathExists(s.path(lsfgLibraryPath))
}
