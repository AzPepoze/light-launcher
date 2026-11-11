package appsvc

import (
	"path/filepath"
	"strings"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

// gameRoot is one game folder: its launchers, with the primary one first.
type gameRoot struct {
	folder    string
	launchers []types.Launcher
}

// rootBucket collects the executables of one game root while grouping.
type rootBucket struct {
	folder string
	exes   []scannedFile
}

// gameRootOf returns the folder that owns an executable: the first folder below the watched
// folder. An executable directly inside the watched folder is its own game root.
func gameRootOf(watched, exePath string) string {
	rel, err := filepath.Rel(watched, exePath)
	if err != nil {
		return exePath
	}
	first, _, nested := strings.Cut(rel, string(filepath.Separator))
	if !nested {
		return exePath
	}
	return filepath.Join(watched, first)
}

// groupGameRoots groups the executables of one watched folder into one game per game root.
// remembered maps a cleaned game root to the launcher the user chose. Roots with no launcher left are dropped.
func groupGameRoots(watched string, exes []scannedFile, remembered map[string]string) []gameRoot {
	var buckets []rootBucket
	indexByKey := map[string]int{}
	for _, exe := range exes {
		folder := gameRootOf(watched, exe.Path)
		key := config.PathKey(folder)
		index, ok := indexByKey[key]
		if !ok {
			index = len(buckets)
			indexByKey[key] = index
			buckets = append(buckets, rootBucket{folder: folder})
		}
		buckets[index].exes = append(buckets[index].exes, exe)
	}

	roots := make([]gameRoot, 0, len(buckets))
	for _, bucket := range buckets {
		launchers := pickPrimaryLauncher(bucket.folder, bucket.exes, remembered[preferenceKey(bucket.folder)])
		if len(launchers) == 0 {
			continue
		}
		roots = append(roots, gameRoot{folder: bucket.folder, launchers: launchers})
	}
	return roots
}

// primaryLauncher returns the launcher that starts the game. pickPrimaryLauncher puts it first.
func primaryLauncher(root gameRoot) types.Launcher {
	return root.launchers[0]
}

// gameName is the folder name for a game with several launchers, and the launcher name otherwise.
func gameName(root gameRoot) string {
	if len(root.launchers) > 1 {
		return config.BaseName(root.folder)
	}
	return primaryLauncher(root).Name
}
