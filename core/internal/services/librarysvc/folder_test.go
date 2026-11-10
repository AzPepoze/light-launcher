package librarysvc

import (
	"encoding/json"
	"testing"

	"light-launcher/core/internal/library"
)

func TestLibraryLinkFolderSetsInstallDir(t *testing.T) {
	router, events := setup(t)
	id := createManual(t, router)
	folder := t.TempDir()
	payload, err := json.Marshal(map[string]any{"id": id, "path": folder})
	if err != nil {
		t.Fatal(err)
	}

	game := call(t, router, "LibraryLinkFolder", string(payload)).(library.Game)
	if game.InstallDir == nil || *game.InstallDir != folder {
		t.Errorf("installDir = %v, want %s", game.InstallDir, folder)
	}
	if len(events.events) == 0 || events.events[len(events.events)-1] != "library.game" {
		t.Errorf("events = %v, want library.game", events.events)
	}
}
