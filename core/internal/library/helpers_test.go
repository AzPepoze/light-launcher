package library

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"light-launcher/core/internal/contract"
)

// fakeSources answers plugin calls from a map of method name to reply function.
// The reply is marshalled to JSON and decoded into the caller's out, as the real host does.
type fakeSources struct {
	infos   []contract.SourceInfo
	replies map[string]func(params any) (any, error)
}

func (f *fakeSources) Sources() []contract.SourceInfo {
	return f.infos
}

func (f *fakeSources) Call(_ context.Context, _, method string, params, out any) error {
	reply, ok := f.replies[method]
	if !ok {
		return fmt.Errorf("no fake for %s", method)
	}
	value, err := reply(params)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, out)
}

func (f *fakeSources) StartDownload(context.Context, string, any, int64) (int64, error) {
	return 0, fmt.Errorf("no fake for StartDownload")
}

// openTestLibrary opens a library in a temp folder that is closed when the test ends.
func openTestLibrary(t *testing.T, opts Options) *Library {
	t.Helper()
	dir := t.TempDir()
	if opts.DBPath == "" {
		opts.DBPath = filepath.Join(dir, "library.db")
	}
	if opts.MediaDir == "" {
		opts.MediaDir = filepath.Join(dir, "media")
	}
	lib, err := Open(context.Background(), opts)
	if err != nil {
		t.Fatalf("open library: %v", err)
	}
	t.Cleanup(func() { lib.Close() })
	return lib
}

// upsertGame stores a manual game with the given title and returns the saved row.
func upsertGame(t *testing.T, lib *Library, sourceID, title string) Game {
	t.Helper()
	id, err := lib.Upsert(context.Background(), contract.GameDTO{
		Source:   "manual",
		SourceID: sourceID,
		URL:      "file://" + sourceID,
		Title:    title,
	})
	if err != nil {
		t.Fatalf("upsert %s: %v", title, err)
	}
	game, err := lib.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("get %d: %v", id, err)
	}
	return game
}

func strPtr(value string) *string {
	return &value
}
