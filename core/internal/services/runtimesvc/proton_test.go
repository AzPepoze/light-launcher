package runtimesvc

import (
	"path/filepath"
	"testing"

	"light-launcher/core/internal/types"
)

func TestProtonVariantsIncludeKnownBuilds(t *testing.T) {
	variants := getProtonVariants()
	if len(variants) < 4 {
		t.Fatalf("got %d variants, want at least 4", len(variants))
	}
	for _, id := range []string{"ge-proton", "proton-cachyos", "kron4ek", "luxtorpeda"} {
		if _, ok := findVariant(id); !ok {
			t.Errorf("variant %q missing", id)
		}
	}
}

func TestFindProtonMatch(t *testing.T) {
	tools := []types.ProtonTool{
		{
			Name:        "GE-Proton9-25",
			Path:        filepath.FromSlash("/home/user/.steam/root/compatibilitytools.d/GE-Proton9-25"),
			IsSteam:     false,
			DisplayName: "Custom: GE-Proton9-25",
		},
		{
			Name:        "Proton 9.0",
			Path:        filepath.FromSlash("/home/user/.steam/root/steamapps/common/Proton 9.0"),
			IsSteam:     true,
			DisplayName: "Steam: Proton 9.0",
		},
	}

	byDisplayName := findProtonMatch("Custom: GE-Proton9-25", tools)
	if byDisplayName == nil || byDisplayName.Name != "GE-Proton9-25" {
		t.Errorf("display name match = %+v, want GE-Proton9-25", byDisplayName)
	}

	byPath := findProtonMatch("/home/user/.steam/root/steamapps/common/Proton 9.0", tools)
	if byPath == nil || !byPath.IsSteam {
		t.Errorf("path match = %+v, want the Steam tool", byPath)
	}

	if missing := findProtonMatch("NonExistent", tools); missing != nil {
		t.Errorf("unknown name matched %+v, want nil", missing)
	}
}

func TestFindProtonMatchPrefersSameOrigin(t *testing.T) {
	tools := []types.ProtonTool{
		{Name: "Proton-X", Path: "/custom/Proton-X", IsSteam: false, DisplayName: "Custom: Proton-X"},
		{Name: "Proton-X", Path: "/steam/Proton-X", IsSteam: true, DisplayName: "Steam: Proton-X"},
	}

	match := findProtonMatch("/elsewhere/steamapps/common/Proton-X", tools)
	if match == nil || !match.IsSteam {
		t.Errorf("match = %+v, want the Steam tool for a steamapps path", match)
	}
}
