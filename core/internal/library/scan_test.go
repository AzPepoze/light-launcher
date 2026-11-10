package library

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPrettyTitle(t *testing.T) {
	cases := map[string]string{
		"Mad_Island_0.34":     "Mad Island 0.34",
		"koh_windows_v02":     "koh v02",
		"Some Game (Patreon)": "Some Game",
		"Cool Game-x64":       "Cool Game",
		"Plain":               "Plain",
	}
	for in, want := range cases {
		if got := prettyTitle(in); got != want {
			t.Errorf("prettyTitle(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDetectExecutableSkipsInstallersAndPrefersShallowNative(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip("launcher suffixes are platform specific; this test covers Linux")
	}
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "uninstall.exe"))
	writeFile(t, filepath.Join(dir, "redist", "vc_redist.x64.exe"))
	writeFile(t, filepath.Join(dir, "bin", "Game.exe"))
	writeFile(t, filepath.Join(dir, "run.sh"))

	if got := DetectExecutable(dir); got != filepath.Join(dir, "run.sh") {
		t.Errorf("DetectExecutable = %q, want run.sh", got)
	}
}

func TestDetectExecutableFindsNothingInEmptyFolder(t *testing.T) {
	if got := DetectExecutable(t.TempDir()); got != "" {
		t.Errorf("DetectExecutable(empty) = %q, want empty", got)
	}
}

func TestScanFolderImportsUnknownFoldersOnce(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip("launcher suffixes are platform specific; this test covers Linux")
	}
	lib := openTestLibrary(t, Options{})
	root := t.TempDir()
	game := filepath.Join(root, "Cool_Game_v1.2")
	writeFile(t, filepath.Join(game, "CoolGame.sh"))
	writeFile(t, filepath.Join(root, "downloads", "junk.sh"))

	report, err := lib.ScanFolder(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if report.Added != 1 {
		t.Errorf("first scan added %d, want 1", report.Added)
	}

	page, _ := lib.List(context.Background(), Filter{})
	if page.Total != 1 {
		t.Fatalf("library total = %d, want 1", page.Total)
	}
	imported := page.Items[0]
	if imported.Title != "Cool Game v1.2" || derefString(imported.Version) != "1.2" {
		t.Errorf("title=%q version=%v", imported.Title, imported.Version)
	}
	if derefString(imported.InstallDir) != game || derefString(imported.ExePath) != filepath.Join(game, "CoolGame.sh") {
		t.Errorf("install dir=%v exe=%v", imported.InstallDir, imported.ExePath)
	}

	again, _ := lib.ScanFolder(context.Background(), root)
	if again.Added != 0 {
		t.Errorf("second scan added %d, want 0", again.Added)
	}
}

func TestScanFolderReadsManifest(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	root := t.TempDir()
	dir := filepath.Join(root, "manifest-game")
	manifest := `{"source":"f95","sourceId":"77","title":"Manifest Game","gameVersion":"2.0"}`
	writeFile(t, filepath.Join(dir, ".interesting-manager", "manifest.json"))
	if err := os.WriteFile(filepath.Join(dir, ".interesting-manager", "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := lib.ScanFolder(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if report.Added != 1 {
		t.Fatalf("added = %d, want 1", report.Added)
	}
	game, err := lib.BySource(context.Background(), "f95", "77")
	if err != nil {
		t.Fatal(err)
	}
	if game.Title != "Manifest Game" || derefString(game.InstalledVersion) != "2.0" || derefString(game.InstallDir) != dir {
		t.Errorf("game = %+v", game)
	}
}

func TestScanFolderMissingRootIsAnError(t *testing.T) {
	lib := openTestLibrary(t, Options{})
	if _, err := lib.ScanFolder(context.Background(), filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Error("missing folder should be an error")
	}
}
