package appsvc

import (
	"fmt"
	"path/filepath"
	"testing"

	"light-launcher/core/internal/config"
)

// BenchmarkGetAutoScannedGames2000Exes times a forced scan of 2,000 executables in 1,000 game
// folders, inside 20 category folders.
func BenchmarkGetAutoScannedGames2000Exes(b *testing.B) {
	isolateConfigHome(b)
	svc := newTestService(b)
	root := config.CleanPath(b.TempDir())
	writeGameTree(b, root, 20, 50)
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
// The 2,000 executables sit in 1,000 game folders inside 20 category folders.
func BenchmarkGroupGameRoots2000Exes(b *testing.B) {
	watched := filepath.Join("games", "W")
	exes := make([]scannedFile, 0, 2000)
	for c := range 20 {
		for g := range 50 {
			name := fmt.Sprintf("Game%02d-%03d", c, g)
			dir := filepath.Join(watched, fmt.Sprintf("Category%02d", c), name)
			exes = append(exes,
				scannedFile{Path: filepath.Join(dir, name+".exe"), Size: 900},
				scannedFile{Path: filepath.Join(dir, name+"32.exe"), Size: 100},
			)
		}
	}

	b.ResetTimer()
	for range b.N {
		if roots := groupGameRoots(watched, exes, nil); len(roots) != 1000 {
			b.Fatalf("got %d roots, want 1000", len(roots))
		}
	}
}
