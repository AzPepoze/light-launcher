package appsvc

import "context"

// extractExeIcon returns no icon on Windows. Icon extraction from .exe files is not implemented yet,
// so the UI shows its placeholder.
func extractExeIcon(_ context.Context, _ string) (string, error) {
	return "", nil
}
