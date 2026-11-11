package library

import (
	"context"
	"testing"

	"light-launcher/core/internal/contract"
)

func TestPreviewReturnsFilesAndSourceLabel(t *testing.T) {
	sources := strictF95Sources()
	sources.infos[0].Label = "F95 Zone"
	lib := openTestLibrary(t, Options{Sources: sources})

	preview, err := lib.Preview(context.Background(), "https://f95/42")
	if err != nil {
		t.Fatal(err)
	}
	if preview.SourceLabel != "F95 Zone" {
		t.Errorf("sourceLabel = %q", preview.SourceLabel)
	}
	if len(preview.Files) != 1 || preview.Files[0].Label != "Windows" {
		t.Errorf("files = %+v, want the resolved Windows file", preview.Files)
	}
}

func TestPreviewWithoutFileCapabilityHasEmptyFiles(t *testing.T) {
	sources := &strictSources{infos: []contract.SourceInfo{{Plugin: "f95plugin", SourceID: "f95", Label: "F95"}}}
	lib := openTestLibrary(t, Options{Sources: sources})

	preview, err := lib.Preview(context.Background(), "https://f95/42")
	if err != nil {
		t.Fatal(err)
	}
	if preview.Files == nil || len(preview.Files) != 0 {
		t.Errorf("files = %#v, want an empty list", preview.Files)
	}
}
