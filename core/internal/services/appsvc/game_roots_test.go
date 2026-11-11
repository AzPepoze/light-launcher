package appsvc

import (
	"path/filepath"
	"slices"
	"testing"

	"light-launcher/core/internal/types"
)

func TestGameRootOf(t *testing.T) {
	watched := filepath.Join("games", "W")
	tests := []struct {
		name string
		exe  []string
		want []string
	}{
		{"nested in a game folder", []string{"games", "W", "MyGame", "Binaries", "Win64", "x.exe"}, []string{"games", "W", "MyGame"}},
		{"one level below", []string{"games", "W", "A", "B.exe"}, []string{"games", "W", "A"}},
		{"directly inside the watched folder", []string{"games", "W", "Single.exe"}, []string{"games", "W", "Single.exe"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := gameRootOf(watched, filepath.Join(test.exe...))
			if want := filepath.Join(test.want...); got != want {
				t.Fatalf("gameRootOf = %q, want %q", got, want)
			}
		})
	}
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
	bone := roots[0]
	if bone.folder != filepath.Join("games", "W", "MyGame") || len(bone.launchers) != 2 {
		t.Fatalf("MyGame root = %+v", bone)
	}
	if !bone.launchers[0].Primary || bone.launchers[0].Kind != types.LauncherMain {
		t.Errorf("first MyGame launcher must be the primary main launcher: %+v", bone.launchers[0])
	}
	if roots[1].folder != filepath.Join("games", "W", "Solo.exe") || len(roots[1].launchers) != 1 {
		t.Errorf("Solo root = %+v", roots[1])
	}
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
