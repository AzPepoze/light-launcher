package appsvc

import (
	"os"
	"path/filepath"
	"testing"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

// watchedGameFolder makes a watched folder with one game folder that holds Game.exe and Other.exe.
func watchedGameFolder(t *testing.T, svc *service) (folder, game, gameExe, otherExe string) {
	t.Helper()
	folder = config.CleanPath(t.TempDir())
	game = filepath.Join(folder, "Game")
	gameExe = filepath.Join(game, "Game.exe")
	otherExe = filepath.Join(game, "Other.exe")
	mustWrite(t, gameExe)
	mustWrite(t, otherExe)
	if err := svc.addScanFolder(folder); err != nil {
		t.Fatalf("addScanFolder: %v", err)
	}
	return folder, game, gameExe, otherExe
}

func TestSetPreferredLauncherChangesPrimaryAndReturnsCard(t *testing.T) {
	svc := newTestService(t)
	_, game, gameExe, otherExe := watchedGameFolder(t, svc)

	card, err := svc.setPreferredLauncher(game, otherExe)
	if err != nil {
		t.Fatalf("setPreferredLauncher: %v", err)
	}
	if card.Path != otherExe || card.Folder != game || card.LauncherCount != 2 {
		t.Fatalf("card = %+v, want primary %s in folder %s with 2 launchers", card, otherExe, game)
	}
	if card.Config.LauncherPath != otherExe {
		t.Fatalf("card config = %q, want the config of the chosen launcher", card.Config.LauncherPath)
	}
	if card.Name != "Game" {
		t.Fatalf("card name = %q, want the folder name for a multi-launcher game", card.Name)
	}

	card, err = svc.clearPreferredLauncher(game)
	if err != nil {
		t.Fatalf("clearPreferredLauncher: %v", err)
	}
	if card.Path != gameExe {
		t.Fatalf("after clear primary = %q, want the name match %q", card.Path, gameExe)
	}
}

func TestSetPreferredLauncherRejectsInvalidChoices(t *testing.T) {
	svc := newTestService(t)
	_, game, _, otherExe := watchedGameFolder(t, svc)
	outside := filepath.Join(t.TempDir(), "Outside.exe")
	mustWrite(t, outside)
	readme := filepath.Join(game, "readme.txt")
	mustWrite(t, readme)
	crash := filepath.Join(game, "UnityCrashHandler64.exe")
	mustWrite(t, crash)
	missing := filepath.Join(game, "Missing.exe")

	tests := []struct {
		name   string
		folder string
		path   string
	}{
		{"empty folder", "", otherExe},
		{"folder not watched", t.TempDir(), otherExe},
		{"launcher outside the folder", game, outside},
		{"launcher is not an exe", game, readme},
		{"launcher is a helper program", game, crash},
		{"launcher file missing", game, missing},
		{"launcher is the folder itself", game, game},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := svc.setPreferredLauncher(test.folder, test.path); err == nil {
				t.Fatal("setPreferredLauncher succeeded, want an error")
			}
			if got := svc.preferences.Preferred(); len(got) != 0 {
				t.Fatalf("rejected choice was saved: %v", got)
			}
		})
	}
}

func TestRememberedLauncherThatIsDeletedFallsBackToRules(t *testing.T) {
	svc := newTestService(t)
	_, game, gameExe, otherExe := watchedGameFolder(t, svc)
	if _, err := svc.setPreferredLauncher(game, otherExe); err != nil {
		t.Fatalf("setPreferredLauncher: %v", err)
	}
	if err := os.Remove(otherExe); err != nil {
		t.Fatal(err)
	}

	groups, err := svc.getAutoScannedGames(true)
	if err != nil {
		t.Fatalf("getAutoScannedGames: %v", err)
	}
	game0 := groups[0].Games[0]
	if game0.Path != gameExe || game0.LauncherCount != 1 {
		t.Fatalf("card = %+v, want the name match %s", game0, gameExe)
	}
}

func TestGetAutoScannedGamesOneCardPerGameFolder(t *testing.T) {
	svc := newTestService(t)
	root := config.CleanPath(t.TempDir())
	const categories, gamesPerCategory = 20, 50
	writeGameTree(t, root, categories, gamesPerCategory)
	if err := svc.addScanFolder(root); err != nil {
		t.Fatalf("addScanFolder: %v", err)
	}

	groups, err := svc.getAutoScannedGames(true)
	if err != nil {
		t.Fatalf("getAutoScannedGames: %v", err)
	}
	games := groups[0].Games
	if want := categories * gamesPerCategory; len(games) != want {
		t.Fatalf("got %d cards for %d game folders", len(games), want)
	}
	for _, game := range games {
		if game.LauncherCount != 2 || filepath.Base(game.Path) != filepath.Base(game.Folder)+".exe" {
			t.Fatalf("card %s has %d launchers, primary %s", game.Folder, game.LauncherCount, game.Path)
		}
	}
}

func TestHandAddedGameHasOneLauncher(t *testing.T) {
	svc := newTestService(t)
	exe := filepath.Join(t.TempDir(), "Hand", "Hand.exe")
	cfg := newDefaultGameConfig("Hand", exe)
	if err := svc.games.Save(&cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	games := svc.getAllGames()
	if len(games) != 1 {
		t.Fatalf("got %d games, want 1", len(games))
	}
	game := games[0]
	want := []types.Launcher{{Name: "Hand", Path: exe, Kind: types.LauncherMain, Primary: true}}
	if game.LauncherCount != 1 || game.Folder != exe || len(game.Launchers) != 1 || game.Launchers[0] != want[0] {
		t.Fatalf("hand-added game = %+v", game)
	}
}
