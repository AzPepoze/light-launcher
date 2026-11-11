package runtimesvc

import (
	"light-launcher/core/internal/types"
)

func (s *service) savePrefixConfig(prefixName string, options types.LaunchOptions) error {
	return s.prefixes.Save(prefixName, options)
}

// loadPrefixConfig returns the saved prefix config, or defaults when it is missing or unreadable.
func (s *service) loadPrefixConfig(prefixName string) types.LaunchOptions {
	return s.prefixes.Load(prefixName)
}

// loadPrefixConfigWithProton adds the matching installed Proton tool to the prefix config.
func (s *service) loadPrefixConfigWithProton(prefixName string) types.PrefixConfigWithProton {
	cfg := s.loadPrefixConfig(prefixName)
	result := types.PrefixConfigWithProton{
		Config:     cfg,
		ProtonPath: cfg.ProtonPath,
	}
	if cfg.ProtonPath == "" {
		return result
	}

	match := findProtonMatch(cfg.ProtonPath, scanProtonVersions())
	if match == nil {
		result.ProtonDisplayName = cfg.ProtonPath
		return result
	}
	result.ProtonDisplayName = match.DisplayName
	result.ProtonName = match.Name
	result.ProtonPath = match.Path
	result.ProtonIsSteam = match.IsSteam
	return result
}
