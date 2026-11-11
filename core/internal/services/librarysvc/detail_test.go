package librarysvc

import (
	"encoding/json"
	"testing"
)

func TestLibraryGetJSONKeepsGameFieldsAndAddsFilesAndChanges(t *testing.T) {
	router, _ := setup(t)
	id := createManual(t, router)

	raw, err := json.Marshal(call(t, router, "LibraryGet", idJSON(id, "")))
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"id", "title", "running", "files", "changes"} {
		if _, ok := wire[key]; !ok {
			t.Errorf("LibraryGet JSON is missing %q", key)
		}
	}
	if _, ok := wire["game"]; ok {
		t.Error("the game fields must stay at the top level, not under game")
	}
}
