package appsvc

import (
	"slices"
	"strings"

	"light-launcher/core/internal/types"
)

// normalizeSettings fills missing lists, adds legacy scan folders and refreshes old default excludes.
// It reports whether the file needs saving.
func normalizeSettings(settings *types.AppSettings) bool {
	if settings.ScanFolderConfigs == nil {
		settings.ScanFolderConfigs = []types.ScanFolderConfig{}
	}
	if settings.ScanFolders == nil {
		settings.ScanFolders = []string{}
	}
	if settings.Blacklist == nil {
		settings.Blacklist = []string{}
	}

	modified := addLegacyScanFolders(settings)
	if backfillExcludeNames(settings) {
		modified = true
	}
	return modified
}

// addLegacyScanFolders turns plain ScanFolders entries into ScanFolderConfigs with default settings.
func addLegacyScanFolders(settings *types.AppSettings) bool {
	added := false
	for _, folder := range settings.ScanFolders {
		cleaned := normalizePath(folder)
		if hasNormalizedScanFolder(settings.ScanFolderConfigs, cleaned) {
			continue
		}
		settings.ScanFolderConfigs = append(settings.ScanFolderConfigs, types.ScanFolderConfig{
			Path:         cleaned,
			Depth:        2,
			ExcludeNames: defaultScanExcludeNames(),
		})
		added = true
	}
	return added
}

// backfillExcludeNames fills empty or legacy exclude lists. Custom lists are kept.
func backfillExcludeNames(settings *types.AppSettings) bool {
	changed := false
	for i := range settings.ScanFolderConfigs {
		cfg := &settings.ScanFolderConfigs[i]
		if len(cfg.ExcludeNames) == 0 || isLegacyExcludeList(cfg.ExcludeNames) {
			cfg.ExcludeNames = defaultScanExcludeNames()
			changed = true
		}
	}
	return changed
}

func hasNormalizedScanFolder(configs []types.ScanFolderConfig, normalized string) bool {
	for _, cfg := range configs {
		if normalizePath(cfg.Path) == normalized {
			return true
		}
	}
	return false
}

func isLegacyExcludeList(excludeNames []string) bool {
	normalized := normalizedExcludeNames(excludeNames)
	for _, legacy := range legacyExcludeSets {
		if len(legacy) == len(normalized) && containsAllLowered(legacy, normalized) {
			return true
		}
	}
	return false
}

// normalizedExcludeNames trims and lowercases names and drops blanks.
func normalizedExcludeNames(excludeNames []string) []string {
	normalized := []string{}
	for _, name := range excludeNames {
		if trimmed := strings.ToLower(strings.TrimSpace(name)); trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	return normalized
}

func containsAllLowered(names []string, list []string) bool {
	for _, name := range names {
		if !slices.Contains(list, strings.ToLower(name)) {
			return false
		}
	}
	return true
}
