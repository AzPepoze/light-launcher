package librarysvc

import (
	"encoding/json"
	"testing"

	"light-launcher/core/internal/library"
)

func TestLibraryUpdateSetsInstallDir(t *testing.T) {
	router, _ := setup(t)
	id := createManual(t, router)
	folder := t.TempDir()
	payload, err := json.Marshal(map[string]any{"id": id, "installDir": folder})
	if err != nil {
		t.Fatal(err)
	}

	game := call(t, router, "LibraryUpdate", string(payload)).(library.Game)
	if game.InstallDir == nil || *game.InstallDir != folder {
		t.Errorf("installDir = %v, want %s", game.InstallDir, folder)
	}
}

func TestLibraryRemoveReportsDeletedFiles(t *testing.T) {
	router, _ := setup(t)
	id := createManual(t, router)

	result := call(t, router, "LibraryRemove", idJSON(id, `"deleteFiles":true`)).(removeResult)
	if !result.OK || result.DeletedFiles {
		t.Errorf("result = %+v, want ok with no files deleted", result)
	}
}
