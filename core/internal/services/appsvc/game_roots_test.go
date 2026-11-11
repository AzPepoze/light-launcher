package appsvc

import (
	"path/filepath"
	"slices"
	"testing"

	"light-launcher/core/internal/types"
)

// watchedPath is a path below the watched folder "games/W".
func watchedPath(parts ...string) string {
	return filepath.Join(append([]string{"games", "W"}, parts...)...)
}

// watchedExe is a candidate executable below the watched folder.
func watchedExe(size int64, parts ...string) scannedFile {
	return scannedFile{Path: watchedPath(parts...), Size: size}
}

// unrealGame is a game folder with an Unreal root stub and a Win64 shipping build.
func unrealGame(name string) []scannedFile {
	return []scannedFile{
		watchedExe(1000, name, name+".exe"),
		watchedExe(500, name, "Binaries", "Win64", name+"-Win64-Shipping.exe"),
	}
}

func TestGroupGameRootsDetectsGameFolders(t *testing.T) {
	var categoryLayout []scannedFile
	for _, category := range []string{"Action", "Puzzle"} {
		for _, game := range []string{"G1", "G2", "G3", "G4"} {
			categoryLayout = append(categoryLayout, unrealGame(filepath.Join(category, game))...)
		}
	}

	// publisher/category/game/game.exe: only the game folders are games.
	var publisherLayout []scannedFile
	var publisherWant []gameRootWant
	for _, publisher := range []string{"P1", "P2"} {
		for _, category := range []string{"C1", "C2", "C3"} {
			for _, game := range []string{"G1", "G2", "G3"} {
				publisherLayout = append(publisherLayout, watchedExe(10, publisher, category, game, game+".exe"))
				publisherWant = append(publisherWant, gameRootWant{watchedPath(publisher, category, game), 1})
			}
		}
	}

	tests := []struct {
		name string
		exes []scannedFile
		want []gameRootWant
	}{
		{
			name: "category folders are not games, their game folders are",
			exes: categoryLayout,
			want: []gameRootWant{
				{watchedPath("Action", "G1"), 2}, {watchedPath("Action", "G2"), 2},
				{watchedPath("Action", "G3"), 2}, {watchedPath("Action", "G4"), 2},
				{watchedPath("Puzzle", "G1"), 2}, {watchedPath("Puzzle", "G2"), 2},
				{watchedPath("Puzzle", "G3"), 2}, {watchedPath("Puzzle", "G4"), 2},
			},
		},
		{
			name: "game folder with two sub-apps and no exe of its own stays one game",
			exes: []scannedFile{
				watchedExe(10, "Suite", "Engine", "Engine.exe"),
				watchedExe(10, "Suite", "Tools", "Tools.exe"),
			},
			want: []gameRootWant{{watchedPath("Suite"), 2}},
		},
		{
			name: "game folder with its own exe and three sub-folders stays one game",
			exes: []scannedFile{
				watchedExe(10, "Main", "Main.exe"),
				watchedExe(10, "Main", "A", "a.exe"),
				watchedExe(10, "Main", "B", "b.exe"),
				watchedExe(10, "Main", "C", "c.exe"),
			},
			want: []gameRootWant{{watchedPath("Main"), 4}},
		},
		{
			name: "three levels: only the game folder level is a game",
			exes: publisherLayout,
			want: publisherWant,
		},
		{
			name: "exe directly in the watched folder next to a category is its own game",
			exes: append([]scannedFile{watchedExe(10, "Loose.exe")}, categoryLayout[:8]...),
			want: []gameRootWant{
				{watchedPath("Loose.exe"), 1},
				{watchedPath("Action", "G1"), 2}, {watchedPath("Action", "G2"), 2},
				{watchedPath("Action", "G3"), 2}, {watchedPath("Action", "G4"), 2},
			},
		},
		{
			name: "single game folder directly under the watched folder",
			exes: []scannedFile{watchedExe(10, "Solo", "Solo.exe")},
			want: []gameRootWant{{watchedPath("Solo"), 1}},
		},
		{
			name: "empty watched folder has no games",
			exes: nil,
			want: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			roots := groupGameRoots(watchedPath(), test.exes, nil)
			got := make([]gameRootWant, 0, len(roots))
			for _, root := range roots {
				got = append(got, gameRootWant{root.folder, len(root.launchers)})
			}
			if !slices.Equal(got, test.want) {
				t.Fatalf("roots\n got: %v\nwant: %v", got, test.want)
			}
		})
	}
}

