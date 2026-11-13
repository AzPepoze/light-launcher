package config

import (
	"encoding/json"
	"strings"
	"testing"

	"light-launcher/core/internal/types"
)

func TestDefaultSettingsHaveLibraryCardSize(t *testing.T) {
	settings := DefaultSettings()
	if settings.LibraryCardSize == nil || *settings.LibraryCardSize != libraryCardSizeDefault {
		t.Fatalf("card size = %v, want %d", settings.LibraryCardSize, libraryCardSizeDefault)
	}
}

func TestNormalizeClampsLibraryCardSize(t *testing.T) {
	tests := []struct {
		name string
		in   *int
		want int
	}{
		{"missing uses default", nil, libraryCardSizeDefault},
		{"below range clamps to min", ptrInt(10), libraryCardSizeMin},
		{"above range clamps to max", ptrInt(999), libraryCardSizeMax},
		{"in range is kept", ptrInt(200), 200},
		{"min is kept", ptrInt(libraryCardSizeMin), libraryCardSizeMin},
		{"max is kept", ptrInt(libraryCardSizeMax), libraryCardSizeMax},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := &types.AppSettings{LibraryCardSize: tt.in}
			normalizeLibraryCardSize(settings)
			if settings.LibraryCardSize == nil || *settings.LibraryCardSize != tt.want {
				t.Fatalf("card size = %v, want %d", settings.LibraryCardSize, tt.want)
			}
		})
	}
}

func TestNormalizeLibraryCardSizeReportsChanges(t *testing.T) {
	inRange := &types.AppSettings{LibraryCardSize: ptrInt(180)}
	if normalizeLibraryCardSize(inRange) {
		t.Error("an in-range card size reported a change")
	}
	outOfRange := &types.AppSettings{LibraryCardSize: ptrInt(500)}
	if !normalizeLibraryCardSize(outOfRange) {
		t.Error("a clamped card size did not report a change")
	}
}

func TestLibraryCardSizeUsesTheWireKey(t *testing.T) {
	size := 200
	blur := true
	data, err := json.Marshal(types.AppSettings{LibraryCardSize: &size, BlurCoversUntilHover: &blur})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"LibraryCardSize":200`, `"BlurCoversUntilHover":true`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("settings JSON = %s, missing %s", data, want)
		}
	}
}

func TestBlurCoversUntilHoverIsOffWhenAbsent(t *testing.T) {
	settings := DefaultSettings()
	if settings.BlurCoversUntilHover != nil {
		t.Errorf("BlurCoversUntilHover = %v, want nil (off)", *settings.BlurCoversUntilHover)
	}
}

func ptrInt(value int) *int {
	return &value
}
