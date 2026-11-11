package appsvc

import (
	"fmt"
	"path/filepath"
	"testing"

	"light-launcher/core/internal/config"
)

// BenchmarkGetAutoScannedGames2000Exes times a forced scan of 2,000 executables in 1,000 game folders.
func BenchmarkGetAutoScannedGames2000Exes(b *testing.B) {
	isolateConfigHome(b)
	svc := newTestService(b)
	root := config.CleanPath(b.TempDir())
	writeGameTree(b, root, 1000)
	if err := svc.addScanFolder(root); err != nil {
		b.Fatalf("addScanFolder: %v", err)
	}
	if _, err := svc.getAutoScannedGames(true); err != nil {
		b.Fatalf("first scan: %v", err)
	}

	b.ResetTimer()
	for range b.N {
		if _, err := svc.getAutoScannedGames(true); err != nil {
			b.Fatalf("getAutoScannedGames: %v", err)
		}
	}
}

// BenchmarkGroupGameRoots2000Exes times only the grouping and launcher picking, without disk access.
func BenchmarkGroupGameRoots2000Exes(b *testing.B) {
	watched := filepath.Join("games", "W")
	exes := make([]scannedFile, 0, 2000)
	for i := range 1000 {
		name := fmt.Sprintf("Game%04d", i)
		exes = append(exes,
			exeFile(900, watched, name, name+".exe"),
			exeFile(100, watched, name, name+"32.exe"),
		)
	}

	b.ResetTimer()
	for range b.N {
		if roots := groupGameRoots(watched, exes, nil); len(roots) != 1000 {
			b.Fatalf("got %d roots, want 1000", len(roots))
		}
	}
}