// gameRootWant is an expected game: its folder and how many launchers it has.
type gameRootWant struct {
	folder    string
	launchers int
}

func TestGroupGameRootsOneGamePerFolder(t *testing.T) {
	watched := filepath.Join("games", "W")
	exes := []scannedFile{
		exeFile(1000, "games", "W", "MyGame", "MyGame.exe"),
		exeFile(500, "games", "W", "MyGame", "Binaries", "Win64", "MyGame-Win64-Shipping.exe"),
		exeFile(10, "games", "W", "Solo.exe"),
		exeFile(900, "games", "W", "Junk", "UnityCrashHandler64.exe"),
	}

	roots := groupGameRoots(watched, exes, nil)

	if len(roots) != 2 {
		t.Fatalf("got %d game roots, want 2 (the junk-only folder has no game)", len(roots))
	}
	myGame, ok := rootAt(roots, watchedPath("MyGame"))
	if !ok || len(myGame.launchers) != 2 {
		t.Fatalf("MyGame root = %+v, found %v", myGame, ok)
	}
	if !myGame.launchers[0].Primary || myGame.launchers[0].Kind != types.LauncherMain {
		t.Errorf("first MyGame launcher must be the primary main launcher: %+v", myGame.launchers[0])
	}
	solo, ok := rootAt(roots, watchedPath("Solo.exe"))
	if !ok || len(solo.launchers) != 1 {
		t.Errorf("Solo root = %+v, found %v", solo, ok)
	}
}

// rootAt finds the game with the given folder.
func rootAt(roots []gameRoot, folder string) (gameRoot, bool) {
	for _, root := range roots {
		if root.folder == folder {
			return root, true
		}
	}
	return gameRoot{}, false
}

func TestGroupGameRootsUsesRememberedLauncher(t *testing.T) {
	watched := filepath.Join("games", "W")
	exes := []scannedFile{
		exeFile(900, "games", "W", "Game", "Game.exe"),
		exeFile(100, "games", "W", "Game", "Other.exe"),
	}
	folder := filepath.Join("games", "W", "Game")
	remembered := map[string]string{
		preferenceKey(folder): filepath.Join("games", "W", "Game", "Other.exe"),
	}

	roots := groupGameRoots(watched, exes, remembered)

	if len(roots) != 1 {
		t.Fatalf("got %d roots, want 1", len(roots))
	}
	if got := primaryLauncher(roots[0]).Path; got != filepath.Join("games", "W", "Game", "Other.exe") {
		t.Fatalf("primary = %q, want the remembered launcher", got)
	}
}

func TestGroupGameRootsKeepsOrderOfFirstAppearance(t *testing.T) {
	watched := filepath.Join("games", "W")
	exes := []scannedFile{
		exeFile(1, "games", "W", "B", "B.exe"),
		exeFile(1, "games", "W", "A", "A.exe"),
		exeFile(1, "games", "W", "B", "Other.exe"),
	}

	roots := groupGameRoots(watched, exes, nil)

	var folders []string
	for _, root := range roots {
		folders = append(folders, root.folder)
	}
	want := []string{filepath.Join("games", "W", "B"), filepath.Join("games", "W", "A")}
	if !slices.Equal(folders, want) {
		t.Fatalf("folders = %v, want %v", folders, want)
	}
}
