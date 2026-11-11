package appsvc

import (
	"os"
	"path/filepath"
	"testing"

	"light-launcher/core/internal/config"
)

func TestCleanFolderName(t *testing.T) {
	tests := []struct {
		folder string
		want   string
	}{
		{"Mad_Island_0.34", "Mad Island"},
		{"Hollow_Knight", "Hollow Knight"},
		{"Mr.Robot", "Mr Robot"},
		{"Game-v1.2.3", "Game"},
		{"Game [v0.5]", "Game"},
		{"Game_Win", "Game"},
		{"Game-pc", "Game"},
		{"Game-win64", "Game"},
		{"Dark-Souls-v1.0-pc", "Dark-Souls"},
		{"My Game_v2", "My Game"},
		{"Game With Spaces_0.5", "Game With Spaces"},
		{"Game2", "Game2"},
		{"Plain", "Plain"},
		{"-pc", "-pc"},
		{"", ""},
	}
	for _, test := range tests {
		if got := cleanFolderName(test.folder); got != test.want {
			t.Errorf("cleanFolderName(%q) = %q, want %q", test.folder, got, test.want)
		}
	}
}

func TestIsGenericName(t *testing.T) {
	tests := map[string]bool{
		"":              true,
		"Unity":         true,
		"Unreal Engine": true,
		"Ren'Py":        true,
		"Godot Engine":  true,
		"NW.js":         true,
		"Electron":      true,
		"Launcher":      true,
		"UnityPlayer":   true,
		"Microsoft(R) Windows(R) Operating System": true,
		"Microsoft Corporation":                    true,
		"Hollow Knight":                            false,
		"Microsoft Flight Simulator":               false,
	}
	for name, want := range tests {
		if got := isGenericName(name); got != want {
			t.Errorf("isGenericName(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestIsGenericExeName(t *testing.T) {
	tests := []struct {
		exe    string
		folder string
		want   bool
	}{
		{"game", "", true},
		{"Launcher", "Hollow Knight", true},
		{"MyGame-32", "MyGame", true},
		{"mygame-64", "MyGame", true},
		{"MyGame-Win64-Shipping", "MyGame", true},
		{"UnityPlayer", "Anything", true},
		{"MyGame", "MyGame", false},
		{"MyGame-32", "Other", false},
		{"Hollow", "Hollow Folder", false},
	}
	for _, test := range tests {
		if got := isGenericExeName(test.exe, test.folder); got != test.want {
			t.Errorf("isGenericExeName(%q, %q) = %v, want %v", test.exe, test.folder, got, test.want)
		}
	}
}

func TestResolveGameName(t *testing.T) {
	tests := []struct {
		name string
		src  gameNameSources
		want string
	}{
		{
			name: "user name beats the product name",
			src:  gameNameSources{userName: "My Name", exeName: "Hollow", product: "Hollow Knight"},
			want: "My Name",
		},
		{
			name: "product name beats the launcher name",
			src:  gameNameSources{exeName: "Hollow", folderName: "Games", product: "Hollow Knight"},
			want: "Hollow Knight",
		},
		{
			name: "file description is the fallback for the product name",
			src:  gameNameSources{exeName: "Hollow", product: "Unity", description: "Hollow Knight"},
			want: "Hollow Knight",
		},
		{
			name: "generic product name is skipped",
			src:  gameNameSources{exeName: "Hollow", folderName: "Folder", product: "Electron"},
			want: "Hollow",
		},
		{
			name: "product name equal to the launcher name is skipped",
			src:  gameNameSources{exeName: "Hollow", folderName: "Folder", product: "launcher"},
			want: "Hollow",
		},
		{
			name: "generic launcher name uses the cleaned folder name",
			src:  gameNameSources{exeName: "game", folderName: "Mad_Island_0.34"},
			want: "Mad Island",
		},
		{
			name: "several launchers use the cleaned folder name",
			src:  gameNameSources{exeName: "Editor", folderName: "My_Game", multiLauncher: true},
			want: "My Game",
		},
		{
			name: "a loose generic launcher keeps its name",
			src:  gameNameSources{exeName: "game"},
			want: "game",
		},
		{
			name: "a non-generic launcher name is kept without a product name",
			src:  gameNameSources{exeName: "Hollow", folderName: "Folder"},
			want: "Hollow",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := resolveGameName(test.src); got != test.want {
				t.Fatalf("resolveGameName = %q, want %q", got, test.want)
			}
		})
	}
}

// writeTestExe writes a synthetic PE file with the given version strings.
func writeTestExe(t *testing.T, path, product, description string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buildTestPE(product, description), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestGetAutoScannedGamesNamesCards(t *testing.T) {
	svc := newTestService(t)
	root := config.CleanPath(t.TempDir())
	writeTestExe(t, filepath.Join(root, "Hollow_Knight_v1.2", "Launcher.exe"), "Hollow Knight", "")
	writeTestExe(t, filepath.Join(root, "Other", "Hollow.exe"), "Electron", "")
	mustWrite(t, filepath.Join(root, "Plain_Game", "game.exe"))
	if err := svc.addScanFolder(root); err != nil {
		t.Fatalf("addScanFolder: %v", err)
	}

	names := scannedNamesByFolder(t, svc, root)
	want := map[string]string{
		"Hollow_Knight_v1.2": "Hollow Knight",
		"Other":              "Hollow",
		"Plain_Game":         "Plain Game",
	}
	for folder, wantName := range want {
		if names[folder] != wantName {
			t.Errorf("card for %s = %q, want %q", folder, names[folder], wantName)
		}
	}
}

func TestGetAutoScannedGamesKeepsUserRename(t *testing.T) {
	svc := newTestService(t)
	root := config.CleanPath(t.TempDir())
	exe := filepath.Join(root, "Hollow_Knight", "Launcher.exe")
	writeTestExe(t, exe, "Hollow Knight", "")
	if err := svc.addScanFolder(root); err != nil {
		t.Fatalf("addScanFolder: %v", err)
	}
	if got := scannedNamesByFolder(t, svc, root)["Hollow_Knight"]; got != "Hollow Knight" {
		t.Fatalf("first card = %q, want the product name", got)
	}

	cfg := svc.games.Find(exe)
	if cfg == nil {
		t.Fatal("expected a saved config for the scanned executable")
	}
	cfg.Name = "My Name"
	if err := svc.games.Save(cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}
	if got := scannedNamesByFolder(t, svc, root)["Hollow_Knight"]; got != "My Name" {
		t.Fatalf("card after rename = %q, want %q", got, "My Name")
	}
}

// scannedNamesByFolder forces a scan and returns the card name of each game folder under root.
func scannedNamesByFolder(t *testing.T, svc *service, root string) map[string]string {
	t.Helper()
	groups, err := svc.getAutoScannedGames(true)
	if err != nil {
		t.Fatalf("getAutoScannedGames: %v", err)
	}
	names := map[string]string{}
	for _, group := range groups {
		for _, game := range group.Games {
			rel, err := filepath.Rel(root, game.Folder)
			if err != nil {
				t.Fatal(err)
			}
			names[filepath.ToSlash(rel)] = game.Name
		}
	}
	return names
}
