package library

import (
	"context"
	"path/filepath"
	"testing"

	"light-launcher/core/internal/types"
)

func localConfig(id, name, exe string) types.LaunchOptions {
	return types.LaunchOptions{
		ID:           id,
		Name:         name,
		LauncherPath: exe,
		GamePath:     filepath.Dir(exe),
		CustomArgs:   "-windowed",
	}
}

func TestBackfillLocalImportsOneRowPerGameFolder(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	root := filepath.Join(t.TempDir(), "Harbor")
	first := localConfig("one", "Harbor", filepath.Join(root, "Harbor.exe"))
	second := localConfig("two", "Harbor Config", filepath.Join(root, "Config.exe"))
	written, err := lib.BackfillLocal(context.Background(), LocalBackfillInput{
		Configs:     []types.LaunchOptions{first, second},
		ScanFolders: []string{filepath.Dir(root)},
		Activity: []types.GameActivity{{
			GamePath: first.LauncherPath, TotalPlaytimeSeconds: 120, LastPlayedAt: 1_700_000_000,
		}},
	})
	if err != nil {
		t.Fatalf("BackfillLocal: %v", err)
	}
	if written != 1 {
		t.Fatalf("written = %d, want 1", written)
	}

	page, err := lib.List(context.Background(), Filter{Source: "manual"})
	if err != nil {
		t.Fatal(err)
	}
	if totalOf(page) != 1 {
		t.Fatalf("total = %d, want 1", totalOf(page))
	}
	game := page.Items[0]
	if game.Origin != "scan" || game.SourceKind != "local" {
		t.Errorf("origin/sourceKind = %q/%q, want scan/local", game.Origin, game.SourceKind)
	}
	if game.Pinned != 0 {
		t.Errorf("pinned = %d, want 0 for a scanned config", game.Pinned)
	}
	if game.LocalConfig == nil || game.LocalConfig.ID != "one" {
		t.Errorf("local config = %+v, want the first config", game.LocalConfig)
	}
	if game.PlaytimeS != 120 {
		t.Errorf("playtime = %d, want 120", game.PlaytimeS)
	}
	if game.FolderPath == nil || *game.FolderPath != root {
		t.Errorf("folder = %v, want %s", game.FolderPath, root)
	}
	launchers, err := lib.Launchers(context.Background(), game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(launchers) != 2 || !launchers[0].Primary {
		t.Fatalf("launchers = %+v, want two with a primary first", launchers)
	}
}

func TestBackfillLocalIsIdempotentAndSkipsBlacklist(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	root := filepath.Join(t.TempDir(), "Harbor")
	config := localConfig("one", "Harbor", filepath.Join(root, "Harbor.exe"))
	_, err := lib.BackfillLocal(context.Background(), LocalBackfillInput{
		Configs:   []types.LaunchOptions{config},
		Blacklist: []string{config.LauncherPath},
	})
	if err != nil {
		t.Fatal(err)
	}
	page, err := lib.List(context.Background(), Filter{Source: "manual"})
	if err != nil {
		t.Fatal(err)
	}
	if totalOf(page) != 0 {
		t.Fatalf("blacklisted game was imported: %d rows", totalOf(page))
	}
	if again, err := lib.BackfillLocal(context.Background(), LocalBackfillInput{Configs: []types.LaunchOptions{config}}); err != nil || again != 0 {
		t.Fatalf("second backfill = %d, %v; want 0, nil", again, err)
	}
}

func TestSyncScannedAddsUpdatesAndSweeps(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	root := filepath.Join(t.TempDir(), "Library")
	gameRoot := filepath.Join(root, "Game")
	game := types.GameInfo{
		Name:   "Game",
		Path:   filepath.Join(gameRoot, "Game.exe"),
		Folder: gameRoot,
		Config: localConfig("scan", "Game", filepath.Join(gameRoot, "Game.exe")),
		Launchers: []types.Launcher{{
			Name: "Game", Path: filepath.Join(gameRoot, "Game.exe"), Kind: types.LauncherMain, Primary: true,
		}},
	}
	groups := []types.ScannedFolderGroup{{FolderPath: root, FolderName: "Library", Games: []types.GameInfo{game}}}
	result, err := lib.SyncScanned(context.Background(), groups)
	if err != nil {
		t.Fatal(err)
	}
	if result.Added != 1 {
		t.Fatalf("added = %d, want 1", result.Added)
	}
	page, err := lib.List(context.Background(), Filter{Source: "manual"})
	if err != nil {
		t.Fatal(err)
	}
	if totalOf(page) != 1 {
		t.Fatalf("total = %d, want 1", totalOf(page))
	}
	if page.Items[0].Origin != "scan" || page.Items[0].ScanFolder == nil {
		t.Errorf("row = %+v, want a scanned row with a scan folder", page.Items[0])
	}

	result, err = lib.SyncScanned(context.Background(), groups)
	if err != nil {
		t.Fatal(err)
	}
	if result.Updated != 1 || result.Added != 0 {
		t.Fatalf("second sync = %+v, want one update", result)
	}

	empty := []types.ScannedFolderGroup{{FolderPath: root, FolderName: "Library"}}
	result, err = lib.SyncScanned(context.Background(), empty)
	if err != nil {
		t.Fatal(err)
	}
	if result.Missing != 1 || result.Removed != 0 {
		t.Fatalf("first sweep = %+v, want one missing", result)
	}
	result, err = lib.SyncScanned(context.Background(), empty)
	if err != nil {
		t.Fatal(err)
	}
	if result.Removed != 1 {
		t.Fatalf("second sweep = %+v, want one removed", result)
	}
	page, err = lib.List(context.Background(), Filter{Source: "manual"})
	if err != nil {
		t.Fatal(err)
	}
	if totalOf(page) != 0 {
		t.Fatalf("swept game remains: %d rows", totalOf(page))
	}
}

func TestSweepScanFolderKeepsPinnedRows(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	root := filepath.Join(t.TempDir(), "Library")
	gameRoot := filepath.Join(root, "Game")
	exe := filepath.Join(gameRoot, "Game.exe")
	config := localConfig("one", "Game", exe)
	_, err := lib.BackfillLocal(context.Background(), LocalBackfillInput{
		Configs:     []types.LaunchOptions{config},
		ScanFolders: []string{root},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Editing a scanned game pins it, so a later sweep must keep it.
	if err := lib.SaveLocalConfig(context.Background(), config); err != nil {
		t.Fatal(err)
	}
	removed, err := lib.SweepScanFolder(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if removed != 0 {
		t.Fatalf("removed pinned rows: %d", removed)
	}
}

type recordingConfigLauncher struct {
	config types.LaunchOptions
	done   chan struct{}
}

func (r *recordingConfigLauncher) LaunchConfig(_ context.Context, config types.LaunchOptions) error {
	r.config = config
	close(r.done)
	return nil
}

func TestLaunchLocalUsesSavedConfig(t *testing.T) {
	launcher := &recordingConfigLauncher{done: make(chan struct{})}
	lib := openTestLibrary(t, Options{ConfigLauncher: launcher})
	root := filepath.Join(t.TempDir(), "Harbor")
	config := localConfig("one", "Harbor", filepath.Join(root, "Harbor.exe"))
	config.PrefixPath = "/prefixes/harbor"
	if _, err := lib.BackfillLocal(context.Background(), LocalBackfillInput{Configs: []types.LaunchOptions{config}}); err != nil {
		t.Fatal(err)
	}
	page, err := lib.List(context.Background(), Filter{Source: "manual"})
	if err != nil {
		t.Fatal(err)
	}
	if totalOf(page) != 1 {
		t.Fatalf("total = %d, want 1", totalOf(page))
	}
	result, err := lib.Launch(context.Background(), page.Items[0].ID)
	if err != nil {
		t.Fatalf("Launch: %v", err)
	}
	if !result.Started {
		t.Error("Launch did not start")
	}
	<-launcher.done
	if launcher.config.PrefixPath != "/prefixes/harbor" {
		t.Errorf("prefix = %q, want the saved prefix", launcher.config.PrefixPath)
	}
	game, err := lib.Get(context.Background(), page.Items[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if game.LastPlayedAt == nil {
		t.Error("last played was not recorded")
	}
}
