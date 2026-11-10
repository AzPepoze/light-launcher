// Package plugins runs out-of-process source plugins: install, verify, supervise and serve them.
package plugins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
)

// SupportedAPIVersion is the protocol major version this host speaks.
const SupportedAPIVersion = "1"

// Manifest is the parsed plugin.json file at the root of a .llplugin package.
type Manifest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	APIVersion  string            `json:"apiVersion"`
	MinHost     string            `json:"minHost"`
	Entry       map[string]string `json:"entry"`
	Contributes Contributes       `json:"contributes"`
	Permissions Permissions       `json:"permissions"`
}

// Contributes lists what the plugin adds to the launcher.
type Contributes struct {
	Sources        []SourceContribution  `json:"sources,omitempty"`
	Settings       *SettingsRef          `json:"settings,omitempty"`
	Pages          []PageContribution    `json:"pages,omitempty"`
	Commands       []CommandContribution `json:"commands,omitempty"`
	Hooks          []string              `json:"hooks,omitempty"`
	LaunchAdapters []NamedItem           `json:"launchAdapters,omitempty"`
	Importers      []NamedItem           `json:"importers,omitempty"`
	Themes         []ThemeContribution   `json:"themes,omitempty"`
}

// SourceContribution is one game source a plugin offers.
type SourceContribution struct {
	ID           string   `json:"id"`
	Label        string   `json:"label"`
	Capabilities []string `json:"capabilities"`
}

// SettingsRef points at the JSON Schema file that builds the settings form.
type SettingsRef struct {
	Ref string `json:"$ref"`
}

// PageContribution is a UI page the plugin serves.
type PageContribution struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

// NamedItem is an id and a label, used by launch adapters and importers.
type NamedItem struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// CommandContribution is an action in the command palette. Method is the plugin method it calls.
type CommandContribution struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Method string `json:"method,omitempty"`
}

// ThemeContribution is a theme file with colors and fonts.
type ThemeContribution struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Path  string `json:"path"`
}

// Permissions lists what the plugin may use through host services.
type Permissions struct {
	Network []string `json:"network"`
	FS      []string `json:"fs"`
	Secrets bool     `json:"secrets"`
	Browser bool     `json:"browser"`
	Exec    bool     `json:"exec"`
}

// ParseManifest decodes and validates plugin.json bytes. Unknown keys are errors.
func ParseManifest(data []byte) (Manifest, error) {
	var manifest Manifest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("plugin.json: %w", err)
	}
	if err := manifest.Validate(); err != nil {
		return Manifest{}, fmt.Errorf("plugin.json: %w", err)
	}
	return manifest, nil
}

// EntryFor returns the executable path for a platform key such as linux-x64.
func (m Manifest) EntryFor(platformKey string) (string, bool) {
	entry, ok := m.Entry[platformKey]
	return entry, ok
}

// PlatformKey returns the manifest entry key for the running OS and CPU.
func PlatformKey() string {
	return platformKeyFor(runtime.GOOS, runtime.GOARCH)
}

func platformKeyFor(goos, goarch string) string {
	return platformOS(goos) + "-" + platformArch(goarch)
}

func platformOS(goos string) string {
	switch goos {
	case "darwin":
		return "mac"
	default:
		return goos
	}
}

func platformArch(goarch string) string {
	switch goarch {
	case "amd64":
		return "x64"
	default:
		return goarch
	}
}
