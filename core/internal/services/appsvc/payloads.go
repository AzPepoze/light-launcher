package appsvc

import "light-launcher/core/internal/types"

// Payload shapes for the UI calls. Field names match the arguments that src/main/ipc/router.ts reads.

type pathPayload struct {
	Path string `json:"path"`
}

type imagePayload struct {
	ImagePath string `json:"imagePath"`
}

type executablePayload struct {
	ExecutablePath string `json:"executablePath"`
}

type urlPayload struct {
	URL string `json:"url"`
}

type settingsPayload struct {
	Settings types.AppSettings `json:"settings"`
}

type optionsPayload struct {
	Options types.LaunchOptions `json:"options"`
}

type folderPayload struct {
	FolderPath string `json:"folderPath"`
}

// searchPayload leaves MaxDepth nil when the UI omits it, which means unlimited depth.
type searchPayload struct {
	FolderPath   string   `json:"folderPath"`
	MaxDepth     *int     `json:"maxDepth"`
	ExcludeNames []string `json:"excludeNames"`
}

type forcePayload struct {
	Force bool `json:"force"`
}

type folderConfigPayload struct {
	FolderPath   string   `json:"folderPath"`
	Depth        int      `json:"depth"`
	ExcludeNames []string `json:"excludeNames"`
}

type renamePayload struct {
	OldPath      string   `json:"oldPath"`
	NewPath      string   `json:"newPath"`
	Depth        int      `json:"depth"`
	ExcludeNames []string `json:"excludeNames"`
}

type customFilePayload struct {
	Title   string         `json:"title"`
	Filters []customFilter `json:"filters"`
}
