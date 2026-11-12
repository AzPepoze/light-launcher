package config

import (
	"os"
	"path/filepath"
	"testing"

	"light-launcher/core/internal/types"
)

func TestGameStoreSaveGeneratesIDAndVersion(t *testing.T) {
	dir := t.TempDir()
	games := NewGameStore(dir)
	options := types.LaunchOptions{Name: "Game", GamePath: "/games/Game.exe"}
	if err := games.Save(&options); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if len(options.ID) != 16 {
		t.Fatalf("generated ID = %q, want 16 hex characters", options.ID)
	}

	target := filepath.Join(dir, options.ID, "config.json")
	var raw map[string]any
	if err := readJSON(t, target, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["version"] != float64(DataVersion) {
		t.Errorf("version = %v, want %d", raw["version"], DataVersion)
	}
}

func TestGameStoreListAndFind(t *testing.T) {
	games := NewGameStore(t.TempDir())
	first := types.LaunchOptions{ID: "a", Name: "One", GamePath: "/games/One.exe"}
	second := types.LaunchOptions{ID: "b", Name: "Two", LauncherPath: "/games/Two.exe"}
	for _, options := range []*types.LaunchOptions{&first, &second} {
		if err := games.Save(options); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	if got := len(games.List()); got != 2 {
		t.Fatalf("List length = %d, want 2", got)
	}
	if found := games.Find("/GAMES/one.EXE"); found == nil || found.ID != "a" {
		t.Errorf("Find by game path = %+v", found)
	}
	if found := games.Find("/games/Two.exe"); found == nil || found.ID != "b" {
		t.Errorf("Find by launcher path = %+v", found)
	}
	if found := games.Find("/games/Missing.exe"); found != nil {
		t.Errorf("Find of a missing game = %+v, want nil", found)
	}
}

func TestGameStoreSkipsBrokenAndLoadsLegacyConfigs(t *testing.T) {
	dir := t.TempDir()
	writeRaw(t, filepath.Join(dir, "broken", "config.json"), "{")
	writeRaw(t, filepath.Join(dir, "legacy", "config.json"), `{"ID": "legacy", "Name": "Old"}`)
	valid := types.LaunchOptions{ID: "ok", Name: "Good", GamePath: "/games/Good.exe"}
	if err := NewGameStore(dir).Save(&valid); err != nil {
		t.Fatalf("Save: %v", err)
	}

	list := NewGameStore(dir).List()
	if len(list) != 2 {
		t.Fatalf("List = %+v, want the valid and the legacy config", list)
	}
}

func TestGameStoreMissingDirectoryListsNothing(t *testing.T) {
	if got := NewGameStore(filepath.Join(t.TempDir(), "missing")).List(); len(got) != 0 {
		t.Fatalf("List = %+v, want empty", got)
	}
}

func TestGameStoreConfigPathPrefersID(t *testing.T) {
	games := NewGameStore(filepath.Join("root", "config"))
	if got, want := games.ConfigPath("Name", "id"), filepath.Join("root", "config", "id", "config.json"); got != want {
		t.Errorf("ConfigPath = %q, want %q", got, want)
	}
	if got, want := games.ConfigPath("Name", ""), filepath.Join("root", "config", "Name", "config.json"); got != want {
		t.Errorf("ConfigPath without id = %q, want %q", got, want)
	}
}

func TestGameStorePersistedFileIsReadable(t *testing.T) {
	dir := t.TempDir()
	games := NewGameStore(dir)
	options := types.LaunchOptions{ID: "x", Name: "Game", GamePath: "/games/Game.exe"}
	if err := games.Save(&options); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(games.ConfigPath("Game", "x"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("config file is empty")
	}
}
