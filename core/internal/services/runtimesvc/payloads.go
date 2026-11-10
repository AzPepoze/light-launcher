package runtimesvc

import "light-launcher/core/internal/types"

// Payload structs. Field names match the object keys the UI sends (see the old IPC router).

type runGameInput struct {
	Options  types.LaunchOptions `json:"options"`
	ShowLogs bool                `json:"showLogs"`
}

type nameInput struct {
	Name string `json:"name"`
}

type prefixNameInput struct {
	PrefixName string `json:"prefixName"`
}

type savePrefixInput struct {
	PrefixName string              `json:"prefixName"`
	Options    types.LaunchOptions `json:"options"`
}

type runPrefixToolInput struct {
	PrefixPath string `json:"prefixPath"`
	ToolName   string `json:"toolName"`
	ProtonPath string `json:"protonPath"`
}

type variantInput struct {
	VariantID string `json:"variantID"`
}

type installProtonInput struct {
	URL     string `json:"url"`
	Version string `json:"version"`
}

type lsfgProfileInput struct {
	Name     string `json:"name"`
	GamePath string `json:"gamePath"`
}

type disableLsfgInput struct {
	ProfileName string `json:"profileName"`
	GamePath    string `json:"gamePath"`
}

type removeProfileInput struct {
	MainExecutablePath string `json:"mainExecutablePath"`
}

type pidInput struct {
	Pid int `json:"pid"`
}
