package runtimesvc

import (
	"path/filepath"
	"strings"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

// ProtonVariant is a known Proton build source on GitHub.
type ProtonVariant struct {
	ID          string `json:"ID"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	RepoOwner   string `json:"RepoOwner"`
	RepoName    string `json:"RepoName"`
}

var knownVariants = []ProtonVariant{
	{
		ID:          "ge-proton",
		Name:        "GE-Proton (GloriousEggroll)",
		Description: "The most popular custom Proton build. Includes many game fixes and codec patches.",
		RepoOwner:   "GloriousEggroll",
		RepoName:    "proton-ge-custom",
	},
	{
		ID:          "proton-cachyos",
		Name:        "Proton-CachyOS",
		Description: "Optimized for performance with CachyOS patches and schedulers.",
		RepoOwner:   "CachyOS",
		RepoName:    "proton-cachyos",
	},
	{
		ID:          "kron4ek",
		Name:        "Proton-Kron4ek",
		Description: "Vanilla builds and TKG builds. Often smaller and faster updates.",
		RepoOwner:   "Kron4ek",
		RepoName:    "Proton-Builds",
	},
	{
		ID:          "luxtorpeda",
		Name:        "Luxtorpeda (Native Tools)",
		Description: "Runs Windows games using native Linux engines (e.g. GZDoom, ScummVM).",
		RepoOwner:   "luxtorpeda-dev",
		RepoName:    "luxtorpeda",
	},
}

func getProtonVariants() []ProtonVariant {
	return knownVariants
}

func findVariant(id string) (ProtonVariant, bool) {
	for _, variant := range knownVariants {
		if variant.ID == id {
			return variant, true
		}
	}
	return ProtonVariant{}, false
}

// findProtonMatch finds the installed tool that a saved path or display name refers to.
// It tries the display name, the full path, then the folder name (preferring the same Steam origin).
func findProtonMatch(savedPath string, tools []types.ProtonTool) *types.ProtonTool {
	if savedPath == "" {
		return nil
	}

	for i := range tools {
		if tools[i].DisplayName == savedPath {
			return &tools[i]
		}
	}
	for i := range tools {
		if tools[i].Path == savedPath {
			return &tools[i]
		}
	}

	dirName := config.BaseName(filepath.Clean(savedPath))
	isSteam := strings.Contains(savedPath, "steamapps/common")
	for i := range tools {
		if tools[i].Name == dirName && tools[i].IsSteam == isSteam {
			return &tools[i]
		}
	}
	for i := range tools {
		if tools[i].Name == dirName {
			return &tools[i]
		}
	}
	return nil
}
