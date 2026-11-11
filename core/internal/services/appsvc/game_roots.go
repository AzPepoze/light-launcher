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

// minContainerChildren is how many game folders a folder must hold to count as a category.
const minContainerChildren = 3

// rootBucket collects the executables of one game while grouping.
type rootBucket struct {
	folder string
	exes   []scannedFile
}

// folderNode is one folder on the way from the watched folder to an executable.
type folderNode struct {
	path     string
	exes     []scannedFile
	children []*folderNode
	byKey    map[string]*folderNode
}

func newFolderNode(path string) *folderNode {
	return &folderNode{path: path, byKey: map[string]*folderNode{}}
}

// child returns the sub-folder at path, creating it on first use. Children keep their first-seen order.
func (n *folderNode) child(path string) *folderNode {
	key := config.PathKey(path)
	if existing, ok := n.byKey[key]; ok {
		return existing
	}
	created := newFolderNode(path)
	n.byKey[key] = created
	n.children = append(n.children, created)
	return created
}

// isCategory reports whether the folder holds game folders instead of being a game itself:
// it has no executable of its own and at least minContainerChildren sub-folders with executables.
func (n *folderNode) isCategory() bool {
	return len(n.exes) == 0 && len(n.children) >= minContainerChildren
}

// collectExes appends every executable at or below the folder.
func (n *folderNode) collectExes(out []scannedFile) []scannedFile {
	out = append(out, n.exes...)
	for _, child := range n.children {
		out = child.collectExes(out)
	}
	return out
}

// buildFolderTree places each executable in the folder tree below the watched folder.
// Executables are relative to watched, so they always fall under it.
func buildFolderTree(watched string, exes []scannedFile) *folderNode {
	root := newFolderNode(watched)
	for _, exe := range exes {
		rel, err := filepath.Rel(watched, exe.Path)
		if err != nil {
			root.exes = append(root.exes, exe)
			continue
		}
		parts := strings.Split(rel, string(filepath.Separator))
		node, path := root, watched
		for _, dir := range parts[:len(parts)-1] {
			path = filepath.Join(path, dir)
			node = node.child(path)
		}
		node.exes = append(node.exes, exe)
	}
	return root
}

// appendLevel adds the games of one level: each executable directly in it is its own game, and
// each sub-folder is a game folder, unless it is a category, whose games are added instead.
func appendLevel(level *folderNode, buckets []rootBucket) []rootBucket {
	for _, exe := range level.exes {
		buckets = append(buckets, rootBucket{folder: exe.Path, exes: []scannedFile{exe}})
	}
	for _, child := range level.children {
		if child.isCategory() {
			buckets = appendLevel(child, buckets)
		} else {
			buckets = append(buckets, rootBucket{folder: child.path, exes: child.collectExes(nil)})
		}
	}
	return buckets
}

// groupGameRoots groups the executables of one watched folder into games. A folder that holds
// several game folders is a category, and its game folders are the games. remembered maps a
// cleaned game root to the launcher the user chose. Helper programs are ignored when counting,
// and roots with no launcher left are dropped.
func groupGameRoots(watched string, exes []scannedFile, remembered map[string]string) []gameRoot {
	tree := buildFolderTree(watched, withoutJunkExecutables(exes))
	buckets := appendLevel(tree, nil)

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
