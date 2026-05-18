package app

import (
	"testing"

	"light-launcher/internal/types"
)

func TestFindProtonMatch(t *testing.T) {
	app := &App{} // Create a minimal App instance for testing

	protonTools := []types.ProtonTool{
		{
			DisplayName: "Proton - Experimental",
			Name:        "Proton-Experimental",
			Path:        "/home/user/.local/share/Steam/steamapps/common/Proton-Experimental",
			IsSteam:     true,
		},
		{
			DisplayName: "Proton Hotfix",
			Name:        "Proton-Hotfix",
			Path:        "/home/user/custom/Proton-Hotfix",
			IsSteam:     false,
		},
	}

	tests := []struct {
		name        string
		savedPath   string
		expected    *types.ProtonTool
		description string
	}{
		{
			name:        "exact path match",
			savedPath:   "/home/user/.local/share/Steam/steamapps/common/Proton-Experimental",
			expected:    &protonTools[0],
			description: "Should match exact path",
		},
		{
			name:        "directory name and steam status match",
			savedPath:   "/different/path/Proton-Experimental",
			expected:    &protonTools[0],
			description: "Should match by directory name + Steam status (resilient to symlinks)",
		},
		{
			name:        "directory name only match",
			savedPath:   "/any/path/Proton-Hotfix",
			expected:    &protonTools[1],
			description: "Should match by directory name only (fallback)",
		},
		{
			name:        "no match returns nil",
			savedPath:   "/path/to/NonExistent",
			expected:    nil,
			description: "Should return nil when no match found",
		},
		{
			name:        "empty path returns nil",
			savedPath:   "",
			expected:    nil,
			description: "Should return nil for empty path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.findProtonMatch(tt.savedPath, protonTools)
			if result != tt.expected {
				if tt.expected == nil && result != nil {
					t.Errorf("%s: expected nil, got %v", tt.description, result.DisplayName)
				} else if tt.expected != nil && result == nil {
					t.Errorf("%s: expected %v, got nil", tt.description, tt.expected.DisplayName)
				} else if tt.expected != nil && result != nil && result.DisplayName != tt.expected.DisplayName {
					t.Errorf("%s: expected %v, got %v", tt.description, tt.expected.DisplayName, result.DisplayName)
				}
			}
		})
	}
}

func TestFindProtonMatchPriority(t *testing.T) {
	app := &App{} // Create a minimal App instance for testing

	// Test that exact match takes priority over name match
	protonTools := []types.ProtonTool{
		{
			DisplayName: "Proton-Experimental (exact)",
			Name:        "Proton-Experimental",
			Path:        "/exact/path/Proton-Experimental",
			IsSteam:     true,
		},
		{
			DisplayName: "Proton-Experimental (fallback)",
			Name:        "Proton-Experimental",
			Path:        "/fallback/path/Proton-Experimental",
			IsSteam:     false,
		},
	}

	// Exact path should match first one
	result := app.findProtonMatch("/exact/path/Proton-Experimental", protonTools)
	if result == nil || result.DisplayName != "Proton-Experimental (exact)" {
		t.Errorf("Exact match priority failed: expected 'Proton-Experimental (exact)', got %v", result)
	}

	// Name + Steam match should prefer Steam version when path contains steamapps/common
	result = app.findProtonMatch("/home/user/.local/share/Steam/steamapps/common/Proton-Experimental", protonTools)
	if result == nil || result.DisplayName != "Proton-Experimental (exact)" {
		t.Errorf("Name+Steam priority failed: expected 'Proton-Experimental (exact)', got %v", result)
	}

	// Name-only match should match non-Steam version when path doesn't contain steamapps
	result = app.findProtonMatch("/different/Proton-Experimental", protonTools)
	if result == nil || result.DisplayName != "Proton-Experimental (fallback)" {
		t.Errorf("Name-only fallback priority failed: expected 'Proton-Experimental (fallback)', got %v", result)
	}
}
