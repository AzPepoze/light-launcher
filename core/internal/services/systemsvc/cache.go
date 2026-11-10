package systemsvc

import (
	"context"
	"fmt"
)

func (s *Service) shaderCacheSize(ctx context.Context) string {
	return formatCacheSize(s.shaderCacheBytes(ctx))
}

// formatCacheSize shows "0 MB" for an empty cache, otherwise MB with one decimal.
func formatCacheSize(bytes int64) string {
	if bytes == 0 {
		return "0 MB"
	}
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
}
